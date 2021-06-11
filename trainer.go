package gibberdet

import (
	"fmt"
	"io"
	"math"
	"unicode/utf8"
)

// Assume we have seen 10 of each character pair. This acts as a kind of
// prior or smoothing factor. This way, if we see a character transition
// live that we've never observed in the past, we won't assume the entire
// string has 0 probability.
const DefaultPairWeight = 10

type Trainer struct {
	alpha      Alphabet
	ascii      *asciiAlphabet
	gram       []float64
	scratch    []byte
	pairWeight float64
}

type TrainerOption func(t *Trainer)

func TrainerPairWeight(w float64) TrainerOption {
	return func(t *Trainer) {
		t.pairWeight = w
	}
}

func NewTrainer(alpha Alphabet, opts ...TrainerOption) *Trainer {
	scratch := make([]byte, 8192)

	t := &Trainer{
		alpha:      alpha,
		gram:       make([]float64, alpha.Len()*alpha.Len()),
		scratch:    scratch,
		pairWeight: DefaultPairWeight,
	}

	for _, o := range opts {
		o(t)
	}

	for i := range t.gram {
		t.gram[i] = t.pairWeight
	}

	return t
}

func (t *Trainer) Add(rdr io.Reader) error {
	var pos int
	var leftover []byte
	var first = true
	var last int

	alphaLen := t.alpha.Len()

	for {
	read:
		if len(leftover) > 0 {
			pos = copy(t.scratch, leftover)
		} else {
			pos = 0
		}

		n, err := rdr.Read(t.scratch[pos:])
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		end := n + pos

		for pos < end {
			r, sz := utf8.DecodeRune(t.scratch[pos:])
			if r == utf8.RuneError {
				if end-pos <= 4 {
					pos += sz
					leftover = t.scratch[pos:end]
					goto read
				} else {
					pos += sz
					continue
				}
			}

			alphaIdx := t.alpha.FindRune(r)
			pos += sz
			if alphaIdx >= 0 {
				if !first {
					t.gram[last*alphaLen+alphaIdx]++
				} else {
					first = false
				}
				last = alphaIdx

			} else if !first {
				first = true
			}
		}
	}

	return nil
}

func (t *Trainer) Compile() (*Model, error) {
	alphaLen := t.alpha.Len()

	gram := make([]float64, len(t.gram))
	copy(gram, t.gram)

	m := &Model{
		alpha: t.alpha,
		gram:  gram,
	}
	m.init()

	// Normalize the counts so that they become log probabilities.
	// We use log probabilities rather than straight probabilities to avoid
	// numeric underflow issues with long texts.
	// This contains a justification:
	// http://squarecog.wordpress.com/2009/01/10/dealing-with-underflow-in-joint-probability-calculations/
	end := alphaLen * alphaLen
	for i := 0; i < end; i += alphaLen {
		var s float64
		for j := 0; j < alphaLen; j++ {
			s += gram[i+j]
		}

		for j := 0; j < alphaLen; j++ {
			gram[i+j] = math.Log(gram[i+j] / s)

			if math.IsNaN(gram[i+j]) {
				return nil, fmt.Errorf("NaN detected for %q, %q", string(m.alpha.Runes()[i/alphaLen]), string(m.alpha.Runes()[j]))
			}
			if math.IsInf(gram[i+j], 0) {
				gram[i+j] = math.SmallestNonzeroFloat64
				// return nil, fmt.Errorf("Inf detected for %q, %q", string(m.alpha.Runes()[i/alphaLen]), string(m.alpha.Runes()[j]))
			}
		}
	}

	return m, nil
}
