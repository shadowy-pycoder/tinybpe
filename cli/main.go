package main

import (
	"fmt"

	"github.com/shadowy-pycoder/tinybpe"
)

func main() {
	tokenizer := tinybpe.NewTokenizer(512)
	tokenizer.Train("testdata/t8.shakespeare.txt", true)
	ids := tokenizer.Encode("Hello World")
	fmt.Println(ids)
	fmt.Println(tokenizer.Decode(ids))
	tokenizer.Save("test")
}
