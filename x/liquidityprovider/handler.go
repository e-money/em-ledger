// This software is Copyright (c) 2019 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package liquidityprovider

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/e-money/em-ledger/x/liquidityprovider/keeper"
	"github.com/e-money/em-ledger/x/liquidityprovider/types"
)

// TODO Accept Keeper argument
func newHandler(k keeper.Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		switch msg := msg.(type) {
		case types.MsgMintTokens:
			return k.MintTokens(ctx, msg.LiquidityProvider, msg.Amount)
		case types.MsgBurnTokens:
			return k.BurnTokensFromBalance(ctx, msg.LiquidityProvider, msg.Amount)
		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized lp message type: %T", msg)
		}
	}
}
