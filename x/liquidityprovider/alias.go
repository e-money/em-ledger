package liquidityprovider

import (
	"github.com/e-money/em-ledger/x/liquidityprovider/keeper"
	"github.com/e-money/em-ledger/x/liquidityprovider/types"
)

const (
	ModuleName = types.ModuleName
)

var (
	ModuleCdc     = types.ModuleCdc
	RegisterCodec = types.RegisterCodec
	NewKeeper     = keeper.NewKeeper
)

type (
	Keeper  = keeper.Keeper
	Account = types.LiquidityProviderAccount
)
