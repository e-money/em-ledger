// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank/exported"
	params "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

type (
	GasPricesKeeper interface {
		SetMinimumGasPrices(gasPricesStr string) error
	}

	BankKeeper interface {
		GetSupply(ctx sdk.Context) exported.SupplyI
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
