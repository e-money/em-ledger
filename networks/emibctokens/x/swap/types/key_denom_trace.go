package types

import "encoding/binary"

var _ binary.ByteOrder

const (
	// DenomTraceKeyPrefix is the prefix to retrieve all DenomTrace
	DenomTraceKeyPrefix = "DenomTrace/value/"
)

// DenomTraceKey returns the store key to retrieve a DenomTrace from the index fields
func DenomTraceKey(
	index string,
) []byte {
	var key []byte

	indexBytes := []byte(index)
	key = append(key, indexBytes...)
	key = append(key, []byte("/")...)

	return key
}
