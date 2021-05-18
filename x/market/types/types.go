// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package types

import (
	"fmt"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (o Order) MarshalJSON() ([]byte, error) {
	s := fmt.Sprintf(`
{
  "order_id": "%v",
  "time_in_force" : "%v",
  "owner": "%v",
  "client_order_id": "%v",
  "price": "%v",
  "source": {
    "denom": "%v",
    "amount": "%v"
  },
  "source_remaining": "%v",
  "source_filled": "%v",
  "destination": {
    "denom": "%v",
    "amount": "%v"
  },
  "destination_filled": "%v",
  "created": "%v",
  "orig_order_created": "%v"
}
`,
		o.ID,
		o.TimeInForce,
		o.Owner,
		o.ClientOrderID,
		o.Price().String(),
		o.Source.Denom,
		o.Source.Amount,
		o.SourceRemaining,
		o.SourceFilled,
		o.Destination.Denom,
		o.Destination.Amount,
		o.DestinationFilled,
		o.Created,
		o.OrigOrderCreated,
	)

	return []byte(s), nil
}

// Signals whether the order can be meaningfully executed, ie will pay for more than one unit of the destination token.
func (o Order) IsFilled() bool {
	return o.SourceRemaining.ToDec().Mul(o.Price()).LT(sdk.OneDec()) || o.DestinationFilled.GTE(o.Destination.Amount)
}

func (o Order) IsValid() error {
	switch o.TimeInForce {
	case TimeInForce_GoodTillCancel, TimeInForce_FillOrKill, TimeInForce_ImmediateOrCancel:
	default:
		return sdkerrors.Wrapf(ErrUnknownTimeInForce, "Unknown 'time in force' specified : %v", o.TimeInForce)
	}

	if o.Source.Amount.LTE(sdk.ZeroInt()) {
		return sdkerrors.Wrapf(ErrInvalidPrice, "Order price is invalid: %s -> %s", o.Source.Amount, o.Destination.Amount)
	}

	if o.Destination.Amount.LTE(sdk.ZeroInt()) {
		return sdkerrors.Wrapf(ErrInvalidPrice, "Order price is invalid: %s -> %s", o.Source.Amount, o.Destination.Amount)
	}

	if o.Source.Denom == o.Destination.Denom {
		return sdkerrors.Wrapf(ErrInvalidInstrument, "'%v/%v' is not a valid instrument", o.Source.Denom, o.Destination.Denom)
	}

	return nil
}

func (o Order) Price() sdk.Dec {
	return o.Destination.Amount.ToDec().Quo(o.Source.Amount.ToDec())
}

func (o Order) String() string {
	return fmt.Sprintf("%d : %v -> %v @ %v\n(%v%v remaining) (%v%v filled) (%v%v filled)\n%v\nCreated:%v\nOriginal Order Created:%v", o.ID, o.Source, o.Destination, o.Price(), o.SourceRemaining, o.Source.Denom, o.SourceFilled, o.Source.Denom, o.DestinationFilled, o.Destination.Denom, o.Owner, o.Created, o.OrigOrderCreated)
}

func (ep ExecutionPlan) DestinationCapacity() sdk.Dec {
	if ep.FirstOrder == nil {
		return sdk.ZeroDec()
	}

	// Find capacity of the first order.
	res := ep.FirstOrder.SourceRemaining.ToDec().Mul(ep.FirstOrder.Price())
	res = sdk.MinDec(res, ep.FirstOrder.Destination.Amount.Sub(ep.FirstOrder.DestinationFilled).ToDec())

	if ep.SecondOrder != nil {
		// Convert first order capacity to second order destination.
		res = res.Mul(ep.SecondOrder.Price())

		// Determine which of the orders have the lowest capacity.
		res = sdk.MinDec(res, ep.SecondOrder.SourceRemaining.ToDec().Mul(ep.SecondOrder.Price()))
		res = sdk.MinDec(res, ep.SecondOrder.Destination.Amount.Sub(ep.SecondOrder.DestinationFilled).ToDec())
	}

	return res
}

func (ep ExecutionPlan) String() string {
	var buf strings.Builder

	var capacityDenom string
	for _, o := range []*Order{ep.FirstOrder, ep.SecondOrder} {
		if o == nil {
			continue
		}

		capacityDenom = o.Destination.Denom
		buf.WriteString(fmt.Sprintf(" - %v\n", o.String()))
	}
	buf.WriteString(fmt.Sprintf("Capacity: %v%s\n", ep.DestinationCapacity(), capacityDenom))
	buf.WriteString(fmt.Sprintf("Price   : %v\n", ep.Price))

	return buf.String()
}

func NewOrder(
	createdTm time.Time,
	timeInForce TimeInForce,
	src, dst sdk.Coin,
	seller sdk.AccAddress,
	clientOrderId string,
    origOrderCreated time.Time) (Order, error) {

	if src.Amount.LTE(sdk.ZeroInt()) || dst.Amount.LTE(sdk.ZeroInt()) {
		return Order{}, sdkerrors.Wrapf(ErrInvalidPrice, "Order price is invalid: %s -> %s", src.Amount, dst.Amount)
	}

	o := Order{
		TimeInForce: timeInForce,

		Owner:         seller.String(),
		ClientOrderID: clientOrderId,

		Source:            src,
		SourceRemaining:   src.Amount,
		SourceFilled:      sdk.ZeroInt(),
		Destination:       dst,
		DestinationFilled: sdk.ZeroInt(),
		Created:           createdTm,
		OrigOrderCreated:  origOrderCreated,
	}

	if err := o.IsValid(); err != nil {
		return Order{}, err
	}

	return o, nil
}

// Convert from TimeInForce string representation to the internal enum type. Case insensitive.
func TimeInForceFromString(p string) (TimeInForce, error) {
	p = strings.ToLower(p)

	switch p {
	case "fok":
		return TimeInForce_FillOrKill, nil
	case "ioc":
		return TimeInForce_ImmediateOrCancel, nil
	case "gtc":
		return TimeInForce_GoodTillCancel, nil
	}

	return 0, fmt.Errorf("unknown time-in-force value: %v", p)
}
