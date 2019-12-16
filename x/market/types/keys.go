// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package types

import "encoding/binary"

const (
	ModuleName   = "market"
	StoreKey     = ModuleName
	RouterKey    = ModuleName
	QuerierRoute = ModuleName
)

var (
	// Parameter key for global order IDs
	GlobalOrderIDKey = []byte("globalOrderID")

	// LevelDB prefixes
	orderPrefix = []byte{0x01}
)

func GetOrderKey(v uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, v)
	return append(orderPrefix, b...)
}
