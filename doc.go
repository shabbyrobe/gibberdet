/*
Package gibberdet is a Go port of the gibberish detection algorithm implemented here:
https://github.com/rrenaud/Gibberish-Detector.

The author originally proposed the technique in an answer on SO and it works pretty well
so it has quite a few ports:
http://stackoverflow.com/questions/6297991/is-there-any-way-to-detect-strings-like-putjbtghguhjjjanika/6298040#comment-7360747

This implementation supports alphabets of arbitrary size, with arbitrary runes,
with a significantly faster path for alphabets that are within the ASCII range.


Usage

Choose or build an alphabet:

	alpha := gibberdet.ASCIIAlpha
	alpha := gibberdet.NewAlphabet("abcde")
	alpha, err := gibberdet.AlphabetFromReader(strings.NewReader(entireInput), nil)

Train the model. Use _lots_ of data:

	trainer := gibberdet.NewTrainer(alpha)
	err := trainer.Add(strings.NewReader("lots and lots and lots of stuff"))
	err := trainer.Add(strings.NewReader("even more stuff"))
	model, err := trainer.Compile()

Or use one of the existing built models in `testdata/`, at the moment the
[OANC](http://www.anc.org/data/oanc/download/) one is probably the best one in
there.

Save/load the model:

	bts, err := model.MarshalBinary()
	var load gibberdet.Model
	err := load.UnmarshalBinary(bts)

Build the test threshold with some good and bad strings:

	good := []string{"hello", "world"} // ... and lots more
	bad := []string{"Ff7DPHaTaip", "9W5L9L30QG"} // ... and lots more
	thresh, err := model.Test(good, bad)

Detect some gibberish. If the values aren't what you expect, use more
training and test data:

	model.GibberScore("hello") >= thresh // hopefully 'true'
	model.GibberScore("aqwxGdRkdF6F0EoVQ") >= thresh // hopefully 'false'

*/
package gibberdet
