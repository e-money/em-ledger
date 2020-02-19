// This software is Copyright (c) 2019 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package authority

import (
	"github.com/e-money/em-ledger/x/authority/client/cli"
	"github.com/e-money/em-ledger/x/authority/keeper"
	"github.com/e-money/em-ledger/x/authority/types"
)

const (
	ModuleName     = types.ModuleName
	StoreKey       = types.StoreKey
	QuerierRoute   = types.QuerierRoute
	QueryGasPrices = types.QueryGasPrices
)

type (
	Keeper = keeper.Keeper

	QueryGasPricesResponse = keeper.QueryGasPricesResponse
)

var (
	ModuleCdc       = types.ModuleCdc
	RegisterCodec   = types.RegisterCodec
	NewKeeper       = keeper.NewKeeper
	BeginBlocker    = keeper.BeginBlocker
	GetGasPricesCmd = cli.GetGasPricesCmd
	GetQueryCmd     = cli.GetQueryCmd
	GetTxCmd        = cli.GetTxCmd
)
