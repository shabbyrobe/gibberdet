package gibberdet

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"unicode/utf8"
)

type Model struct {
	alpha          Alphabet
	ascii          *asciiAlphabet
	gram           []float64
	zeroGram       float64
	gibberStringFn func(string) float64
	gibberBytesFn  func([]byte) float64
	fast           bool
}

func (m *Model) init() {
	const zeroGramWeight = 2
	m.zeroGram = math.Log(1/(float64(m.alpha.Len()))) * zeroGramWeight

	var ok bool
	if m.ascii, ok = m.alpha.(*asciiAlphabet); ok {
		m.gibberStringFn = m.gibberStringScoreByByte
		m.gibberBytesFn = m.gibberBytesScoreByByte
	} else {
		m.gibberStringFn = m.gibberStringScoreByRune
		m.gibberBytesFn = m.gibberBytesScoreByRune
	}
}

func (m *Model) Fast(active bool) (prev bool) {
	prev = m.fast
	m.fast = active
	return prev
}

func (m *Model) Alphabet() Alphabet {
	return m.alpha
}

func (m *Model) Test(goodInput []string, badInput []string) (thresh float64, err error) {
	if len(goodInput) == 0 || len(badInput) == 0 {
		return 0, fmt.Errorf("gibberdet: empty test")
	}

	var minGood = math.MaxFloat64
	var goodMiss int
	var maxBad float64
	for _, s := range goodInput {
		v := m.GibberScore(s)
		if v > 0 && v < minGood {
			minGood = v
		} else if v <= 0 {
			goodMiss++
		}
	}

	for _, s := range badInput {
		v := m.GibberScore(s)
		if v > maxBad {
			maxBad = v
		}
	}

	if float64(goodMiss)/float64(len(goodInput)) > 0.2 {
		return thresh, fmt.Errorf("gibberdet: test failed; too many good items not found in model: %d / %d", goodMiss, len(goodInput))
	}

	thresh = (minGood + maxBad) / 2
	if minGood <= maxBad {
		return thresh, fmt.Errorf("gibberdet: test failed; good threshold %f is less than bad %f", minGood, maxBad)
	}

	return thresh, nil
}

func (m *Model) GibberScore(s string) float64 {
	return m.gibberStringFn(s)
}

func (m *Model) GibberScoreBytes(s []byte) float64 {
	return m.gibberBytesFn(s)
}

func (m *Model) gibberStringScoreByByte(s string) float64 {
	if len(s) < 2 {
		return 0
	}

	// Return the average transition prob from l through log_prob_mat.
	var logProb float64

	var alphaA, alphaB int
	var alphaLen = m.ascii.Len()

	i := 0

	alphaA = m.ascii.FindByte(s[i])

pair:
	alphaB = m.ascii.FindByte(s[i])
	if alphaA < 0 || alphaB < 0 {
		logProb += m.zeroGram
	} else {
		logProb += m.gram[alphaA*alphaLen+alphaB]
	}
	alphaA = alphaB

	i++
	if i >= len(s) {
		goto done
	}
	goto pair

done:
	// The exponentiation translates from log probs to probs.
	if m.fast {
		return expFast(logProb / float64(len(s)-1))
	}
	return math.Exp(logProb / float64(len(s)-1))
}

func (m *Model) gibberBytesScoreByByte(s []byte) float64 {
	if len(s) < 2 {
		return 0
	}

	// Return the average transition prob from l through log_prob_mat.
	var logProb float64

	var alphaA, alphaB int
	var alphaLen = m.ascii.Len()

	i := 0

	alphaA = m.ascii.FindByte(s[i])

pair:
	alphaB = m.ascii.FindByte(s[i])
	if alphaA < 0 || alphaB < 0 {
		logProb += m.zeroGram
	} else {
		logProb += m.gram[alphaA*alphaLen+alphaB]
	}
	alphaA = alphaB

	i++
	if i >= len(s) {
		goto done
	}
	goto pair

done:
	// The exponentiation translates from log probs to probs.
	if m.fast {
		return expFast(logProb / float64(len(s)-1))
	}
	return math.Exp(logProb / float64(len(s)-1))
}

