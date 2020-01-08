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

	// Query endpoints supported by the market querier
	QueryInstruments = "instruments"
	QueryInstrument  = "instrument"
)

var (
	// Parameter key for global order IDs
	globalOrderIDKey = []byte("globalOrderID")

	// LevelDB prefixes
	keysPrefix  = []byte{0x01}
	orderPrefix = []byte{0x02}
)

func GetOrderIDGeneratorKey() []byte {
	return append(keysPrefix, globalOrderIDKey...)
}

func GetOrderKey(v uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, v)
	return append(orderPrefix, b...)
}
