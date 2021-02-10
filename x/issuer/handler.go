// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package issuer

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/e-money/em-ledger/x/issuer/keeper"
	"github.com/e-money/em-ledger/x/issuer/types"
)

// TODO Accept Keeper argument
func newHandler(k keeper.Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		switch msg := msg.(type) {
		case *types.MsgIncreaseMintable:
			return handleMsgIncreaseMintableAmount(ctx, msg, k)
		case *types.MsgDecreaseMintable:
			return handleMsgDecreaseMintableAmount(ctx, msg, k)
		case *types.MsgRevokeLiquidityProvider:
			return handleMsgRevokeLiquidityProvider(ctx, msg, k)
		case *types.MsgSetInflation:
			return handleMsgSetInflation(ctx, msg, k)
		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "Unrecognized issuance Msg type: %T", msg)
		}
	}
}

func handleMsgSetInflation(ctx sdk.Context, msg *types.MsgSetInflation, k keeper.Keeper) (*sdk.Result, error) {
	issuer, err := sdk.AccAddressFromBech32(msg.Issuer)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "issuer")
	}
	return k.SetInflationRate(ctx, issuer, msg.InflationRate, msg.Denom)
}

func handleMsgRevokeLiquidityProvider(ctx sdk.Context, msg *types.MsgRevokeLiquidityProvider, k keeper.Keeper) (*sdk.Result, error) {
	issuer, err := sdk.AccAddressFromBech32(msg.Issuer)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "issuer")
	}
	liquidityProvider, err := sdk.AccAddressFromBech32(msg.LiquidityProvider)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "liquidity provider")
	}
	return k.RevokeLiquidityProvider(ctx, liquidityProvider, issuer)
}

func handleMsgDecreaseMintableAmount(ctx sdk.Context, msg *types.MsgDecreaseMintable, k keeper.Keeper) (*sdk.Result, error) {
	issuer, err := sdk.AccAddressFromBech32(msg.Issuer)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "issuer")
	}
	liquidityProvider, err := sdk.AccAddressFromBech32(msg.LiquidityProvider)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "liquidity provider")
	}

	return k.DecreaseMintableAmountOfLiquidityProvider(ctx, liquidityProvider, issuer, msg.MintableDecrease)
}

func handleMsgIncreaseMintableAmount(ctx sdk.Context, msg *types.MsgIncreaseMintable, k keeper.Keeper) (*sdk.Result, error) {
	issuer, err := sdk.AccAddressFromBech32(msg.Issuer)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "issuer")
	}
	liquidityProvider, err := sdk.AccAddressFromBech32(msg.LiquidityProvider)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "liquidity provider")
	}

	return k.IncreaseMintableAmountOfLiquidityProvider(ctx, liquidityProvider, issuer, msg.MintableIncrease)
}
