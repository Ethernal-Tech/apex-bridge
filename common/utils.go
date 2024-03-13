package common

import (
	"encoding/hex"
	"net/url"
	"strings"

	"github.com/ethereum/go-ethereum/common"
)

func IsValidURL(input string) bool {
	_, err := url.ParseRequestURI(input)
	return err == nil
}

func HexToAddress(s string) common.Address {
	return common.HexToAddress(s)
}

func DecodeHex(s string) ([]byte, error) {
	if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X") {
		s = s[2:]
	}

	return hex.DecodeString(s)
}
