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

	EventTypeBuybackBurn   = types.EventTypeBuybackBurn
	AttributeKeyBurnAmount = types.AttributeKeyBurnAmount
)

type (
	Keeper               = keeper.Keeper
	StakingKeeper        = keeper.StakingKeeper
	QueryBalanceResponse = keeper.QueryBalanceResponse
)

var (
	NewKeeper = keeper.NewKeeper
)
