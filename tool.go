//+build ignore

package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/shabbyrobe/gibberdet"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	if len(os.Args) < 2 {
		return fmt.Errorf("usage: tool.go (build|test|use)")
	}
	switch os.Args[1] {
	case "build":
		return build(os.Args[2:])
	case "test":
		return test(os.Args[2:])
	case "use":
		return use(os.Args[2:])
	default:
		return fmt.Errorf("unknown command")
	}
}

func build(args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("usage: tool.go build <infile> <outfile>")
	}

	bts, err := ioutil.ReadFile(args[0])
	if err != nil {
		return err
	}

	a, err := gibberdet.AlphabetFromReader(bytes.NewReader(bts), nil)
	if err != nil {
		return err
	}

	m := gibberdet.NewModel(a)
	if err := m.Train(bytes.NewReader(bts)); err != nil {
		return err
	}

	enc, err := m.MarshalBinary()
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(args[1], enc, 0644); err != nil {
		return err
	}

	return nil
}

func test(args []string) error {
	return nil
}

func use(args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("usage: tool.go use <model> <teststr>")
	}

	bts, err := ioutil.ReadFile(args[0])
	if err != nil {
		return err
	}

	var m gibberdet.Model
	if err := m.UnmarshalBinary(bts); err != nil {
		return err
	}

	m.Fast(true)
	fmt.Println("Slow:", m.GibberScore(args[1]))
	m.Fast(false)
	fmt.Println("Fast:", m.GibberScore(args[1]))

	return nil
}
