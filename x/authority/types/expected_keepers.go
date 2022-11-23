package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	params "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

type (
	GasPricesKeeper interface {
		SetMinimumGasPrices(gasPricesStr string) error
	}

	BankKeeper interface {
		GetPaginatedTotalSupply(ctx sdk.Context, pagination *query.PageRequest) (sdk.Coins, *query.PageResponse, error)
	}

	UpgradeKeeper interface {
		ApplyUpgrade(ctx sdk.Context, plan types.Plan)
		GetUpgradePlan(ctx sdk.Context) (plan types.Plan, havePlan bool)
		HasHandler(name string) bool
		ScheduleUpgrade(ctx sdk.Context, plan types.Plan) error
		SetUpgradeHandler(name string, upgradeHandler types.UpgradeHandler)
	}

	ParamsKeeper interface {
		GetSubspace(name string) (ss params.Subspace, found bool)
	}
)
