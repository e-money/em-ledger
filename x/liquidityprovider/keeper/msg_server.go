package keeper

import (
	"context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/e-money/em-ledger/x/liquidityprovider/types"
)

var _ types.MsgServer = msgServer{}

type liquidityProvKeeper interface {
	MintTokens(ctx sdk.Context, liquidityProvider string, amount sdk.Coins) (*sdk.Result, error)
	BurnTokensFromBalance(ctx sdk.Context, liquidityProvider string, amount sdk.Coins) (*sdk.Result, error)
}
type msgServer struct {
	k liquidityProvKeeper
}

func NewMsgServerImpl(keeper liquidityProvKeeper) types.MsgServer {
	return &msgServer{k: keeper}
}

func (m msgServer) MintTokens(c context.Context, msg *types.MsgMintTokens) (*types.MsgMintTokensResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	_, err := sdk.AccAddressFromBech32(msg.LiquidityProvider)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "liquidity provider")
	}
	result, err := m.k.MintTokens(ctx, msg.LiquidityProvider, msg.Amount)

	if err != nil {
		return nil, err
	}
	for _, e := range result.Events {
		ctx.EventManager().EmitEvent(sdk.Event(e))
	}
	return &types.MsgMintTokensResponse{}, nil

}

func (m msgServer) BurnTokens(c context.Context, msg *types.MsgBurnTokens) (*types.MsgBurnTokensResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	_, err := sdk.AccAddressFromBech32(msg.LiquidityProvider)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "liquidity provider")
	}
	result, err := m.k.BurnTokensFromBalance(ctx, msg.LiquidityProvider, msg.Amount)
	if err != nil {
		return nil, err
	}
	for _, e := range result.Events {
		ctx.EventManager().EmitEvent(sdk.Event(e))
	}
	return &types.MsgBurnTokensResponse{}, nil
}
