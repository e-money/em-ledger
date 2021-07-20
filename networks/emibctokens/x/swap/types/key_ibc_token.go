package types

import "encoding/binary"

var _ binary.ByteOrder

const (
	// IbcTokenKeyPrefix is the prefix to retrieve all IbcToken
	IbcTokenKeyPrefix = "IbcToken/value/"
)

// IbcTokenKey returns the store key to retrieve a IbcToken from the index fields
func IbcTokenKey(
	index string,
) []byte {
	var key []byte

	indexBytes := []byte(index)
	key = append(key, indexBytes...)
	key = append(key, []byte("/")...)

	return key
}
