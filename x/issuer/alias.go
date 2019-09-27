package issuer

import (
	"emoney/x/issuer/keeper"
	"emoney/x/issuer/types"
)

const (
	StoreKey   = types.StoreKey
	ModuleName = types.ModuleName
)

var (
	ModuleCdc = types.ModuleCdc
	NewKeeper = keeper.NewKeeper
	NewIssuer = types.NewIssuer
)

type (
	Keeper = keeper.Keeper
	Issuer = types.Issuer
)
