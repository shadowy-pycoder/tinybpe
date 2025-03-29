package tinybpe

import (
	"os"
	"testing"
)

func BenchmarkTrainAndSave(b *testing.B) {
	if b.N > 1 {
		b.ResetTimer()
		tokenizer := NewTokenizer(b.N)
		tokenizer.Train("testdata/t8.shakespeare.txt", false)
		tokenizer.Save("test")
	}
}

func BenchmarkLoad(b *testing.B) {
	if b.N > 1 {
		b.ResetTimer()
		if _, err := os.Stat("models/test.model"); err == nil {
			tokenizer := NewTokenizer(b.N)
			for b.Loop() {
				tokenizer.Load("models/test.model")
			}
		} else {
			b.Skip("no test model to load")
		}
	}
}
