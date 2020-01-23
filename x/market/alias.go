// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package market

import (
	"github.com/e-money/em-ledger/x/market/keeper"
	"github.com/e-money/em-ledger/x/market/types"
)

const (
	ModuleName   = types.ModuleName
	RouterKey    = types.RouterKey
	StoreKey     = types.StoreKey
	QuerierRoute = types.QuerierRoute
)

var (
	NewKeeper    = keeper.NewKeeper
	BeginBlocker = keeper.BeginBlocker
)

type (
	Keeper = keeper.Keeper
)
