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
	"strings"
)

const minVocabSize int = 256
const maxVocabSize int = int(^uint(0) >> 1)

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

func NewTokenizer(vocabSize int) *Tokenizer {
	if vocabSize < minVocabSize || vocabSize == maxVocabSize {
		log.Fatalf("vocabSize must be within [%d, %d) range\n", minVocabSize, maxVocabSize)
	}
	tokenizer := Tokenizer{vocabSize: vocabSize}
	vocab := make(map[TokenId][]byte, vocabSize)
	for i := range minVocabSize {
		vocab[TokenId(i)] = []byte{byte(i)}
	}
	tokenizer.vocab = vocab
	tokenizer.merges = make(map[Pair]TokenId)
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
			fmt.Printf("Iteration %d/%d: [%d, %d] -> %d (%q)\n", i+1,
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

func (t *Tokenizer) buildVocabFromMerges() {
	vocab := make(map[TokenId][]byte, minVocabSize)
	for i := range minVocabSize {
		vocab[TokenId(i)] = []byte{byte(i)}
	}
	t.vocab = vocab
	for _, pair := range slices.SortedStableFunc(maps.Keys(t.merges), func(a, b Pair) int {
		return cmp.Compare(t.merges[a], t.merges[b])
	}) {
		t.vocab[t.merges[pair]] = joinBytes(t.vocab[pair.Left], t.vocab[pair.Right])
	}
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
