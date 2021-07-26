package keeper

import (
	"context"

	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/e-money/em-ledger/x/authority/types"
)

var _ types.MsgServer = msgServer{}

type authorityKeeper interface {
	createIssuer(ctx sdk.Context, authority sdk.AccAddress, issuerAddress sdk.AccAddress, denoms []string) (*sdk.Result, error)
	destroyIssuer(ctx sdk.Context, authority sdk.AccAddress, issuerAddress sdk.AccAddress) (*sdk.Result, error)
	replaceAuthority(ctx sdk.Context, authority, newAuthority sdk.AccAddress) (*sdk.Result, error)
	SetGasPrices(ctx sdk.Context, authority sdk.AccAddress, gasprices sdk.DecCoins) (*sdk.Result, error)
	ScheduleUpgrade(ctx sdk.Context, authority sdk.AccAddress, plan upgradetypes.Plan) (*sdk.Result, error)
	ApplyUpgrade(ctx sdk.Context, authority sdk.AccAddress, plan upgradetypes.Plan) (*sdk.Result, error)
	GetUpgradePlan(ctx sdk.Context) (plan upgradetypes.Plan, havePlan bool)
}
type msgServer struct {
	k authorityKeeper
}

func NewMsgServerImpl(keeper authorityKeeper) types.MsgServer {
	return &msgServer{k: keeper}
}

func (m msgServer) CreateIssuer(goCtx context.Context, msg *types.MsgCreateIssuer) (*types.MsgCreateIssuerResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	authority, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "authority")
	}
	issuer, err := sdk.AccAddressFromBech32(msg.Issuer)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "issuer")
	}
	result, err := m.k.createIssuer(ctx, authority, issuer, msg.Denominations)
	if err != nil {
		return nil, err
	}
	for _, e := range result.Events {
		ctx.EventManager().EmitEvent(sdk.Event(e))
	}
	return &types.MsgCreateIssuerResponse{}, nil
}

func (m msgServer) DestroyIssuer(goCtx context.Context, msg *types.MsgDestroyIssuer) (*types.MsgDestroyIssuerResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	authority, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "authority")
	}
	issuer, err := sdk.AccAddressFromBech32(msg.Issuer)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "issuer")
	}

	result, err := m.k.destroyIssuer(ctx, authority, issuer)
	if err != nil {
		return nil, err
	}

	for _, e := range result.Events {
		ctx.EventManager().EmitEvent(sdk.Event(e))
	}
	return &types.MsgDestroyIssuerResponse{}, nil
}

func (m msgServer) SetGasPrices(goCtx context.Context, msg *types.MsgSetGasPrices) (*types.MsgSetGasPricesResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	authority, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "authority")
	}

	result, err := m.k.SetGasPrices(ctx, authority, msg.GasPrices)
	if err != nil {
		return nil, err
	}

	for _, e := range result.Events {
		ctx.EventManager().EmitEvent(sdk.Event(e))
	}
	return &types.MsgSetGasPricesResponse{}, nil
}

func (m msgServer) ReplaceAuthority(goCtx context.Context, msg *types.MsgReplaceAuthority) (*types.MsgReplaceAuthorityResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	authorityAcc, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "authority: "+msg.Authority)
	}

	newAuthorityAcc, err := sdk.AccAddressFromBech32(msg.NewAuthority)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "new authority: "+msg.NewAuthority)
	}

	result, err := m.k.replaceAuthority(ctx, authorityAcc, newAuthorityAcc)
	if err != nil {
		return nil, err
	}

	for _, e := range result.Events {
		ctx.EventManager().EmitEvent(sdk.Event(e))
	}

	return &types.MsgReplaceAuthorityResponse{}, nil
}

func (m msgServer) ScheduleUpgrade(
	goCtx context.Context, msg *types.MsgScheduleUpgrade,
) (*types.MsgScheduleUpgradeResponse, error) {

	ctx := sdk.UnwrapSDKContext(goCtx)

	authority, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "authority")
	}

	result, err := m.k.ScheduleUpgrade(ctx, authority, msg.Plan)
	if err != nil {
		return nil, err
	}

	for _, e := range result.Events {
		ctx.EventManager().EmitEvent(sdk.Event(e))
	}

	return &types.MsgScheduleUpgradeResponse{}, nil
}

func (m msgServer) ApplyUpgrade(
	goCtx context.Context, msg *types.MsgApplyUpgrade,
) (*types.MsgApplyUpgradeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	authority, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "authority")
	}

	result, err := m.k.ApplyUpgrade(ctx, authority, msg.Plan)
	if err != nil {
		return nil, err
	}

	for _, e := range result.Events {
		ctx.EventManager().EmitEvent(sdk.Event(e))
	}

	return &types.MsgApplyUpgradeResponse{}, nil
}
