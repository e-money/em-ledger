package types

import sdkslashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"

const (
	// module name
	ModuleName = sdkslashingtypes.ModuleName

	// StoreKey is the store key string for slashing
	StoreKey = sdkslashingtypes.ModuleName

	// RouterKey is the message route for slashing
	RouterKey = sdkslashingtypes.ModuleName

	// QuerierRoute is the querier route for slashing
	QuerierRoute = sdkslashingtypes.ModuleName

	// The module account holding the slashing penalties until they are paid out to the remaining validators.
	PenaltyAccount = "slashing_penalties"
)
