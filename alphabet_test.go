package gibberdet

import (
	"strings"
	"testing"
)

var BenchIntResult int

var miscChineseAlpha = NewAlphabet([]rune("天地玄黃宇宙洪荒日月盈昃辰宿列張寒來暑往秋收冬藏閏餘成歲律召調陽雲騰致雨露結爲霜金生麗水玉出崑岡劍號巨闕珠稱夜光果珍李柰菜重芥薑海鹹河淡鱗潛羽翔龍師火帝鳥官人皇始制文字乃服衣裳推位讓國有虞陶唐"))

func TestAlphabetASCIIFromReader(t *testing.T) {
	chars := "abcdefg"
	rdr := strings.NewReader(chars)
	al, err := AlphabetFromReader(rdr, nil)
	if err != nil {
		t.Fatal(err)
	}

	var expected [256]bool
	for _, c := range chars {
		expected[c] = true
	}

	asc := al.(*asciiAlphabet)
	for i := 0; i < 256; i++ {
		if asc.FindByte(byte(i)) >= 0 != expected[i] {
			t.Fatal(byte(i))
		}
	}
}

func TestAlphabetRuneFromReader(t *testing.T) {
	chars := "天地玄黃"
	rdr := strings.NewReader(chars)
	al, err := AlphabetFromReader(rdr, nil)
	if err != nil {
		t.Fatal(err)
	}

	rn := al.(*runeAlphabet)
	for _, c := range chars {
		if rn.FindRune(c) < 0 {
			t.Fatal(c)
		}
	}

	for i := 0; i < 256; i++ {
		if rn.FindRune(rune(i)) >= 0 {
			t.Fatal(rune(i))
		}
	}
}

func BenchmarkAlphabetFindRuneASCIIInterface(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		BenchIntResult = ASCIIAlpha.FindRune('a')
	}
}

func BenchmarkAlphabetFindByteASCIIInterface(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		BenchIntResult = ASCIIAlpha.FindByte('a')
	}
}

func BenchmarkAlphabetFindByteASCIIDirect(b *testing.B) {
	b.ReportAllocs()
	alpha := newASCIIAlphabet([]rune(alpha))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		BenchIntResult = alpha.FindByte('a')
	}
}

func BenchmarkAlphabetFindRuneWide(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		BenchIntResult = miscChineseAlpha.FindRune('道')
	}
}
