# TinyBPE Library

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Reference](https://pkg.go.dev/badge/github.com/shadowy-pycoder/tinybpe.svg)](https://pkg.go.dev/github.com/shadowy-pycoder/tinybpe)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/shadowy-pycoder/tinybpe)
[![Go Report Card](https://goreportcard.com/badge/github.com/shadowy-pycoder/tinybpe)](https://goreportcard.com/report/github.com/shadowy-pycoder/tinybpe)
![GitHub Release](https://img.shields.io/github/v/release/shadowy-pycoder/tinybpe)
![GitHub Downloads (all assets, all releases)](https://img.shields.io/github/downloads/shadowy-pycoder/tinybpe/total)


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

# TinyBPE CLI

See here: [README.md](./cmd/tinybpe/README.md)