func (m *Model) gibberBytesScoreByRune(s []byte) float64 {
	if utf8.RuneCount(s) < 2 {
		return 0
	}

	// Return the average transition prob from l through log_prob_mat.
	var logProb float64

	var last int
	var first = true
	var alphaLen = m.alpha.Len()

	for i := 0; i < len(s); {
		r, sz := utf8.DecodeRune(s[i:])
		i += sz

		alphaIdx := m.alpha.FindRune(r)
		if first {
			first = false
		} else if alphaIdx < 0 && last < 0 {
			logProb += m.zeroGram
		} else {
			logProb += m.gram[last*alphaLen+alphaIdx]
		}
		last = alphaIdx
	}

	// The exponentiation translates from log probs to probs.
	if m.fast {
		return expFast(logProb / float64(len(s)-1))
	}
	return math.Exp(logProb / float64(len(s)-1))
}

func (m *Model) gibberStringScoreByRune(s string) float64 {
	if utf8.RuneCountInString(s) < 2 {
		return 0
	}

	// Return the average transition prob from l through log_prob_mat.
	var logProb float64

	var last int
	var first = true
	var alphaLen = m.alpha.Len()

	for _, r := range s {
		alphaIdx := m.alpha.FindRune(r)
		if alphaIdx < 0 {
			if !first {
				first = true
			}
			continue
		}
		if first {
			first = false
		} else {
			logProb += m.gram[last*alphaLen+alphaIdx]
		}
		last = alphaIdx
	}

	// The exponentiation translates from log probs to probs.
	if m.fast {
		return expFast(logProb / float64(len(s)-1))
	}
	return math.Exp(logProb / float64(len(s)-1))
}

func (m *Model) MarshalBinary() (data []byte, err error) {
	alpha, err := marshalAlphabet(m.alpha)
	if err != nil {
		return nil, err
	}

	var enc = make([]byte, 8)
	var buf bytes.Buffer

	binary.LittleEndian.PutUint32(enc, uint32(len(alpha)))
	buf.Write(enc[:4])

	buf.Write(alpha)

	binary.LittleEndian.PutUint32(enc, uint32(len(m.gram)))
	buf.Write(enc[:4])

	for _, f := range m.gram {
		bits := math.Float64bits(f)
		binary.LittleEndian.PutUint64(enc, bits)
		buf.Write(enc)
	}

	var outer bytes.Buffer
	outer.WriteString("gibbermodel!")
	binary.LittleEndian.PutUint32(enc, uint32(buf.Len()))
	outer.Write(enc[:4])
	outer.Write(buf.Bytes())

	return outer.Bytes(), nil
}

func (m *Model) UnmarshalBinary(data []byte) (err error) {
	if !bytes.HasPrefix(data, []byte("gibbermodel!")) {
		return fmt.Errorf("gibberdet: model does not start with 'gibbermodel!'")
	}

	pos := len("gibbermodel!")
	sz := int(binary.LittleEndian.Uint32(data[pos:]))
	if len(data)-pos < sz {
		return fmt.Errorf("gibberdet: model size mismatch")
	}
	pos += 4

	alphaSz := int(binary.LittleEndian.Uint32(data[pos:]))
	pos += 4
	alpha := bytes.Runes(data[pos : pos+alphaSz])
	pos += alphaSz

	gramSz := int(binary.LittleEndian.Uint32(data[pos:]))
	pos += 4

	grams := make([]float64, 0, gramSz)
	if pos+(gramSz*8) != len(data) {
		return fmt.Errorf("gibberdet: gram data size mismatch")
	}
	for ; pos < len(data); pos += 8 {
		u := binary.LittleEndian.Uint64(data[pos:])
		grams = append(grams, math.Float64frombits(u))
	}

	*m = Model{
		fast:  true,
		alpha: NewAlphabet(alpha),
		gram:  grams,
	}
	m.init()

	return nil
}
