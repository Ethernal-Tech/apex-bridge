package successtxprocessors

import (
	"slices"

	"github.com/Ethernal-Tech/cardano-infrastructure/sendtx"
)

// withNUL returns s with a NUL character (U+0000) as the whole value, before, after,
// inside and in place of one of its characters. Cardano metadata text only has to be
// valid UTF-8, which NUL is, so every variant reaches the oracle unchanged.
func withNUL(s string) map[string]string {
	mid := len(s) / 2

	return map[string]string{
		"only NUL":                  "\x00",
		"leading NUL":               "\x00" + s,
		"trailing NUL":              s + "\x00",
		"NUL inside":                s[:mid] + "\x00" + s[mid:],
		"NUL replacing a character": s[:mid] + "\x00" + s[mid+1:],
	}
}

// metadataAddrWithNUL returns addr split into metadata chunks with a NUL added: every
// withNUL variant, plus NUL as a chunk of its own before, between and after the chunks.
func metadataAddrWithNUL(addr string) map[string][]string {
	result := make(map[string][]string)

	for name, value := range withNUL(addr) {
		result[name] = sendtx.AddrToMetaDataAddr(value)
	}

	chunks := sendtx.AddrToMetaDataAddr(addr)
	result["NUL chunk first"] = append([]string{"\x00"}, chunks...)
	result["NUL chunk last"] = append(slices.Clone(chunks), "\x00")
	result["NUL chunk between"] = slices.Insert(slices.Clone(chunks), 1, "\x00")

	return result
}
