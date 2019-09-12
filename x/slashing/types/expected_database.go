package types

import "github.com/tendermint/tendermint/libs/db"

// A reduced database interface which ensures that all modifications to state are written elsewhere.
type ReadOnlyDB interface {

	// Get returns nil iff key doesn't exist.
	// A nil key is interpreted as an empty byteslice.
	// CONTRACT: key, value readonly []byte
	Get([]byte) []byte

	// Has checks if a key exists.
	// A nil key is interpreted as an empty byteslice.
	// CONTRACT: key, value readonly []byte
	Has(key []byte) bool

	Iterator(start, end []byte) db.Iterator
}
