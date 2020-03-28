package gibberdet

import (
	"bufio"
	"bytes"
	"io"
)

const (
	alpha   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	numeric = "0123456789"
	wsp     = " "
)

var (
	ASCIIAlpha = NewAlphabet([]rune(alpha + wsp))
	ASCIIAlnum = NewAlphabet([]rune(alpha + numeric + wsp))
)

type Alphabet interface {
	Runes() []rune
	Len() int
	FindRune(rn rune) (pos int)
	FindByte(b byte) (pos int)
}

type alphaNode struct {
	next [256]*alphaNode
	set  bool
	pos  int
	sz   int
}

type runeAlphabet struct {
	root  *alphaNode
	runes []rune
	ln    int
	max   rune
	enc   [4]byte
}

func NewAlphabet(runes []rune) Alphabet {
	ra := newRuneAlphabet(runes)
	if ra.max < 128 {
		return ra.asASCII()
	}
	return ra
}

func AlphabetFromReader(rdr io.Reader, scratch []byte) (Alphabet, error) {
	if scratch == nil {
		scratch = make([]byte, 8192)
	}

	buf := bufio.NewReader(rdr)

	a := newRuneAlphabet(nil)
	for {
		r, _, err := buf.ReadRune()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
		a.add(r)
	}

	if a.max < 128 {
		return a.asASCII(), nil
	}

	return a, nil
}

func newRuneAlphabet(runes []rune) *runeAlphabet {
	node := &alphaNode{}
	al := &runeAlphabet{
		root:  node,
		runes: make([]rune, 0, len(runes)),
	}
	for _, rn := range runes {
		al.add(rn)
	}
	return al
}

func (al *runeAlphabet) Runes() []rune {
	return al.runes
}

func (al *runeAlphabet) Len() int {
	return al.ln
}

func (al *runeAlphabet) FindByte(b byte) (pos int) {
	return al.FindRune(rune(b))
}

func (al *runeAlphabet) FindRune(rn rune) (pos int) {
	cur := al.root
	n := 0
	for rv := rn; rv > 0; rv >>= 8 {
		b := uint8(rv)
		if cur.next[b] == nil {
			break
		}
		cur = cur.next[b]
		n++
	}
	if cur == nil || !cur.set || cur.sz != n {
		return -1
	}
	return cur.pos
}

func (al *runeAlphabet) add(rn rune) {
	cur := al.root
	n := 0
	for rv := rn; rv > 0; rv >>= 8 {
		b := uint8(rv)
		if cur.next[b] == nil {
			cur.next[b] = &alphaNode{}
		}
		cur = cur.next[b]
		n++
	}
	if !cur.set {
		al.runes = append(al.runes, rn)
		cur.set = true
		cur.pos = al.ln
		cur.sz = n
		al.ln++
		if rn > al.max {
			al.max = rn
		}
	}
}

func (al *runeAlphabet) asASCII() *asciiAlphabet {
	return newASCIIAlphabet(al.runes)
}

type asciiAlphabet struct {
	pos   [256]int // 256 instead of 128 to avoid bounds check
	runes []rune
}

func newASCIIAlphabet(runes []rune) *asciiAlphabet {
	al := &asciiAlphabet{
		runes: runes,
	}
	_ = al.pos[255]
	for i := 0; i < 256; i++ {
		al.pos[i] = -1
	}

	for idx, rn := range runes {
		if rn > 127 {
			panic("expected ASCII")
		}
		b := byte(rn)
		al.pos[b] = idx
	}
	return al
}

func (al *asciiAlphabet) Runes() []rune {
	return al.runes
}

func (al *asciiAlphabet) Len() int {
	return len(al.runes)
}

func (al *asciiAlphabet) FindByte(b byte) (pos int) {
	return al.pos[b]
}

func (al *asciiAlphabet) FindRune(rn rune) (pos int) {
	if rn > 127 {
		return 0
	}
	return al.pos[byte(rn)]
}

func marshalAlphabet(a Alphabet) (data []byte, err error) {
	return []byte(string(a.Runes())), nil
}

func unmarshalAlphabet(data []byte, into *Alphabet) (err error) {
	*into = NewAlphabet(bytes.Runes(data))
	return nil
}
