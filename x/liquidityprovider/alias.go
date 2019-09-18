package liquidityprovider

import (
	"emoney/x/liquidityprovider/keeper"
	"emoney/x/liquidityprovider/types"
)

const (
	ModuleName = types.ModuleName
)

var (
	ModuleCdc = types.ModuleCdc
	NewKeeper = keeper.NewKeeper
)

type (
	Keeper = keeper.Keeper
)
