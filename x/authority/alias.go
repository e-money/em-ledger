package authority

import (
	"emoney/x/authority/keeper"
	"emoney/x/authority/types"
)

const (
	ModuleName   = types.ModuleName
	StoreKey     = types.StoreKey
	QuerierRoute = types.QuerierRoute
)

type (
	Keeper = keeper.Keeper
)

var (
	ModuleCdc     = types.ModuleCdc
	RegisterCodec = types.RegisterCodec
)
