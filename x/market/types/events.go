// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// market module event types
const (
	EventTypeCancel = "market_cancel"
	EventTypeFill   = "market_fill"
	EventTypeTouch  = "market_touch"

	AttributeKeyClientOrderID     = "client_order_id"
	AttributeKeyOrderID           = "order_id"
	AttributeKeyOwner             = "owner"
	AttributeKeySource            = "source"
	AttributeKeySourceRemaining   = "source_remaining"
	AttributeKeySourceFilled      = "source_filled"
	AttributeKeyDestination       = "destination"
	AttributeKeyDestinationFilled = "destination_filled"

	AttributeValueCategory = ModuleName
)

func EmitCancelEvent(ctx sdk.Context, order Order) {
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(EventTypeCancel,
			sdk.NewAttribute(AttributeKeyClientOrderID, order.ClientOrderID),
			sdk.NewAttribute(AttributeKeyOrderID, fmt.Sprintf("%d", order.ID)),
			sdk.NewAttribute(AttributeKeyOwner, order.Owner.String()),
		),
	)
}
func EmitFilledEvent(ctx sdk.Context, order Order) {
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(EventTypeFill,
			sdk.NewAttribute(AttributeKeyClientOrderID, order.ClientOrderID),
			sdk.NewAttribute(AttributeKeyOrderID, fmt.Sprintf("%d", order.ID)),
			sdk.NewAttribute(AttributeKeyOwner, order.Owner.String()),
			sdk.NewAttribute(AttributeKeySource, order.Source.String()),
			sdk.NewAttribute(AttributeKeySourceRemaining, fmt.Sprintf("%v%v", order.SourceRemaining.String(), order.Source.Denom)),
			sdk.NewAttribute(AttributeKeySourceFilled, fmt.Sprintf("%v%v", order.SourceFilled.String(), order.Source.Denom)),
			sdk.NewAttribute(AttributeKeyDestination, order.Destination.String()),
			sdk.NewAttribute(AttributeKeyDestinationFilled, fmt.Sprintf("%v%v", order.DestinationFilled.String(), order.Destination.Denom)),
		),
	)
}
