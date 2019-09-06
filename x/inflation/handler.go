package inflation

import (
	"emoney/x/inflation/internal/types"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func newHandler(keeper Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case types.MsgSetInflation:
			return handleMsgSetInflation(ctx, keeper, msg)
		default:
			errMsg := fmt.Sprintf("Unrecognized inflation Msg type: %v", msg.Type())
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

func handleMsgSetInflation(ctx sdk.Context, keeper Keeper, msg types.MsgSetInflation) sdk.Result {
	state := keeper.GetState(ctx)

	// TODO Add logging

	if !keeper.IsInflationAdministrator(ctx, msg.Principal) {
		return types.ErrUnauthorizedInflationChange(msg.Principal).Result()
	}

	asset := state.FindByDenom(msg.Denom)
	if asset == nil {
		errMsg := fmt.Sprintf("Unrecognized asset denomination: %v", msg.Denom)
		// TODO Convert to a local error defined in types/ package
		return sdk.ErrUnknownRequest(errMsg).Result()
	}

	asset.Inflation = msg.Inflation

	keeper.SetState(ctx, state)
	return sdk.Result{}
}
