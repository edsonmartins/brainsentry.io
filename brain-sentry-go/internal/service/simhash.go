package service

import (
	"hash/fnv"
	"strings"
	"unicode"
)

// ComputeSimHash computes a 64-bit SimHash fingerprint for text content.
// Similar texts will have fingerprints with small Hamming distance.
func ComputeSimHash(text string) uint64 {
	tokens := tokenize(text)
	if len(tokens) == 0 {
		return 0
	}

	var v [64]int
	for _, token := range tokens {
		h := hashToken(token)
		for i := 0; i < 64; i++ {
			if (h>>uint(i))&1 == 1 {
				v[i]++
			} else {
				v[i]--
			}
		}
	}

	var fingerprint uint64
	for i := 0; i < 64; i++ {
		if v[i] > 0 {
			fingerprint |= 1 << uint(i)
		}
	}
	return fingerprint
}

// SimHashHammingDistance returns the Hamming distance between two SimHash values.
func SimHashHammingDistance(a, b uint64) int {
	xor := a ^ b
	count := 0
	for xor != 0 {
		count++
		xor &= xor - 1 // clear lowest set bit
	}
	return count
}

// SimHashToHex converts a SimHash uint64 to a 16-character hex string.
func SimHashToHex(h uint64) string {
	const hex = "0123456789abcdef"
	var buf [16]byte
	for i := 15; i >= 0; i-- {
		buf[i] = hex[h&0xf]
		h >>= 4
	}
	return string(buf[:])
}

// SimHashFromHex converts a 16-character hex string back to uint64.
func SimHashFromHex(s string) uint64 {
	var h uint64
	for _, c := range s {
		h <<= 4
		switch {
		case c >= '0' && c <= '9':
			h |= uint64(c - '0')
		case c >= 'a' && c <= 'f':
			h |= uint64(c - 'a' + 10)
		}
	}
	return h
}

func tokenize(text string) []string {
	lower := strings.ToLower(text)
	var tokens []string
	var current strings.Builder
	for _, r := range lower {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			current.WriteRune(r)
		} else {
			if current.Len() > 2 { // skip very short tokens
				tokens = append(tokens, current.String())
			}
			current.Reset()
		}
	}
	if current.Len() > 2 {
		tokens = append(tokens, current.String())
	}
	return tokens
}

func hashToken(token string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(token))
	return h.Sum64()
}
