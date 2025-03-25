package tinybpe

import "testing"

func BenchmarkGetAnswers(b *testing.B) {
	b.ResetTimer()
	numIter := b.N
	if numIter < minVocabSize {
		numIter += minVocabSize
	}

	tokenizer := NewTokenizer(numIter)
	tokenizer.Train("testdata/t8.shakespeare.txt", false)
	tokenizer.Save("test")
}
