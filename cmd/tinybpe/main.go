package main

import (
	"fmt"

	"github.com/shadowy-pycoder/tinybpe"
)

func main() {
	tokenizer := tinybpe.NewTokenizer(512)
	tokenizer.Train("testdata/t8.shakespeare.txt", true)
	ids := tokenizer.Encode("Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.")
	fmt.Println(ids)
	decoded, _ := tokenizer.Decode(ids)
	fmt.Println(decoded)
	tokenizer.Save("test3")
	tokenizer.Load("./models/test3.model")
	ids = tokenizer.Encode("Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.")
	fmt.Println(ids)
	decoded, _ = tokenizer.Decode(ids)
	fmt.Println(decoded)
	ids = append(ids, 1337)
	decoded, err := tokenizer.Decode(ids)
	if err != nil {
		fmt.Println(err)
	}
}
