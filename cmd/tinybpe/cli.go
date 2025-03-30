package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/shadowy-pycoder/tinybpe"
)

const (
	app             string = "tinybpe"
	defautVocabSize int    = 512
)

var (
	usagePrefix = `TinyBPE by shadowy-pycoder 

GitHub: https://github.com/shadowy-pycoder/tinybpe

Usage: tinybpe COMMAND`
	usageCommands = `
Commands:

  help         Show help message and exit
  version      Show version and exit
  train        Start training process from the set of files or stdin 
  encode       Tokenize the set of files or stdin with trained model
  decode       Convert tokens to text with trained model
`
	trainUsagePrefix = `Usage tinybpe train [COMMAND] | [OPTIONS] [FILES... | STDIN] 
Options:
`
	trainUsageCommands = `
Commands:

  help         Show help message and exit
`
	encodeUsagePrefix = `Usage tinybpe encode [COMMAND] | -m MODEL [FILES... | STDIN]
Options:
`
	encodeUsageCommands = `
Commands:

  help         Show help message and exit
`
	decodeUsagePrefix = `Usage tinybpe decode [COMMAND] | -m MODEL [FILES... | STDIN]
Options:
`
	decodeUsageCommands = `
Commands:

  help         Show help message and exit
`
)

func read(input []string, buf *bytes.Buffer) error {
	var f []byte
	var err error
	if len(input) == 0 {
		f, err = io.ReadAll(os.Stdin)
		if err != nil {
			return err
		}
		buf.Write(f)
	} else {
		for _, file := range input {
			f, err = os.ReadFile(file)
			if err != nil {
				return err
			}
			buf.Write(f)
		}
	}
	return nil
}

type runner interface {
	init([]string) error
	run() error
	name() string
}

type cmdTrain struct {
	fs *flag.FlagSet

	buf       bytes.Buffer
	modelName string
	size      int
	verbose   bool
}

func newCmdTrain() *cmdTrain {
	ctr := &cmdTrain{
		fs: flag.NewFlagSet("train", flag.ContinueOnError),
	}
	ctr.size = defautVocabSize
	errMsg := fmt.Sprintf("%s: values for -s flag should be within the range [%d...%d]", app, tinybpe.MinVocabSize, tinybpe.MaxVocabSize)
	ctr.fs.StringVar(&ctr.modelName, "o", "", "the name of the model and vocab")
	ctr.fs.Func("s", fmt.Sprintf("the size of vocabulary [%d...%d] (default %d)",
		tinybpe.MinVocabSize, tinybpe.MaxVocabSize, defautVocabSize), func(flagValue string) error {
		i, err := strconv.Atoi(flagValue)
		if err != nil {
			fmt.Fprintln(os.Stderr, errMsg)
			os.Exit(2)
		}
		if i < tinybpe.MinVocabSize || i >= tinybpe.MaxVocabSize {
			fmt.Fprintln(os.Stderr, errMsg)
			os.Exit(2)
		}
		ctr.size = i
		return nil
	})
	ctr.fs.BoolVar(&ctr.verbose, "v", false, "show training progress")
	return ctr
}

func (ctr *cmdTrain) name() string {
	return ctr.fs.Name()
}

func (ctr *cmdTrain) init(args []string) error {
	ctr.fs.Usage = func() {
		fmt.Println(trainUsagePrefix)
		ctr.fs.PrintDefaults()
		fmt.Println(trainUsageCommands)
	}
	if len(args) > 0 && args[0] == "help" {
		ctr.fs.Usage()
		os.Exit(0)
	}
	return ctr.fs.Parse(args)
}

func (ctr *cmdTrain) run() error {
	if err := read(ctr.fs.Args(), &ctr.buf); err != nil {
		return err
	}
	if ctr.modelName == "" {
		ctr.modelName = fmt.Sprintf("model_%d_%v", ctr.size, time.Now().Unix())
	}
	start := time.Now()
	tokenizer := tinybpe.NewTokenizer()
	tokenizer.Train(ctr.buf.Bytes(), ctr.size, ctr.verbose)
	if ctr.verbose {
		elapsed := time.Since(start)
		fmt.Printf("Elapsed time: %s\n", elapsed)
		fmt.Printf("Average time per iteration: %.4fs\n", elapsed.Seconds()/float64(ctr.size-tinybpe.MinVocabSize))
	}
	err := tokenizer.Save(ctr.modelName)
	if err != nil {
		return err
	}
	fmt.Printf("Model saved: ./models/%s.model\n", ctr.modelName)
	return nil
}

