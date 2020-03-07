package gibberdet

import (
	"bytes"
	"strings"
	"testing"
	"unicode/utf8"
)

func findPair(m *Model, ab string) float64 {
	runes := bytes.Runes([]byte(ab))
	if len(runes) != 2 {
		panic(nil)
	}
	ae, be := make([]byte, 4), make([]byte, 4)
	ae = ae[:utf8.EncodeRune(ae, runes[0])]
	be = be[:utf8.EncodeRune(be, runes[1])]
	ai := m.alpha.Find(ae)
	bi := m.alpha.Find(be)
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
