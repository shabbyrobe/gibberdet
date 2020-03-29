package gibberdet

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"
)

func findPair(m *Model, ab string) float64 {
	runes := bytes.Runes([]byte(ab))
	if len(runes) != 2 {
		panic(nil)
	}
	ai := m.alpha.FindRune(runes[0])
	if ai < 0 {
		return -1
	}
	bi := m.alpha.FindRune(runes[1])
	if bi < 0 {
		return -1
	}
	return m.gram[ai*m.alpha.Len()+bi]
}

func TestModelASCIIFindGram(t *testing.T) {
	a := NewAlphabet([]rune("abc"))
	tr := NewTrainer(a)
	if err := tr.Add(strings.NewReader("aabbcc")); err != nil {
		t.Fatal(err)
	}
	m, err := tr.Compile()
	if err != nil {
		t.Fatal(err)
	}

	found1 := findPair(m, "aa")
	found2 := findPair(m, "ab")
	found3 := findPair(m, "bb")
	notFound := findPair(m, "ca")

	if found1 == notFound {
		t.Fatal()
	}
	if found1 != found2 || found1 != found3 {
		t.Fatal()
	}
}

func TestModelRuneFindGram(t *testing.T) {
	a := NewAlphabet([]rune("可界河落布意"))
	tr := NewTrainer(a)
	if err := tr.Add(strings.NewReader("可界河落布意")); err != nil {
		t.Fatal(err)
	}
	m, err := tr.Compile()
	if err != nil {
		t.Fatal(err)
	}

	found1 := findPair(m, "可界")
	found2 := findPair(m, "河落")
	found3 := findPair(m, "布意")
	notFound := findPair(m, "ca")

	if found1 == notFound {
		t.Fatal()
	}
	if found1 != found2 || found1 != found3 {
		t.Fatal()
	}
}

func TestModelASCIIScore(t *testing.T) {
	b, _ := ioutil.ReadFile("testdata/oanc-en.gibber")
	var m Model
	if err := m.UnmarshalBinary(b); err != nil {
		t.Fatal(err)
	}

	for idx, tc := range []struct {
		in    string
		score float64
	}{
		{"test", 0.025},
	} {
		t.Run(fmt.Sprintf("good/%d", idx), func(t *testing.T) {
			score := m.GibberScore(tc.in)
			if score < tc.score {
				t.Fatal(tc.in, score)
			}
			if score != m.GibberScoreBytes([]byte(tc.in)) {
				t.Fatal()
			}
		})
	}

	for idx, tc := range []struct {
		in    string
		score float64
	}{
		{"2c38qnuonuf", 0.004},
		{"*)J(*&)(J", 0.0002},
	} {
		t.Run(fmt.Sprintf("bad/%d", idx), func(t *testing.T) {
			score := m.GibberScore(tc.in)
			if score > tc.score {
				t.Fatal(tc.in, score)
			}
			if score != m.GibberScoreBytes([]byte(tc.in)) {
				t.Fatal()
			}
		})
	}
}

var BenchScoreResult float64

func BenchmarkGibberScoreByteDelegate(b *testing.B) {
	var m Model
	bts, err := ioutil.ReadFile("testdata/gutenberg-en.gibber")
	if err != nil {
		panic(err)
	}

	if err := m.UnmarshalBinary(bts); err != nil {
		panic(err)
	}

	b.ReportAllocs()
	b.Run("byte", func(b *testing.B) {
		input := []byte("hello world")
		for i := 0; i < b.N; i++ {
			BenchScoreResult = m.GibberScoreBytes(input)
		}
	})

	b.Run("string", func(b *testing.B) {
		input := "hello world"
		for i := 0; i < b.N; i++ {
			BenchScoreResult = m.GibberScore(input)
		}
	})
}

func BenchmarkGibberScoreRuneDelegate(b *testing.B) {
	var m Model
	bts, err := ioutil.ReadFile("testdata/test-cn.gibber")
	if err != nil {
		panic(err)
	}

	if err := m.UnmarshalBinary(bts); err != nil {
		panic(err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		BenchScoreResult = m.GibberScore("可界河落布意")
	}
}
