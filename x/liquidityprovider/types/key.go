// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package types

const (
	ModuleName   = "liquidityprovider"
	QuerierRoute = ModuleName
	RouterKey    = ModuleName
	StoreKey   = ModuleName
)

// IAVL Store prefixes
var (
	ProviderKeyPrefix   = []byte{0x00}
	MintableKeyPrefix   = []byte{0x01}
)