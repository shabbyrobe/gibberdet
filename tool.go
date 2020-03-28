//+build ignore

package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/shabbyrobe/gibberdet"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	if len(os.Args) < 2 {
		return fmt.Errorf("usage: tool.go (train|test|use|oanc)")
	}
	switch os.Args[1] {
	case "train":
		return train(os.Args[2:])
	case "test":
		return test(os.Args[2:])
	case "use":
		return use(os.Args[2:])
	case "oanc":
		return oanc(os.Args[2:])
	default:
		return fmt.Errorf("unknown command")
	}
}

func train(args []string) error {
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

	m, err := gibberdet.Train(a, bytes.NewReader(bts))
	if err != nil {
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

// Download and build a model from the Open ANC database:
// http://www.anc.org/data/oanc/download/
func oanc(args []string) error {
	if len(args) < 1 || len(args) > 2 {
		return fmt.Errorf("usage: tool.go oanc <dest> [<infile>]")
	}

	var inFile string
	if len(args) == 2 {
		inFile = args[1]
	}

	if inFile == "" {
		url := "http://www.anc.org/OANC/OANC_GrAF.zip"

		tf, err := ioutil.TempFile("", "")
		if err != nil {
			return err
		}

		rs, err := http.Get("")
		fmt.Printf("downloading corpus %s to %s, %s kb (file is retained)\n", url, tf.Name(), rs.Header["Content-Length"])
		if err != nil {
			tf.Close()
			return err
		}

		if _, err := io.Copy(tf, rs.Body); err != nil {
			tf.Close()
			rs.Body.Close()
			return err
		}
		rs.Body.Close()

		if err := tf.Close(); err != nil {
			return err
		}

		inFile = tf.Name()
	}

	r, err := zip.OpenReader(inFile)
	if err != nil {
		return err
	}
	defer r.Close()

	train := gibberdet.NewTrainer(gibberdet.ASCIIAlnum)

	for _, finf := range r.File {
		if filepath.Ext(finf.Name) == ".txt" {
			fmt.Println("adding", finf.Name)
			rc, err := finf.Open()
			if err != nil {
				return err
			}
			if err := train.Add(rc); err != nil {
				rc.Close()
				return err
			}
			rc.Close()
		}
	}

	model, err := train.Compile()
	if err != nil {
		return err
	}

	data, err := model.MarshalBinary()
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(args[0], data, 0600); err != nil {
		return err
	}

	return nil
}
