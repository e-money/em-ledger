// This software is Copyright (c) 2019 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package types

import db "github.com/tendermint/tm-db"

// A reduced database interface which ensures that all modifications to state are written elsewhere.
type ReadOnlyDB interface {

	// Get returns nil iff key doesn't exist.
	// A nil key is interpreted as an empty byteslice.
	// CONTRACT: key, value readonly []byte
	Get([]byte) ([]byte, error)

	// Has checks if a key exists.
	// A nil key is interpreted as an empty byteslice.
	// CONTRACT: key, value readonly []byte
	Has(key []byte) (bool, error)

	Iterator(start, end []byte) (db.Iterator, error)
}
