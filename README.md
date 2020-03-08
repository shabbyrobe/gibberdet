gibberdet
=========

A Go port of the gibberish detection algorithm implemented here:
https://github.com/rrenaud/Gibberish-Detector.

The author originally proposed the technique in an answer on SO and it works pretty well
so it has quite a few ports:
http://stackoverflow.com/questions/6297991/is-there-any-way-to-detect-strings-like-putjbtghguhjjjanika/6298040#comment-7360747

This implementation supports alphabets of arbitrary size, with arbitrary runes,
with a significantly faster path for alphabets that are within the ASCII range.


Usage
-----

Choose or build an alphabet:

```go
alpha := gibberdet.ASCIIAlpha
alpha := gibberdet.NewAlphabet("abcde")
alpha, err := gibberdet.AlphabetFromReader(strings.NewReader(entireInput), nil)
```

Train the model:

```go
model := gibberdet.NewModel(alpha)
err := model.Train(strings.NewReader("lots and lots and lots of stuff"))
```

Build the test threshold with some good and bad strings:

```go
good := []string{"hello", "world"} // ... and lots more
bad := []string{"Ff7DPHaTaip", "9W5L9L30QG"} // ... and lots more
thresh, err := model.Test(good, bad)
```

Detect some gibberish. If the values aren't what you expect, use more
training and test data:

```go
model.GibberScore("hello") >= thresh // hopefully 'true'
model.GibberScore("aqwxGdRkdF6F0EoVQ") >= thresh // hopefully 'false'
```


Silly Benchmark Game
--------------------

Training and testing are unoptimised, but GibberScore should run pretty quickly. All
score calls have 0 allocs. For ASCII-only alphabets, GibberScore is quite a bit faster
with pure ASCII than with alphabets with runes >= 128. On my i7-8550U CPU @ 1.80GHz:

    BenchmarkASCII-8   	53777239	        22.2 ns/op	       0 B/op	       0 allocs/op
    BenchmarkRune-8   	14762448	        82.2 ns/op	       0 B/op	       0 allocs/op


How it works
------------

_From [rrenaud's](https://github.com/rrenaud/) original README_:

> The markov chain first 'trains' or 'studies' a few MB of English text,
> recording how often characters appear next to each other. Eg, given the text
> "Rob likes hacking" it sees Ro, ob, o[space], [space]l, ... It just counts
> these pairs. After it has finished reading through the training data, it
> normalizes the counts. Then each character has a probability distribution of 27
> followup character (26 letters + space) following the given initial.
> 
> So then given a string, it measures the probability of generating that string
> according to the summary by just multiplying out the probabilities of the
> adjacent pairs of characters in that string. EG, for that "Rob likes hacking"
> string, it would compute prob['r']['o'] * prob['o']['b'] * prob['b'][' '] ...
> This probability then measures the amount of 'surprise' assigned to this string
> according the data the model observed when training. If there is funny business
> with the input string, it will pass through some pairs with very low counts in
> the training phase, and hence have low probability/high surprise.
> 
> I then look at the amount of surprise per character for a few known good
> strings, and a few known bad strings, and pick a threshold between the most
> surprising good string and the least surprising bad string. Then I use that
> threshold whenever to classify any new piece of text.
