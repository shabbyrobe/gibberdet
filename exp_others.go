//+build -amd64,-arm

package gibberdet

import "math"

func expFast(f float64) float64 {
	return math.Exp(f)
}
