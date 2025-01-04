package controller

import (
	"bytes"
	"math/rand/v2"
)

func GeneratePW(length uint, random *rand.Rand) []byte {
	var (
		builder  = bytes.NewBuffer(nil)
		alphabet = runes('a', 'z') + runes('A', 'Z') + runes('0', '9')
	)

	if random == nil {
		random = rand.New(rand.NewPCG(0, 0))
	}

	for range length {
		builder.WriteRune(rune(alphabet[random.IntN(len(alphabet))]))
	}

	return builder.Bytes()
}

func runes(start, end rune) string {
	result := make([]rune, 0, int(end-start))

	for current := start; current != end; current++ {
		result = append(result, current)
	}

	return string(result)
}
