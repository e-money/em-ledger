// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package types

import (
	"fmt"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// market module event types
const (
	EventTypeCancel = "market_cancel"
	EventTypeFilled = "market_fill"
	EventNewOrder   = "market_new"

	AttributeKeyClientOrderID     = "client_order_id"
	AttributeKeyOrderID           = "order_id"
	AttributeKeyOwner             = "owner"
	AttributeKeyPartialFill       = "partial_fill"
	AttributeKeyPrice             = "price"
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
			sdk.NewAttribute(AttributeKeyOrderID, fmt.Sprintf("%d", order.ID)),
			sdk.NewAttribute(AttributeKeyOwner, order.Owner.String()),
			sdk.NewAttribute(AttributeKeyClientOrderID, order.ClientOrderID),
		),
	)
}

func EmitNewOrderEvent(ctx sdk.Context, order Order) {
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(EventNewOrder,
			sdk.NewAttribute(AttributeKeyOrderID, fmt.Sprintf("%d", order.ID)),
			sdk.NewAttribute(AttributeKeyOwner, order.Owner.String()),
			sdk.NewAttribute(AttributeKeyClientOrderID, order.ClientOrderID),
			sdk.NewAttribute(AttributeKeySource, order.Source.String()),
			sdk.NewAttribute(AttributeKeyDestination, order.Destination.String()),
		),
	)
}

func EmitFilledEvent(ctx sdk.Context, order Order, partial_fill bool, price sdk.Dec) {
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(EventTypeFilled,
			sdk.NewAttribute(AttributeKeyOrderID, fmt.Sprintf("%d", order.ID)),
			sdk.NewAttribute(AttributeKeyOwner, order.Owner.String()),
			sdk.NewAttribute(AttributeKeyPartialFill, strconv.FormatBool(partial_fill)),
			sdk.NewAttribute(AttributeKeyPrice, price.String()),
			sdk.NewAttribute(AttributeKeyClientOrderID, order.ClientOrderID),
			sdk.NewAttribute(AttributeKeySource, order.Source.String()),
			sdk.NewAttribute(AttributeKeySourceRemaining, fmt.Sprintf("%v%v", order.SourceRemaining.String(), order.Source.Denom)),
			sdk.NewAttribute(AttributeKeySourceFilled, fmt.Sprintf("%v%v", order.SourceFilled.String(), order.Source.Denom)),
			sdk.NewAttribute(AttributeKeyDestination, order.Destination.String()),
			sdk.NewAttribute(AttributeKeyDestinationFilled, fmt.Sprintf("%v%v", order.DestinationFilled.String(), order.Destination.Denom)),
		),
	)
}
