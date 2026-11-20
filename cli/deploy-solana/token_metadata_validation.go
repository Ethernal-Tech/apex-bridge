package clideploysolana

import "fmt"

// Metaplex Token Metadata limits, in bytes (mpl-token-metadata MAX_*_LENGTH).
// The program rejects longer values with NameTooLong, SymbolTooLong or UriTooLong;
// checking them here fails fast, before a transaction is built and paid for.
const (
	metaplexMaxNameLength   = 32
	metaplexMaxSymbolLength = 10
	metaplexMaxURILength    = 200
)

// validateTokenMetadata checks name, symbol and uri against the Metaplex limits.
// The flag names are used in the error messages.
func validateTokenMetadata(name, nameFlag, symbol, symbolFlag, uri, uriFlag string) error {
	if name == "" {
		return fmt.Errorf("token name not specified: --%s", nameFlag)
	}

	if len(name) > metaplexMaxNameLength {
		return fmt.Errorf("token name is %d bytes, at most %d allowed: --%s",
			len(name), metaplexMaxNameLength, nameFlag)
	}

	if symbol == "" {
		return fmt.Errorf("token symbol not specified: --%s", symbolFlag)
	}

	if len(symbol) > metaplexMaxSymbolLength {
		return fmt.Errorf("token symbol is %d bytes, at most %d allowed: --%s",
			len(symbol), metaplexMaxSymbolLength, symbolFlag)
	}

	if len(uri) > metaplexMaxURILength {
		return fmt.Errorf("token URI is %d bytes, at most %d allowed: --%s",
			len(uri), metaplexMaxURILength, uriFlag)
	}

	return nil
}
