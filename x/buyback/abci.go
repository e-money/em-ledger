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
		account      = k.GetBuybackAccountAddr()
	)

	for _, balance := range bk.GetAllBalances(ctx, account) {
		if balance.Denom == stakingDenom {
			continue
		}

		price := k.GetBestPrice(ctx, balance.Denom, stakingDenom)
		if price == nil {
			// There are no passive orders to fill for this instrument
			continue
		}

		// Calculate the amount of staking tokens that can be purchased at that price
		destinationAmount := balance.Amount.ToDec().Mul(*price).TruncateInt()
		if destinationAmount.LT(sdk.OneInt()) {
			continue
		}

		order, err := markettypes.NewOrder(
			ctx.BlockTime(),
			markettypes.TimeInForce_GoodTillCancel,
			balance,
			sdk.NewCoin(stakingDenom, destinationAmount),
			account,
			generateClientOrderId(ctx, balance),
		)

		if err != nil {
			ctx.Logger().Error("Error creating buyback order", "err", err)
			continue
		}

		if err := k.SendOrderToMarket(ctx, order); err != nil {
			ctx.Logger().Error("Error sending buyback order to market", "err", err)
			continue
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
