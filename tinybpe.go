package tinybpe

import (
	"log"
	"maps"
	"os"
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
	vocab        map[TokenId][]byte
	vocabSize    int
	inverseVocab map[Pair]TokenId
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
	tokenizer.inverseVocab = make(map[Pair]TokenId)
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
	for pr := range maps.Keys(counts) {
		tokenId, ok := t.inverseVocab[pr]
		if ok && tokenId < minCount {
			minCount = tokenId
			minPair = pr
		}
	}
	return minPair
}
func (t *Tokenizer) Train(path string) {
	b, err := os.ReadFile(path)
	if err != nil {
		log.Fatalln("Failed opening file")
	}
	ids := make([]TokenId, len(b))
	for i := range b {
		ids[i] = TokenId(b[i])
	}
	for i := range t.vocabSize - minVocabSize {
		maxPair := t.getMaxPair(ids)
		newTokenId := TokenId(minVocabSize + i)
		t.vocab[newTokenId] = append(t.vocab[maxPair.Left], t.vocab[maxPair.Right]...)
		t.inverseVocab[maxPair] = newTokenId
		ids = replacePairWithNewToken(maxPair, newTokenId, ids)
	}
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
		ids = replacePairWithNewToken(pair, t.inverseVocab[pair], ids)
	}
	return ids
}
