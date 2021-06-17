package keeper

import (
	"context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/e-money/em-ledger/x/issuer/types"
)

var _ types.MsgServer = msgServer{}

type issuerKeeper interface {
	IncreaseMintableAmountOfLiquidityProvider(ctx sdk.Context, liquidityProvider string, issuer sdk.AccAddress, mintableIncrease sdk.Coins) (*sdk.Result, error)
	DecreaseMintableAmountOfLiquidityProvider(ctx sdk.Context, liquidityProvider string, issuer sdk.AccAddress, mintableDecrease sdk.Coins) (*sdk.Result, error)
	RevokeLiquidityProvider(ctx sdk.Context, liquidityProvider string, issuerAddress sdk.AccAddress) (*sdk.Result, error)
	SetInflationRate(ctx sdk.Context, issuer sdk.AccAddress, inflationRate sdk.Dec, denom string) (*sdk.Result, error)
}

type msgServer struct {
	k issuerKeeper
}

func NewMsgServerImpl(keeper issuerKeeper) types.MsgServer {
	return &msgServer{k: keeper}
}

func (m msgServer) IncreaseMintable(c context.Context, msg *types.MsgIncreaseMintable) (*types.MsgIncreaseMintableResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	issuer, err := sdk.AccAddressFromBech32(msg.Issuer)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "issuer")
	}

	_, err = sdk.AccAddressFromBech32(msg.LiquidityProvider)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "liquidity provider:"+msg.LiquidityProvider)
	}

	result, err := m.k.IncreaseMintableAmountOfLiquidityProvider(ctx, msg.LiquidityProvider, issuer, msg.MintableIncrease)
	if err != nil {
		return nil, err
	}

	for _, e := range result.Events {
		ctx.EventManager().EmitEvent(sdk.Event(e))
	}
	return &types.MsgIncreaseMintableResponse{}, nil
}

func (m msgServer) DecreaseMintable(c context.Context, msg *types.MsgDecreaseMintable) (*types.MsgDecreaseMintableResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	issuer, err := sdk.AccAddressFromBech32(msg.Issuer)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "issuer")
	}

	_, err = sdk.AccAddressFromBech32(msg.LiquidityProvider)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "liquidity provider:"+msg.LiquidityProvider)
	}

	result, err := m.k.DecreaseMintableAmountOfLiquidityProvider(ctx, msg.LiquidityProvider, issuer, msg.MintableDecrease)
	if err != nil {
		return nil, err
	}

	for _, e := range result.Events {
		ctx.EventManager().EmitEvent(sdk.Event(e))
	}
	return &types.MsgDecreaseMintableResponse{}, nil
}

func (m msgServer) RevokeLiquidityProvider(c context.Context, msg *types.MsgRevokeLiquidityProvider) (*types.MsgRevokeLiquidityProviderResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	issuer, err := sdk.AccAddressFromBech32(msg.Issuer)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "issuer")
	}
	result, err := m.k.RevokeLiquidityProvider(ctx, msg.LiquidityProvider, issuer)
	if err != nil {
		return nil, err
	}
	for _, e := range result.Events {
		ctx.EventManager().EmitEvent(sdk.Event(e))
	}
	return &types.MsgRevokeLiquidityProviderResponse{}, nil
}

func (m msgServer) SetInflation(c context.Context, msg *types.MsgSetInflation) (*types.MsgSetInflationResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	issuer, err := sdk.AccAddressFromBech32(msg.Issuer)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "issuer")
	}

	result, err := m.k.SetInflationRate(ctx, issuer, msg.InflationRate, msg.Denom)
	if err != nil {
		return nil, err
	}
	for _, e := range result.Events {
		ctx.EventManager().EmitEvent(sdk.Event(e))
	}
	return &types.MsgSetInflationResponse{}, nil
}
