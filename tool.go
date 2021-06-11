//+build ignore

package main

import (
	"archive/zip"
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"unicode/utf8"

	"github.com/shabbyrobe/gibberdet"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	if len(os.Args) < 2 {
		return fmt.Errorf("usage: tool.go (train|test|gib|gibfile|oanc)")
	}
	switch os.Args[1] {
	case "train":
		return train(os.Args[2:])
	case "test":
		return test(os.Args[2:])
	case "gib":
		return gib(os.Args[2:])
	case "gibfile":
		return gibfile(os.Args[2:])
	case "oanc":
		return oanc(os.Args[2:])
	default:
		return fmt.Errorf("unknown command")
	}
}

func train(args []string) error {
	var alphaKind = "asciialnum"
	var alphaFile string

	fs := flag.NewFlagSet("", 0)
	fs.StringVar(&alphaKind, "alphakind", "asciialnum", ""+
		"Alphabet to use. Accepts 'asciialpha', 'asciialnum', 'asciifile' or 'runefile'")
	fs.StringVar(&alphaFile, "alphafile", "", ""+
		"File containing alphabet")
	if err := fs.Parse(args); err != nil {
		return err
	}

	args = fs.Args()
	if len(args) != 2 {
		return fmt.Errorf(
			"usage: tool.go build -alphakind (asciialnum|asciialpha|asciifile|runefile) " +
				"-alphafile=<alphafile> <infile> <outfile>")
	}

	inFile, outFile := args[0], args[1]

	var a gibberdet.Alphabet
	switch alphaKind {
	case "asciialnum":
		a = gibberdet.ASCIIAlnum
	case "asciialpha":
		a = gibberdet.ASCIIAlpha
	case "asciifile", "runefile":
		af, err := os.Open(alphaFile)
		if err != nil {
			return err
		}
		defer af.Close()

		var rdr io.Reader = af
		if alphaKind == "asciifile" {
			rdr = &filterAsciiReader{af}
		}

		a, err = gibberdet.AlphabetFromReader(rdr, nil)
		if err != nil {
			return err
		}
		af.Close()
	}

	f, err := os.Open(inFile)
	if err != nil {
		return err
	}
	defer f.Close()

	m, err := gibberdet.Train(a, f)
	if err != nil {
		return err
	}

	enc, err := m.MarshalBinary()
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(outFile, enc, 0644); err != nil {
		return err
	}

	return nil
}

func test(args []string) error {
	if len(args) != 3 {
		return fmt.Errorf("usage: tool.go test <model> <goodfile> <badfile>")
	}

	bts, err := ioutil.ReadFile(args[0])
	if err != nil {
		return err
	}

	var m gibberdet.Model
	if err := m.UnmarshalBinary(bts); err != nil {
		return err
	}

	good, err := readStringList(args[1])
	if err != nil {
		return err
	}

	bad, err := readStringList(args[2])
	if err != nil {
		return err
	}

	thresh, err := m.Test(good, bad)
	if err != nil {
		return err
	}

	fmt.Println(thresh)

	return nil
}

func gibfile(args []string) error {
	var minSize int

	fs := flag.NewFlagSet("", 0)
	fs.IntVar(&minSize, "minsz", 5, "skip terms shorter than this many runes")
	if err := fs.Parse(args); err != nil {
		return err
	}
	args = fs.Args()

	if len(args) != 2 {
		return fmt.Errorf("usage: tool.go gibfile <model> <file>")
	}

	bts, err := ioutil.ReadFile(args[0])
	if err != nil {
		return err
	}

	strs, err := readStringList(args[1])
	if err != nil {
		return err
	}

	var m gibberdet.Model
	if err := m.UnmarshalBinary(bts); err != nil {
		return err
	}

	buf := bufio.NewWriter(os.Stdout)
	for _, s := range strs {
		if utf8.RuneCountInString(s) < minSize {
			continue
		}
		v := m.GibberScore(s)
		fmt.Fprintf(buf, "%0.8f\t%s\n", v, s)
	}
	buf.Flush()

	return nil
}

func gib(args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("usage: tool.go gib <model> <teststr>")
	}

	bts, err := ioutil.ReadFile(args[0])
	if err != nil {
		return err
	}

	var m gibberdet.Model
	if err := m.UnmarshalBinary(bts); err != nil {
		return err
	}

	fmt.Println("str:", m.GibberScore(args[1]))
	fmt.Println("bts:", m.GibberScoreBytes([]byte(args[1])))

	return nil
}

// Download and build a model from the Open ANC database:
// http://www.anc.org/data/oanc/download/
func oanc(args []string) error {
	if len(args) < 1 || len(args) > 2 {
		return fmt.Errorf("usage: tool.go oanc <dest> [<infile>]")
	}

	var inFile string
	if len(args) >= 2 {
		inFile = args[1]
	}

	withUnderscores := true

	if inFile == "" {
		url := "http://www.anc.org/OANC/OANC_GrAF.zip"

		tf, err := ioutil.TempFile("", "")
		if err != nil {
			return err
		}

		rs, err := http.Get(url)
		if err != nil {
			tf.Close()
			return err
		}
		fmt.Printf("downloading corpus %s to %s, %s kb (file is retained)\n", url, tf.Name(), rs.Header["Content-Length"])

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

	// Exclude numbers as a high incidence of numbers is usually indicative of gibberish
	train := gibberdet.NewTrainer(gibberdet.ASCIIAlphaWordPunct, gibberdet.TrainerPairWeight(0))

	for _, finf := range r.File {
		if filepath.Ext(finf.Name) == ".txt" {
			fmt.Println("adding", finf.Name)
			rc, err := finf.Open()
			if err != nil {
				return err
			}
			bts, err := ioutil.ReadAll(rc)
			if err != nil {
				rc.Close()
				return err
			}
			rc.Close()

			if err := train.Add(bytes.NewReader(bts)); err != nil {
				return err
			}

			// If you want the model not to penalise words_separated_by_underscores,
			// this should help:
			if withUnderscores {
				bts = spaceReplacePattern.ReplaceAll(bts, []byte("_"))
				if err := train.Add(bytes.NewReader(bts)); err != nil {
					return err
				}
			}
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

var spaceReplacePattern = regexp.MustCompile(`\s+`)

func readStringList(fname string) (out []string, err error) {
	bts, err := ioutil.ReadFile(fname)
	if err != nil {
		return nil, err
	}
	scn := bufio.NewScanner(bytes.NewReader(bts))
	for scn.Scan() {
		out = append(out, scn.Text())
	}
	return out, nil
}

type filterAsciiReader struct {
	rdr io.Reader
}

func (r *filterAsciiReader) Read(b []byte) (n int, err error) {
	n, err = r.rdr.Read(b)
	if n > 0 && err == io.EOF {
		err = nil
	}
	if err != nil {
		return n, err
	}

	var on int
	for i, o := 0, 0; i < n; i++ {
		if b[i] < 0x20 || b[i] > 0x7e {
			continue
		}
		if i != o {
			on++
			b[o] = b[i]
			o++
		}
	}
	return on, nil
}
