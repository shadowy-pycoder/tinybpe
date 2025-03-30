package tinybpe

import (
	"bytes"
	"cmp"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

const (
	MinVocabSize int    = 256
	MaxVocabSize int    = int(^uint(0) >> 1)
	crlf         string = "\r\n"
)

type TokenId int

type Pair struct {
	Left  TokenId
	Right TokenId
}

func replacePairWithNewToken(pair Pair, newTokenId TokenId, ids, newIds []TokenId) []TokenId {
	for idx := 0; idx < len(ids); {
		if ids[idx] == pair.Left && idx < len(ids)-1 && ids[idx+1] == pair.Right {
			newIds = append(newIds, newTokenId)
			idx += 2
		} else {
			newIds = append(newIds, ids[idx])
			idx++
		}
	}
	return newIds
}

type Tokenizer struct {
	vocab  map[TokenId][]byte
	merges map[Pair]TokenId
}

func buildVocabFromMerges(merges map[Pair]TokenId) map[TokenId][]byte {
	vocab := make(map[TokenId][]byte, len(merges)+MinVocabSize)
	for i := range MinVocabSize {
		vocab[TokenId(i)] = []byte{byte(i)}
	}
	for _, pair := range slices.SortedStableFunc(maps.Keys(merges), func(a, b Pair) int {
		return cmp.Compare(merges[a], merges[b])
	}) {
		vocab[merges[pair]] = joinBytes(vocab[pair.Left], vocab[pair.Right])
	}
	return vocab
}

func NewTokenizer() *Tokenizer {
	tokenizer := Tokenizer{}
	tokenizer.merges = make(map[Pair]TokenId)
	tokenizer.vocab = buildVocabFromMerges(tokenizer.merges)
	return &tokenizer
}

func (t *Tokenizer) getMaxPair(ids []TokenId, counts map[Pair]int) Pair {
	var pair, maxPair Pair
	var count, maxCount int
	for i := range len(ids) - 1 {
		pair = Pair{Left: ids[i], Right: ids[i+1]}
		count = counts[pair] + 1
		counts[pair] = count
		if count > maxCount {
			maxCount = count
			maxPair = pair
		}
	}
	return maxPair
}

func (t *Tokenizer) getMinPair(ids []TokenId, counts map[Pair]int) Pair {
	var pair Pair
	var count int
	for i := range len(ids) - 1 {
		pair = Pair{Left: ids[i], Right: ids[i+1]}
		count = counts[pair] + 1
		counts[pair] = count
	}
	minCount := TokenId(MaxVocabSize)
	minPair := Pair{Left: minCount, Right: minCount}
	for _, pr := range slices.Backward(slices.SortedStableFunc(maps.Keys(counts), func(a, b Pair) int {
		return cmp.Compare(counts[a], counts[b])
	})) {
		tokenId, ok := t.merges[pr]
		if ok && tokenId < minCount {
			minCount = tokenId
			minPair = pr
		}
	}
	return minPair
}
func joinBytes(bs ...[]byte) []byte {
	n := 0
	for _, v := range bs {
		n += len(v)
	}
	b, i := make([]byte, n), 0
	for _, v := range bs {
		i += copy(b[i:], v)
	}
	return b
}

func (t *Tokenizer) Train(raw []byte, vocabSize int, verbose bool) {
	if vocabSize < MinVocabSize || vocabSize == MaxVocabSize {
		panic(fmt.Sprintf("vocabSize must be within [%d, %d) range\n", MinVocabSize, MaxVocabSize))
	}
	ids := make([]TokenId, len(raw))
	for i := range raw {
		ids[i] = TokenId(raw[i])
	}
	iterNum := vocabSize - MinVocabSize
	counts := make(map[Pair]int, len(ids)/2)
	newIds := make([]TokenId, 0, len(ids))
	for i := range iterNum {
		maxPair := t.getMaxPair(ids, counts)
		newTokenId := TokenId(MinVocabSize + i)
		t.vocab[newTokenId] = joinBytes(t.vocab[maxPair.Left], t.vocab[maxPair.Right])
		t.merges[maxPair] = newTokenId
		ids = replacePairWithNewToken(maxPair, newTokenId, ids, newIds)
		if verbose {
			fmt.Printf("Iteration %d/%d: [%d, %d] -> %d [%q] %d occurrences\n", i+1,
				iterNum,
				maxPair.Left,
				maxPair.Right,
				newTokenId,
				t.vocab[newTokenId],
				counts[maxPair])
		}
		clear(counts)
		newIds = newIds[:0]
	}
}

func (t *Tokenizer) Decode(ids []TokenId) (string, error) {
	var sb strings.Builder
	for i, idx := range ids {
		decoded, ok := t.vocab[idx]
		if !ok {
			return "", fmt.Errorf("can't decode token `%d` at position %d", idx, i)
		}
		sb.Write(decoded)
	}
	return sb.String(), nil
}

func (t *Tokenizer) Encode(raw []byte) []TokenId {
	ids := make([]TokenId, len(raw))
	for i, b := range raw {
		ids[i] = TokenId(b)
	}
	counts := make(map[Pair]int, len(ids)/2)
	newIds := make([]TokenId, 0, len(ids))
	for len(ids) >= 2 {
		pair := t.getMinPair(ids, counts)
		if pair.Left == TokenId(MaxVocabSize) || pair.Right == TokenId(MaxVocabSize) {
			break
		}
		ids = replacePairWithNewToken(pair, t.merges[pair], ids, newIds)
		clear(counts)
		newIds = newIds[:0]
	}
	return ids
}

func makeDirectoryIfNotExists(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.Mkdir(path, os.ModeDir|0755)
	}
	return nil
}

