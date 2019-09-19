package issuer

import (
	"emoney/x/issuer/keeper"
	"emoney/x/issuer/types"
)

var (
	ModuleCdc = types.ModuleCdc
	NewKeeper = keeper.NewKeeper
)

type (
	Keeper = keeper.Keeper
)
