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

type Alphabet struct {
	root  *alphaNode
	runes []rune
	ln    int
	enc   [4]byte
}

func NewAlphabet(runes []rune) Alphabet {
	node := &alphaNode{}
	al := Alphabet{
		root:  node,
		runes: make([]rune, 0, len(runes)),
	}
	for _, rn := range runes {
		al.add(rn)
	}
	return al
}

func AlphabetFromReader(rdr io.Reader, scratch []byte) (Alphabet, error) {
	if scratch == nil {
		scratch = make([]byte, 8192)
	}

	buf := bufio.NewReader(rdr)

	a := NewAlphabet(nil)
	for {
		r, _, err := buf.ReadRune()
		if err == io.EOF {
			break
		} else if err != nil {
			return Alphabet{}, err
		}
		a.add(r)
	}

	return a, nil
}

func (al *Alphabet) Size() int {
	return al.ln
}

func (al *Alphabet) Find(rn rune) (pos int) {
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

func (al *Alphabet) add(rn rune) {
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
	}
}

func (al *Alphabet) MarshalBinary() (data []byte, err error) {
	return []byte(string(al.runes)), nil
}

func (al *Alphabet) UnmarshalBinary(data []byte) (err error) {
	*al = NewAlphabet(bytes.Runes(data))
	return nil
}

type alphaNode struct {
	next [256]*alphaNode
	set  bool
	pos  int
	sz   int
}
