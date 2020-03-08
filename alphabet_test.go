package gibberdet

import "testing"

var BenchIntResult int

var miscChineseAlpha = NewAlphabet([]rune("天地玄黃宇宙洪荒日月盈昃辰宿列張寒來暑往秋收冬藏閏餘成歲律召調陽雲騰致雨露結爲霜金生麗水玉出崑岡劍號巨闕珠稱夜光果珍李柰菜重芥薑海鹹河淡鱗潛羽翔龍師火帝鳥官人皇始制文字乃服衣裳推位讓國有虞陶唐"))

func BenchmarkAlphabetFindRuneASCIIInterface(b *testing.B) {
	for i := 0; i < b.N; i++ {
		BenchIntResult = ASCIIAlpha.FindRune('a')
	}
}

func BenchmarkAlphabetFindByteASCIIInterface(b *testing.B) {
	for i := 0; i < b.N; i++ {
		BenchIntResult = ASCIIAlpha.FindByte('a')
	}
}

func BenchmarkAlphabetFindByteASCIIDirect(b *testing.B) {
	alpha := newASCIIAlphabet([]rune(alpha))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		BenchIntResult = alpha.FindByte('a')
	}
}

func BenchmarkAlphabetFindRuneWide(b *testing.B) {
	for i := 0; i < b.N; i++ {
		BenchIntResult = miscChineseAlpha.FindRune('道')
	}
}
