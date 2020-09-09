package buyback

import (
	"fmt"

	"github.com/e-money/em-ledger/x/market/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func BeginBlocker(ctx sdk.Context, k Keeper) {
	if !k.UpdateBuybackMarket(ctx) {
		return
	}

	// For simplicity, all current orders are cancelled and replaced with new ones.
	k.CancelCurrentModuleOrders(ctx)

	var (
		stakingDenom = k.GetStakingTokenDenom(ctx)
		marketData   = k.GetMarketData(ctx)
		pricingInfo  = groupMarketDataBySource(marketData, stakingDenom)
		account      = k.GetBuybackAccount(ctx)
	)

	for _, balance := range account.GetCoins() {
		pricedata, found := pricingInfo[balance.Denom]
		if !found {
			// do not have market data to create an order.
			continue
		}

		if pricedata.LastPrice == nil {
			// do not have market data to create an order.
			continue
		}

		// Calculate the amount of staking tokens that can be purchased at that price
		tokenCount := balance.Amount.ToDec().Mul(*pricedata.LastPrice).RoundInt()
		if tokenCount.LT(sdk.OneInt()) {
			continue
		}

		order, err := types.NewOrder(
			types.TimeInForce_GoodTilCancel,
			balance,
			sdk.NewCoin(stakingDenom, tokenCount),
			account.GetAddress(),
			generateClientOrderId(ctx, balance),
		)

		if err != nil {
			ctx.Logger().Error("Error creating buyback order", "err", err)
			panic(err)
		}

		result, err := k.SendOrderToMarket(ctx, order)
		if err != nil {
			ctx.Logger().Error("Error sending buyback order to market", "err", err)
			panic(err)
		}

		ctx.EventManager().EmitEvents(result.Events)
	}

	err := k.BurnStakingToken(ctx)
	if err != nil {
		panic(err)
	}
}

func generateClientOrderId(ctx sdk.Context, balance sdk.Coin) string {
	return fmt.Sprintf("buyback-%v-%v", balance.Denom, ctx.BlockHeight())
}

// Return market data on trades that purchased the given denom.
func groupMarketDataBySource(marketData []types.MarketData, denom string) map[string]types.MarketData {
	result := make(map[string]types.MarketData)

	for _, md := range marketData {
		if md.Destination != denom {
			continue
		}

		result[md.Source] = md
	}

	return result
}
