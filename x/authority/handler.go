// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package authority

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/e-money/em-ledger/x/authority/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func newHandler(keeper Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (result *sdk.Result, err error) {
		switch msg := msg.(type) {
		case *types.MsgCreateIssuer:
			authority, err := sdk.AccAddressFromBech32(msg.Authority)
			if err != nil {
				return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "authority")
			}
			issuer, err := sdk.AccAddressFromBech32(msg.Issuer)
			if err != nil {
				return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "issuer")
			}
			return keeper.CreateIssuer(ctx, authority, issuer, msg.Denominations)
		case *types.MsgDestroyIssuer:
			authority, err := sdk.AccAddressFromBech32(msg.Authority)
			if err != nil {
				return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "authority")
			}
			issuer, err := sdk.AccAddressFromBech32(msg.Issuer)
			if err != nil {
				return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "issuer")
			}

			return keeper.DestroyIssuer(ctx, authority, issuer)
		case *types.MsgSetGasPrices:
			authority, err := sdk.AccAddressFromBech32(msg.Authority)
			if err != nil {
				return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "authority")
			}
			return keeper.SetGasPrices(ctx, authority, msg.GasPrices)
		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized %s message type: %T", ModuleName, msg)
		}
	}
}
