package buyback

import (
	"github.com/e-money/em-ledger/x/buyback/internal/keeper"
	"github.com/e-money/em-ledger/x/buyback/internal/types"
)

const (
	ModuleName   = types.ModuleName
	QuerierRoute = types.QuerierRoute
	AccountName  = types.AccountName
	StoreKey     = types.StoreKey
	QueryBalance = types.QueryBalance

	EventTypeBuyback   = types.EventTypeBuyback
	AttributeKeyAction = types.AttributeKeyAction
	AttributeKeyAmount = types.AttributeKeyAmount
)

type (
	Keeper               = keeper.Keeper
	StakingKeeper        = keeper.StakingKeeper
	QueryBalanceResponse = types.QueryBalanceResponse
	GenesisState         = types.GenesisState
)

var (
	NewKeeper = keeper.NewKeeper
)
