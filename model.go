package gibberdet

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"unicode/utf8"
)

type Model struct {
	alpha         Alphabet
	ascii         *asciiAlphabet
	gram          []float64
	gibberScoreFn func(string) float64
	fast          bool
}

func NewModel(alpha Alphabet) *Model {
	// Assume we have seen 10 of each character pair.  This acts as a kind of
	// prior or smoothing factor.  This way, if we see a character transition
	// live that we've never observed in the past, we won't assume the entire
	// string has 0 probability.
	const weight = 10

	m := &Model{
		fast:  true,
		alpha: alpha,
		gram:  make([]float64, alpha.Len()*alpha.Len()),
	}
	for i := range m.gram {
		m.gram[i] = weight
	}
	m.init()
	return m
}

func (m *Model) init() {
	var ok bool
	if m.ascii, ok = m.alpha.(*asciiAlphabet); ok {
		m.gibberScoreFn = m.gibberScoreByByte
	} else {
		m.gibberScoreFn = m.gibberScoreByRune
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
	var maxBad float64
	for _, s := range goodInput {
		v := m.GibberScore(s)
		if v < minGood {
			minGood = v
		}
	}

	for _, s := range badInput {
		v := m.GibberScore(s)
		if v > maxBad {
			maxBad = v
		}
	}

	thresh = (minGood + maxBad) / 2
	if minGood <= maxBad {
		return thresh, fmt.Errorf("gibberdet: test failed; good threshold %f is less than bad %f", minGood, maxBad)
	}

	return thresh, nil
}

func (m *Model) Train(rdr io.Reader) error {
	scratch := make([]byte, 8192)

	var pos int
	var leftover []byte
	var first = true
	var last int

	alphaLen := m.alpha.Len()

	for {
	read:
		if len(leftover) > 0 {
			copy(scratch, leftover)
			pos = len(leftover)
		} else {
			pos = 0
		}

		n, err := rdr.Read(scratch[pos:])
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		end := n + pos

		for pos < end {
			r, sz := utf8.DecodeRune(scratch[pos:])
			if r == utf8.RuneError {
				if end-pos <= 4 {
					pos += sz
					leftover = scratch[pos:end]
					goto read
				} else {
					pos += sz
					continue
				}
			}

			alphaIdx := m.alpha.FindRune(r)
			pos += sz
			if alphaIdx >= 0 {
				if !first {
					m.gram[last*alphaLen+alphaIdx]++
				} else {
					first = false
				}
				last = alphaIdx

			} else if !first {
				first = true
			}
		}
	}

	// Normalize the counts so that they become log probabilities.
	// We use log probabilities rather than straight probabilities to avoid
	// numeric underflow issues with long texts.
	// This contains a justification:
	// http://squarecog.wordpress.com/2009/01/10/dealing-with-underflow-in-joint-probability-calculations/
	end := alphaLen * alphaLen
	for i := 0; i < end; i += alphaLen {
		var s float64
		for j := 0; j < alphaLen; j++ {
			s += m.gram[i+j]
		}
		for j := 0; j < alphaLen; j++ {
			m.gram[i+j] = math.Log(m.gram[i+j] / s)
		}
	}

	return nil
}

func (m *Model) GibberScore(s string) float64 {
	return m.gibberScoreFn(s)
}

func (m *Model) gibberScoreByByte(s string) float64 {
	if len(s) < 2 {
		return 0
	}

	// Return the average transition prob from l through log_prob_mat.
	var logProb float64
	var transitionCnt int

	var alphaA, alphaB int
	var alphaLen = m.ascii.Len()

	i := 0

first:
	alphaA = m.ascii.FindByte(s[i])
	if alphaA >= 0 {
		goto nextPair
	} else {
		goto nextFirst
	}

pair:
	alphaB = m.ascii.FindByte(s[i])
	if alphaB < 0 {
		goto nextFirst
	}
	logProb += m.gram[alphaA*alphaLen+alphaB]
	transitionCnt++
	alphaA = alphaB

nextPair:
	i++
	if i >= len(s) {
		goto done
	}
	goto pair

nextFirst:
	i++
	if i >= len(s) {
		goto done
	}
	goto first

done:
	if transitionCnt == 0 {
		return 0
	}

	// The exponentiation translates from log probs to probs.
	// return math.Exp(logProb / float64(transitionCnt))
	if m.fast {
		return expFast(logProb / float64(transitionCnt))
	}
	return math.Exp(logProb / float64(transitionCnt))
}

func (m *Model) gibberScoreByRune(s string) float64 {
	if len(s) < 2 {
		return 0
	}

	// Return the average transition prob from l through log_prob_mat.
	var logProb float64
	var transitionCnt float64

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
			// fmt.Printf("<" + string(lastRune) + string(r) + "> ")
			// fmt.Println(m.gram[last*m.alpha.ln+alphaIdx])
			logProb += m.gram[last*alphaLen+alphaIdx]
			transitionCnt += 1
		}
		last = alphaIdx
	}

	if transitionCnt == 0 {
		return 0
	}

	// The exponentiation translates from log probs to probs.
	if m.fast {
		return expFast(logProb / float64(transitionCnt))
	}
	return math.Exp(logProb / float64(transitionCnt))
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
