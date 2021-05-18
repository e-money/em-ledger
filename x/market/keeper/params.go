package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/e-money/em-ledger/x/market/types"
)

// InitParamsStore writes the params to the module store
func (k Keeper) InitParamsStore(ctx sdk.Context) uint64 {
	defaultSet := types.DefaultTxParams()
	var trxFee uint64
	if !k.paramStore.Has(ctx, types.KeyTrxFee) {
		k.paramStore.Set(ctx, types.KeyTrxFee, defaultSet.TrxFee)
	}
	if !k.paramStore.Has(ctx, types.KeyLiquidTrxFee) {
		k.paramStore.Set(ctx, types.KeyLiquidTrxFee, defaultSet.LiquidTrxFee)
	}
	if !k.paramStore.Has(ctx, types.KeyLiquidityRebateMinutesSpan) {
		k.paramStore.Set(
			ctx, types.KeyLiquidityRebateMinutesSpan,
			defaultSet.LiquidityRebateMinutesSpan,
		)
	}

	return trxFee
}

// GetTrxFee retrieves the trx fee from the paramStore
func (k Keeper) GetTrxFee(ctx sdk.Context) uint64 {
	var trxFee uint64
	k.paramStore.Get(ctx, types.KeyTrxFee, &trxFee)

	return trxFee
}

// GetLiquidTrxFee retrieves the liquid trx fee from the paramStore
func (k Keeper) GetLiquidTrxFee(ctx sdk.Context) uint64 {
	var liqTrxFee uint64
	k.paramStore.Get(ctx, types.KeyLiquidTrxFee, &liqTrxFee)

	return liqTrxFee
}

// GetLiquidityRebateMinutesSpan retrieves the Minutes interval for eligible
// replacing transactions to receive a rebate. For a rebate to apply, the
// replacing transaction should occur these minutes after the signer's original
// trx.
func (k Keeper) GetLiquidityRebateMinutesSpan(ctx sdk.Context) int64 {
	var liqTrxFee int64
	k.paramStore.Get(ctx, types.KeyLiquidityRebateMinutesSpan, &liqTrxFee)

	return liqTrxFee
}

// GetParams returns the total market parameters set.
func (k Keeper) GetParams(ctx sdk.Context) types.TxParams {
	return types.NewTxParams(k.GetTrxFee(ctx), k.GetLiquidTrxFee(ctx),
		k.GetLiquidityRebateMinutesSpan(ctx))
}

// SetParams sets the total market parameters set.
func (k Keeper) SetParams(ctx sdk.Context, params types.TxParams) {
	k.paramStore.SetParamSet(ctx, &params)
}

