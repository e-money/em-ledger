// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package liquidityprovider

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/e-money/em-ledger/x/liquidityprovider/keeper"
	"github.com/e-money/em-ledger/x/liquidityprovider/types"
)

func newHandler(k keeper.Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		switch msg := msg.(type) {
		case *types.MsgMintTokens:
			liquidityProvider, err := sdk.AccAddressFromBech32(msg.LiquidityProvider)
			if err != nil {
				return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "liquidity provider")
			}
			return k.MintTokens(ctx, liquidityProvider, msg.Amount)
		case *types.MsgBurnTokens:
			liquidityProvider, err := sdk.AccAddressFromBech32(msg.LiquidityProvider)
			if err != nil {
				return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "liquidity provider")
			}
			return k.BurnTokensFromBalance(ctx, liquidityProvider, msg.Amount)
		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized lp message type: %T", msg)
		}
	}
}
