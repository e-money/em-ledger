// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package slashing

import (
	sdktypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	"github.com/e-money/em-ledger/x/slashing/keeper"
)

const (
	ModuleName   = sdktypes.ModuleName
	RouterKey    = sdktypes.RouterKey
	StoreKey     = sdktypes.StoreKey
	QuerierRoute = sdktypes.QuerierRoute
)

var (
	NewKeeper    = keeper.NewKeeper
	BeginBlocker = keeper.BeginBlocker
)

type (
	Keeper = keeper.Keeper
)
