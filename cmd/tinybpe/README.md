# TinyBPE CLI

## Installation

You can download the binary for your platform from [Releases](https://github.com/shadowy-pycoder/tinybpe/releases) page.

Alternatively, you can install it using `go install` command (requires Go [1.24](https://go.dev/doc/install) or later):

```shell
go install -ldflags "-s -w" -trimpath github.com/shadowy-pycoder/tinybpe/cmd/tinybpe@latest
```
This will install the `tinybpe` binary to your `$GOPATH/bin` directory.

Another alternative is to build from source:

```shell
git clone https://github.com/shadowy-pycoder/tinybpe.git
cd tinybpe
make build
./bin/tinybpe
```

## Usage

```shell
tinybpe help
                                                                  
TinyBPE by shadowy-pycoder 

GitHub: https://github.com/shadowy-pycoder/tinybpe

Usage: tinybpe COMMAND

Commands:

  help         Show help message and exit
  version      Show version and exit
  train        Start training process from the set of files or stdin 
  encode       Tokenize the set of files or stdin with trained model
  decode       Convert tokens to text with trained model
``` 

```shell
tinybpe train help

Usage tinybpe train [COMMAND] | [OPTIONS] [FILES... | STDIN] 
Options:

  -o string
    	the name of the model and vocab
  -s value
    	the size of vocabulary [256...9223372036854775807] (default 512)
  -v	show training progress

Commands:

  help         Show help message and exit
```

```shell
tinybpe encode help

Usage tinybpe encode [COMMAND] | -m MODEL [FILES... | STDIN]
Options:

  -m string
    	path to trained model

Commands:

  help         Show help message and exit
```

```shell
tinybpe decode help

Usage tinybpe decode [COMMAND] | -m MODEL [FILES... | STDIN]
Options:

  -m string
    	path to trained model

Commands:

  help         Show help message and exit
```

### Examples

Train model from `stdin`
```shell
cat ./testdata/t8.shakespeare.txt | tinybpe train -v -s 512 -o test
```
or from file path (you can specify multiple file paths separated space)
```shell
tinybpe train -v -s 512 -o test ./testdata/t8.shakespeare.txt
# Output:
# Iteration 1/256: [32, 32] -> 256 ["  "] 392445 occurrences
# ...
# Iteration 256/256: [274, 100] -> 511 ["end"] 2250 occurrences
# Elapsed time: 1m1.659258082s
# Average time per iteration: 0.2409s
# Model saved: ./models/test.model
```

Encode with trained model `test`
```shell
echo -n "Hello World" | tinybpe encode -m ./models/test.model                        
# Output: [72,445,273,87,271,108,100]
```

Decode with trained model `test`
```shell
echo -n "[72,445,273,87,271,108,100]" | tinybpe decode -m ./models/test.model
# Output: Hello World
```

Example from [Wiki (Byte pair encoding)](https://en.wikipedia.org/wiki/Byte_pair_encoding)

```shell
tinybpe train -v -s 512 -o wiki ./testdata/wiki_bpe.txt 
# Iteration 1/3: [97, 97] -> 256 ["aa"] 4 occurrences
# Iteration 2/3: [256, 97] -> 257 ["aaa"] 2 occurrences
# Iteration 3/3: [257, 98] -> 258 ["aaab"] 2 occurrences
# Elapsed time: 65.834µs
# Average time per iteration: 0.0000s
# Model saved: ./models/wiki.model

tinybpe encode -m ./models/wiki.model ./testdata/wiki_bpe.txt
# [258,100,258,97,99]

echo -n "[258,100,258,97,99]"| tinybpe decode -m ./models/wiki.model
# aaabdaaabac
```

