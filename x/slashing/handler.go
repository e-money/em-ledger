// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package slashing

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/e-money/em-ledger/x/slashing/types"
)

func NewHandler(k Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())

		switch msg := msg.(type) {
		case MsgUnjail:
			return handleMsgUnjail(ctx, msg, k)

		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized slashing message type: %T", msg)
		}
	}
}

// Validators must submit a transaction to unjail itself after
// having been jailed (and thus unbonded) for downtime
func handleMsgUnjail(ctx sdk.Context, msg MsgUnjail, k Keeper) (*sdk.Result, error) {
	// TODO Move all of this business logic to the keeper.
	validator := k.sk.Validator(ctx, msg.ValidatorAddr)
	if validator == nil {
		return nil, ErrNoValidatorForAddress
	}

	// cannot be unjailed if no self-delegation exists
	selfDel := k.sk.Delegation(ctx, sdk.AccAddress(msg.ValidatorAddr), msg.ValidatorAddr)
	if selfDel == nil {
		return nil, ErrMissingSelfDelegation
	}

	if validator.TokensFromShares(selfDel.GetShares()).TruncateInt().LT(validator.GetMinSelfDelegation()) {
		return nil, ErrSelfDelegationTooLowToUnjail
	}

	// cannot be unjailed if not jailed
	if !validator.IsJailed() {
		return nil, ErrValidatorNotJailed
	}

	consAddr := sdk.ConsAddress(validator.GetConsPubKey().Address())

	info, found := k.getValidatorSigningInfo(ctx, consAddr)
	if !found {
		return nil, ErrNoValidatorForAddress
	}

	// cannot be unjailed if tombstoned
	if info.Tombstoned {
		return nil, ErrValidatorJailed
	}

	// cannot be unjailed until out of jail
	if ctx.BlockHeader().Time.Before(info.JailedUntil) {
		return nil, ErrValidatorJailed
	}

	k.sk.Unjail(ctx, consAddr)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.ValidatorAddr.String()),
		),
	)

	return &sdk.Result{Events: ctx.EventManager().Events()}, nil
}
