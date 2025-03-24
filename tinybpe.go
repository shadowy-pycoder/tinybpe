package tinybpe

import (
	"bytes"
	"cmp"
	"fmt"
	"log"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

const (
	minVocabSize int = 256
	maxVocabSize int = int(^uint(0) >> 1)
	crlf             = "\r\n"
)

type TokenId int

type Pair struct {
	Left  TokenId
	Right TokenId
}

func replacePairWithNewToken(pair Pair, newTokenId TokenId, ids []TokenId) []TokenId {
	newIds := make([]TokenId, 0, len(ids))
	idx := 0
	for idx < len(ids) {
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
	vocab     map[TokenId][]byte
	vocabSize int
	merges    map[Pair]TokenId
}

func buildVocabFromMerges(merges map[Pair]TokenId) map[TokenId][]byte {
	vocab := make(map[TokenId][]byte, len(merges)+minVocabSize)
	for i := range minVocabSize {
		vocab[TokenId(i)] = []byte{byte(i)}
	}
	for _, pair := range slices.SortedStableFunc(maps.Keys(merges), func(a, b Pair) int {
		return cmp.Compare(merges[a], merges[b])
	}) {
		vocab[merges[pair]] = joinBytes(vocab[pair.Left], vocab[pair.Right])
	}
	return vocab
}

func NewTokenizer(vocabSize int) *Tokenizer {
	if vocabSize < minVocabSize || vocabSize == maxVocabSize {
		log.Fatalf("vocabSize must be within [%d, %d) range\n", minVocabSize, maxVocabSize)
	}
	tokenizer := Tokenizer{vocabSize: vocabSize}
	tokenizer.merges = make(map[Pair]TokenId)
	tokenizer.vocab = buildVocabFromMerges(tokenizer.merges)
	return &tokenizer
}

func (t *Tokenizer) getMaxPair(ids []TokenId) Pair {
	counts := make(map[Pair]int, len(ids)/2)
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

func (t *Tokenizer) getMinPair(ids []TokenId) Pair {
	counts := make(map[Pair]int, len(ids)/2)
	var pair Pair
	var count int
	for i := range len(ids) - 1 {
		pair = Pair{Left: ids[i], Right: ids[i+1]}
		count = counts[pair] + 1
		counts[pair] = count
	}
	minCount := TokenId(maxVocabSize)
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

func (t *Tokenizer) Train(path string, verbose bool) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	ids := make([]TokenId, len(b))
	for i := range b {
		ids[i] = TokenId(b[i])
	}
	iterNum := t.vocabSize - minVocabSize
	for i := range iterNum {
		maxPair := t.getMaxPair(ids)
		newTokenId := TokenId(minVocabSize + i)
		t.vocab[newTokenId] = joinBytes(t.vocab[maxPair.Left], t.vocab[maxPair.Right])
		t.merges[maxPair] = newTokenId
		ids = replacePairWithNewToken(maxPair, newTokenId, ids)
		if verbose {
			fmt.Printf("Iteration %d/%d: [%d, %d] -> %d [%q]\n", i+1,
				iterNum,
				maxPair.Left,
				maxPair.Right,
				newTokenId,
				t.vocab[newTokenId])
		}
	}
	return nil
}

func (t *Tokenizer) Decode(ids []TokenId) string {
	var sb strings.Builder
	for _, idx := range ids {
		sb.Write(t.vocab[idx])
	}
	return sb.String()
}

func (t *Tokenizer) Encode(text string) []TokenId {
	ids := make([]TokenId, len(text))
	for i, idx := range []byte(text) {
		ids[i] = TokenId(idx)
	}
	for len(ids) >= 2 {
		pair := t.getMinPair(ids)
		if pair.Left == TokenId(maxVocabSize) || pair.Right == TokenId(maxVocabSize) {
			break
		}
		ids = replacePairWithNewToken(pair, t.merges[pair], ids)
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

func (t *Tokenizer) Load(modelName string) error {
	modelPath := filepath.FromSlash(modelName)
	if filepath.Ext(modelPath) != ".model" {
		return fmt.Errorf("model file should have .model extension")
	}
	b, err := os.ReadFile(modelPath)
	if err != nil {
		return err
	}
	lines := bytes.Lines(b)
	for line := range lines {
		if !bytes.Equal(bytes.TrimRight(line, crlf), []byte(Version)) {
			return fmt.Errorf("version does not match")
		}
		break
	}

	idx := TokenId(minVocabSize)
	merges := make(map[Pair]TokenId)
	for line := range lines {
		merge := strings.Split(string(bytes.TrimRight(line, crlf)), " ")
		if len(merge) != 2 {
			return fmt.Errorf("malformed file")
		}
		left, err := strconv.Atoi(merge[0])
		if err != nil {
			return err
		}
		right, err := strconv.Atoi(merge[1])
		if err != nil {
			return err
		}
		pair := Pair{Left: TokenId(left), Right: TokenId(right)}
		merges[pair] = idx
		idx++
	}
	t.merges = merges
	t.vocab = buildVocabFromMerges(merges)
	return nil
}
