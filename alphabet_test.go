package gibberdet

import "testing"

var BenchIntResult int

var miscChineseAlpha = NewAlphabet([]rune("的一是不了人我在有他这为之大来以个中上们到说国和地也子时道出而要于就下得可你年生自会那后能对着事其里所去行过家十用发天如然作方成者多日都三小军二无同么经法当起与好看学进种将还分此心前面又定见只主没公从"))

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
