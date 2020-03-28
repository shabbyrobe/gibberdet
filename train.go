package gibberdet

import (
	"fmt"
	"io"
)

func Train(alpha Alphabet, rdr ...io.Reader) (*Model, error) {
	if len(rdr) < 1 {
		return nil, fmt.Errorf("gibberdet: requires at least one reader")
	}
	tr := NewTrainer(alpha)
	for _, r := range rdr {
		if err := tr.Add(r); err != nil {
			return nil, err
		}
	}
	return tr.Compile()
}
