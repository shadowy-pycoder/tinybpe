# TinyBPE Library

## Installation

```shell
go get github.com/shadowy-pycoder/tinybpe@latest
```

## Usage

```go
package main

import (
	"fmt"
	"os"

	"github.com/shadowy-pycoder/tinybpe"
)

func main() {
	tokenizer := tinybpe.NewTokenizer()
	f, err := os.ReadFile("./testdata/t8.shakespeare.txt")
	if err != nil {
		panic(err)
	}
	vocabSize := 512
	verbose := true
	tokenizer.Train(f, vocabSize, verbose)
	if err := tokenizer.Save("test"); err != nil {
		panic(err)
	}
	tokens := tokenizer.Encode([]byte("Hello World"))
	fmt.Println(tokens)
	text, err := tokenizer.Decode(tokens)
	if err != nil {
		panic(err)
	}
	fmt.Println(text)
}
```