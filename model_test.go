package gibberdet

import (
	"bytes"
	"io/ioutil"
	"strings"
	"testing"
)

func findPair(m *Model, ab string) float64 {
	runes := bytes.Runes([]byte(ab))
	if len(runes) != 2 {
		panic(nil)
	}
	ai := m.alpha.Find(runes[0])
	bi := m.alpha.Find(runes[1])
	return m.gram[ai*m.alpha.ln+bi]
}

func TestModel(t *testing.T) {
	a := NewAlphabet([]rune("abc"))
	m := NewModel(a)
	if err := m.Train(strings.NewReader("aabbcc")); err != nil {
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

var BenchScoreResult float64

func BenchmarkGibberScore(b *testing.B) {
	var m Model
	bts, err := ioutil.ReadFile("model.gibber")
	if err != nil {
		panic(err)
	}

	if err := m.UnmarshalBinary(bts); err != nil {
		panic(err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		BenchScoreResult = m.GibberScore("hello world")
	}
}
