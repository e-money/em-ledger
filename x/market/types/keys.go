// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package types

const (
	ModuleName   = "market"
	StoreKey     = ModuleName
	RouterKey    = ModuleName
	QuerierRoute = ModuleName
)

var (
	// Parameter key for global order IDs
	GlobalOrderIDKey = []byte("globalOrderID")
)