func (t *Tokenizer) Save(modelName string) error {
	if err := makeDirectoryIfNotExists("models"); err != nil {
		return err
	}
	var bb bytes.Buffer
	bb.WriteString(Version)
	bb.WriteString("\n")
	for _, pair := range slices.SortedStableFunc(maps.Keys(t.merges), func(a, b Pair) int {
		return cmp.Compare(t.merges[a], t.merges[b])
	}) {
		bb.WriteString(fmt.Sprintf("%d %d\n", pair.Left, pair.Right))
	}
	basePath := filepath.FromSlash(modelName)
	modelPath := filepath.Join("models", fmt.Sprintf("%s.model", basePath))
	if err := os.WriteFile(modelPath, bb.Bytes(), 0644); err != nil {
		return err
	}
	bb.Reset()
	invMerges := make(map[TokenId]Pair, len(t.merges))
	for pair, idx := range maps.All(t.merges) {
		invMerges[idx] = pair
	}
	for _, idx := range slices.SortedStableFunc(maps.Keys(t.vocab), func(a, b TokenId) int {
		return cmp.Compare(a, b)
	}) {
		mergedToken := t.vocab[idx]
		pair, ok := invMerges[idx]
		if ok {
			bb.WriteString(fmt.Sprintf("[%q][%q] -> [%q] %d\n", t.vocab[pair.Left], t.vocab[pair.Right], mergedToken, idx))
		} else {
			bb.WriteString(fmt.Sprintf("[%q] %d\n", mergedToken, idx))
		}
	}
	vocabPath := filepath.Join("models", fmt.Sprintf("%s.vocab", basePath))
	if err := os.WriteFile(vocabPath, bb.Bytes(), 0644); err != nil {
		return err
	}
	return nil
}

func Load(modelName string) (*Tokenizer, error) {
	modelPath := filepath.FromSlash(modelName)
	if filepath.Ext(modelPath) != ".model" {
		return nil, fmt.Errorf("model file should have .model extension")
	}
	b, err := os.ReadFile(modelPath)
	if err != nil {
		return nil, err
	}
	lines := bytes.Lines(b)
	for line := range lines {
		if !bytes.Equal(bytes.TrimRight(line, crlf), []byte(Version)) {
			return nil, fmt.Errorf("version does not match")
		}
		break
	}

	idx := TokenId(MinVocabSize)
	merges := make(map[Pair]TokenId)
	for line := range lines {
		merge := strings.Split(string(bytes.TrimRight(line, crlf)), " ")
		if len(merge) != 2 {
			return nil, fmt.Errorf("malformed model file")
		}
		left, err := strconv.Atoi(merge[0])
		if err != nil {
			return nil, err
		}
		right, err := strconv.Atoi(merge[1])
		if err != nil {
			return nil, err
		}
		pair := Pair{Left: TokenId(left), Right: TokenId(right)}
		merges[pair] = idx
		idx++
	}
	tokenizer := Tokenizer{}
	tokenizer.merges = merges
	tokenizer.vocab = buildVocabFromMerges(merges)
	return &tokenizer, nil
}
