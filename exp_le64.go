//+build arm amd64

package gibberdet

import "math"

// Based on constants taken from the 'approximate' library, which collects
// a stack of useful discovered techniques for fast approximate math with
// links to resources:
// https://github.com/ekmett/approximate/blob/master/cbits/fast.c
//
// For the purpose of this lib, the fast exp calculation seems to be more than
// adequate, and it's _significantly_ faster. (Turns out it's only about 30% faster
// on an RPi 4, sadly)
func expFast(a float64) float64 {
	var ux, vx uint64
	ux = uint64(3248660424278399*a + 0x3fdf127e83d16f12)
	vx = uint64(0x3fdf127e83d16f12 - 3248660424278399*a)
	return math.Float64frombits(ux) / math.Float64frombits(vx)
}
