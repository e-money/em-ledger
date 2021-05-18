// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package types

import (
	"fmt"
	"strconv"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// market module event types
const (
	EventTypeMarket = ModuleName

	AttributeKeyAction            = "action"
	AttributeKeyOrderID           = "order_id"
	AttributeKeyOwner             = "owner"
	AttributeKeyClientOrderID     = "client_order_id"
	AttributeKeyPrice             = "price"
	AttributeKeySource            = "source"
	AttributeKeySourceRemaining   = "source_remaining"
	AttributeKeySourceFilled      = "source_filled"
	AttributeKeyDestination       = "destination"
	AttributeKeyDestinationFilled = "destination_filled"
	AttributeKeyAggressive        = "aggressive"
	AttributeKeyCreated           = "created"
)

func EmitAcceptEvent(ctx sdk.Context, order Order) {
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(EventTypeMarket,
			sdk.NewAttribute(AttributeKeyAction, "accept"),
			sdk.NewAttribute(AttributeKeyOrderID, fmt.Sprintf("%d", order.ID)),
			sdk.NewAttribute(AttributeKeyOwner, order.Owner),
			sdk.NewAttribute(AttributeKeyClientOrderID, order.ClientOrderID),
			sdk.NewAttribute(AttributeKeySource, order.Source.String()),
			sdk.NewAttribute(AttributeKeyDestination, order.Destination.String()),
			sdk.NewAttribute(AttributeKeyCreated, order.Created.Format(time.RFC3339)),
		),
	)
}

func EmitExpireEvent(ctx sdk.Context, order Order) {
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(EventTypeMarket,
			sdk.NewAttribute(AttributeKeyAction, "expire"),
			sdk.NewAttribute(AttributeKeyOrderID, fmt.Sprintf("%d", order.ID)),
			sdk.NewAttribute(AttributeKeyOwner, order.Owner),
			sdk.NewAttribute(AttributeKeyClientOrderID, order.ClientOrderID),
			sdk.NewAttribute(AttributeKeySource, order.Source.String()),
			sdk.NewAttribute(AttributeKeySourceFilled, fmt.Sprintf("%v%v", order.SourceFilled.String(), order.Source.Denom)),
			sdk.NewAttribute(AttributeKeySourceRemaining, fmt.Sprintf("%v%v", order.SourceRemaining.String(), order.Source.Denom)),
			sdk.NewAttribute(AttributeKeyDestination, order.Destination.String()),
			sdk.NewAttribute(AttributeKeyDestinationFilled, fmt.Sprintf("%v%v", order.DestinationFilled.String(), order.Destination.Denom)),
		),
	)
}

func EmitFillEvent(ctx sdk.Context, order Order, aggressive bool, sourceFilled sdk.Int, destinationFilled sdk.Int) {
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(EventTypeMarket,
			sdk.NewAttribute(AttributeKeyAction, "fill"),
			sdk.NewAttribute(AttributeKeyOrderID, fmt.Sprintf("%d", order.ID)),
			sdk.NewAttribute(AttributeKeyOwner, order.Owner),
			sdk.NewAttribute(AttributeKeyClientOrderID, order.ClientOrderID),
			sdk.NewAttribute(AttributeKeyAggressive, strconv.FormatBool(aggressive)),
			sdk.NewAttribute(AttributeKeySourceFilled, fmt.Sprintf("%v%v", sourceFilled.String(), order.Source.Denom)),
			sdk.NewAttribute(AttributeKeyDestinationFilled, fmt.Sprintf("%v%v", destinationFilled.String(), order.Destination.Denom)),
		),
	)
}

func EmitUpdateEvent(ctx sdk.Context, order Order) {
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(EventTypeMarket,
			sdk.NewAttribute(AttributeKeyAction, "update"),
			sdk.NewAttribute(AttributeKeyOrderID, fmt.Sprintf("%d", order.ID)),
			sdk.NewAttribute(AttributeKeyOwner, order.Owner),
			sdk.NewAttribute(AttributeKeyClientOrderID, order.ClientOrderID),
			sdk.NewAttribute(AttributeKeySourceRemaining, fmt.Sprintf("%v%v", order.SourceRemaining.String(), order.Source.Denom)),
		),
	)
}
