package buyback

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/e-money/em-ledger/x/buyback/internal/types"
	markettypes "github.com/e-money/em-ledger/x/market/types"
)

func BeginBlocker(ctx sdk.Context, k Keeper, bk types.BankKeeper) {
	if !k.UpdateBuybackMarket(ctx) {
		return
	}

	// For simplicity, all current orders are cancelled and replaced with new ones.
	k.CancelCurrentModuleOrders(ctx)

	var (
		stakingDenom = k.GetStakingTokenDenom(ctx)
		pricingInfo  = groupMarketDataBySource(k.GetMarketData(ctx), stakingDenom)
		account      = k.GetBuybackAccountAddr()
	)

	for _, balance := range bk.GetAllBalances(ctx, account) {
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
		destinationAmount := balance.Amount.ToDec().Mul(*pricedata.LastPrice).TruncateInt()
		if destinationAmount.LT(sdk.OneInt()) {
			continue
		}

		order, err := markettypes.NewOrder(
			markettypes.TimeInForce_GoodTillCancel,
			balance,
			sdk.NewCoin(stakingDenom, destinationAmount),
			account,
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
		for _, ev := range result.Events {
			ctx.EventManager().EmitEvent(sdk.Event(ev))
		}
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
func groupMarketDataBySource(marketData []markettypes.MarketData, denom string) map[string]markettypes.MarketData {
	result := make(map[string]markettypes.MarketData)

	for _, md := range marketData {
		if md.Destination != denom {
			continue
		}

		result[md.Source] = md
	}

	return result
}
