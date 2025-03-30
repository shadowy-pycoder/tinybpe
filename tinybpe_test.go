package tinybpe

import (
	"os"
	"testing"
)

func BenchmarkTrainAndSave(b *testing.B) {
	if b.N > 1 {
		b.ResetTimer()
		tokenizer := NewTokenizer()
		f, err := os.ReadFile("testdata/t8.shakespeare.txt")
		if err != nil {
			b.Fatal(err)
		}
		tokenizer.Train(f, b.N, false)
		err = tokenizer.Save("test")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkLoad(b *testing.B) {
	if b.N > 1 {
		b.ResetTimer()
		if _, err := os.Stat("models/test.model"); err == nil {
			for b.Loop() {
				Load("models/test.model")
			}
		} else {
			b.Skip("no test model to load")
		}
	}
}
