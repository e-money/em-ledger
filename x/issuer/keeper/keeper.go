package keeper

import (
	"emoney/x/issuer/types"
	lp "emoney/x/liquidityprovider"
	"fmt"
	"github.com/tendermint/tendermint/libs/log"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Keeper struct {
	lpKeeper lp.Keeper
}

func NewKeeper(lpk lp.Keeper) Keeper {
	return Keeper{
		lpKeeper: lpk,
	}
}

func (k Keeper) IncreaseCreditOfLiquidityProvider(ctx sdk.Context, liquidityProvider sdk.AccAddress, issuer sdk.AccAddress, creditIncrease sdk.Coins) {
	// TODO Verify that the issuer is allowed to modify LPs
	log := k.logger(ctx)

	lpAcc := k.lpKeeper.GetLiquidityProviderAccount(ctx, liquidityProvider)
	if lpAcc != nil {
		log.Info("Increasing liquidity provider credit", "account", liquidityProvider, "increase", creditIncrease)
		lpAcc.Credit.Add(creditIncrease)
	} else {
		// Account was not previously a liquidity provider
		k.lpKeeper.CreateLiquidityProvider(ctx, liquidityProvider, creditIncrease)
	}
}

func (k Keeper) logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