type cmdEncode struct {
	fs *flag.FlagSet

	buf   bytes.Buffer
	model string
}

func newCmdEncode() *cmdEncode {
	cen := &cmdEncode{
		fs: flag.NewFlagSet("encode", flag.ContinueOnError),
	}
	cen.fs.StringVar(&cen.model, "m", "", "path to trained model")
	return cen
}

func (cen *cmdEncode) name() string {
	return cen.fs.Name()
}

func (cen *cmdEncode) init(args []string) error {
	cen.fs.Usage = func() {
		fmt.Println(encodeUsagePrefix)
		cen.fs.PrintDefaults()
		fmt.Println(encodeUsageCommands)
	}
	if len(args) > 0 && args[0] == "help" {
		cen.fs.Usage()
		os.Exit(0)
	}
	return cen.fs.Parse(args)
}

func (cen *cmdEncode) run() error {
	if cen.model == "" {
		return fmt.Errorf("model path is empty")
	}
	tokenizer, err := tinybpe.Load(cen.model)
	if err != nil {
		return err
	}
	if err := read(cen.fs.Args(), &cen.buf); err != nil {
		return err
	}
	tokens := tokenizer.Encode(cen.buf.Bytes())
	encoded, err := json.Marshal(tokens)
	if err != nil {
		return err
	}
	fmt.Println(string(encoded))
	return nil
}

type cmdDecode struct {
	fs *flag.FlagSet

	buf   bytes.Buffer
	model string
}

func newCmdDecode() *cmdDecode {
	cde := &cmdDecode{
		fs: flag.NewFlagSet("decode", flag.ContinueOnError),
	}
	cde.fs.StringVar(&cde.model, "m", "", "path to trained model")
	return cde
}

func (cde *cmdDecode) name() string {
	return cde.fs.Name()
}

func (cde *cmdDecode) init(args []string) error {
	cde.fs.Usage = func() {
		fmt.Println(encodeUsagePrefix)
		cde.fs.PrintDefaults()
		fmt.Println(encodeUsageCommands)
	}
	if len(args) > 0 && args[0] == "help" {
		cde.fs.Usage()
		os.Exit(0)
	}
	return cde.fs.Parse(args)
}

func (cde *cmdDecode) run() error {
	if cde.model == "" {
		return fmt.Errorf("model path is empty")
	}
	tokenizer, err := tinybpe.Load(cde.model)
	if err != nil {
		return err
	}
	if err := read(cde.fs.Args(), &cde.buf); err != nil {
		return err
	}
	var tokens []tinybpe.TokenId
	err = json.NewDecoder(&cde.buf).Decode(&tokens)
	if err != nil {
		return fmt.Errorf("can't decode tokens: %w", err)
	}
	text, err := tokenizer.Decode(tokens)
	if err != nil {
		return err
	}
	fmt.Println(text)
	return nil
}

func root(args []string) error {
	flags := flag.NewFlagSet(app, flag.ExitOnError)
	flags.Usage = func() {
		fmt.Println(usagePrefix)
		flags.PrintDefaults()
		fmt.Println(usageCommands)
	}
	if len(args) == 0 {
		return fmt.Errorf("unknown subcommand")
	}
	subcommand := args[0]

	if subcommand == "version" {
		fmt.Println(tinybpe.Version)
		return nil
	} else if subcommand == "help" {
		flags.Usage()
		return nil
	}
	cmds := []runner{
		newCmdTrain(),
		newCmdEncode(),
		newCmdDecode(),
	}
	for _, cmd := range cmds {
		if cmd.name() == subcommand {
			cmd.init(os.Args[2:])
			return cmd.run()
		}
	}
	return fmt.Errorf("unknown subcommand")
}
