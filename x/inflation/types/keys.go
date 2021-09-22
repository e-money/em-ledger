// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package types

// the one key to use for the keeper store
var MinterKey = []byte{0x00}

// nolint
const (
	// module name
	ModuleName = "inflation"

	// default paramspace for params keeper
	DefaultParamspace = ModuleName

	// StoreKey is the default store key for mint
	StoreKey = ModuleName

	// QuerierRoute is the querier route for the inflation store.
	QuerierRoute = StoreKey

	// Query endpoints supported by the inflation querier
	QueryInflation = ModuleName
)
