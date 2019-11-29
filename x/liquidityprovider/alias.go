// This software is Copyright (c) 2019 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

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
