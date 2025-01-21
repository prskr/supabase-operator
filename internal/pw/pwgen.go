/*
Copyright 2025 Peter Kurfer.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package pw

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
