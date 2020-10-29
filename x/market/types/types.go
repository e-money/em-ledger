// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package types

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	_ = iota
	Order_Limit
	Order_Market
)

const (
	_ TimeInForce = iota
	TimeInForce_GoodTilCancel
	TimeInForce_ImmediateOrCancel
	TimeInForce_FillOrKill
)

type (
	TimeInForce int

	Instrument struct {
		Source, Destination string
	}

	Order struct {
		ID uint64 `json:"id" yaml:"id"`

		TimeInForce TimeInForce `json:"time_in_force" yaml:"time_in_force"`

		Owner         sdk.AccAddress `json:"owner" yaml:"owner"`
		ClientOrderID string         `json:"client_order_id" yaml:"client_order_id"`

		Source            sdk.Coin `json:"source" yaml:"source"`
		SourceRemaining   sdk.Int  `json:"source_remaining" yaml:"source_remaining"`
		SourceFilled      sdk.Int  `json:"source_filled" yaml:"source_filled"`
		Destination       sdk.Coin `json:"destination" yaml:"destination"`
		DestinationFilled sdk.Int  `json:"destination_filled" yaml:"destination_filled"`
	}

	ExecutionPlan struct {
		Price sdk.Dec

		FirstOrder,
		SecondOrder *Order
	}

	MarketData struct {
		Source, Destination string
		LastPrice           *sdk.Dec
		Timestamp           *time.Time
	}
)

func (o Order) MarshalJSON() ([]byte, error) {
	s := fmt.Sprintf(`
{
  "id": %v,
  "time_in_force" : %v,
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
}
`,
		o.ID,
		o.TimeInForce,
		o.Owner.String(),
		o.ClientOrderID,
		o.Price().String(),
		o.Source.Denom,
		o.Source.Amount,
		o.SourceRemaining,
		o.SourceFilled,
		o.Destination.Denom,
		o.Destination.Amount,
		o.DestinationFilled,
	)

	return []byte(s), nil
}

// Manual handling of de-/serialization in order to include private fields
func (o Order) MarshalAmino() ([]byte, error) {
	w := new(bytes.Buffer)

	for _, v := range o.allFields() {
		_, err := ModuleCdc.MarshalBinaryLengthPrefixedWriter(w, v)
		if err != nil {
			return []byte{}, err
		}
	}

	return w.Bytes(), nil
}

func (o *Order) UnmarshalAmino(bz []byte) error {
	r := bytes.NewBuffer(bz)

	for _, v := range o.allFields() {
		_, err := ModuleCdc.UnmarshalBinaryLengthPrefixedReader(r, v, 1024)
		if err != nil {
			return err
		}
	}

	return nil
}

// Ensure field order of de-/serialization
func (o *Order) allFields() []interface{} {
	return []interface{}{
		&o.ID,

		&o.TimeInForce,

		&o.Owner,
		&o.ClientOrderID,

		&o.Source,
		&o.SourceRemaining,
		&o.SourceFilled,
		&o.Destination,
		&o.DestinationFilled,
	}
}

// Signals whether the order can be meaningfully executed, ie will pay for more than one unit of the destination token.
func (o Order) IsFilled() bool {
	return o.SourceRemaining.ToDec().Mul(o.Price()).LT(sdk.OneDec()) || o.DestinationFilled.GTE(o.Destination.Amount)
}

func (o Order) IsValid() error {
	switch o.TimeInForce {
	case TimeInForce_GoodTilCancel, TimeInForce_FillOrKill, TimeInForce_ImmediateOrCancel:
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
	return fmt.Sprintf("%d : %v -> %v @ %v\n(%v%v remaining) (%v%v filled) (%v%v filled)\n%v", o.ID, o.Source, o.Destination, o.Price(), o.SourceRemaining, o.Source.Denom, o.SourceFilled, o.Source.Denom, o.DestinationFilled, o.Destination.Denom, o.Owner.String())
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

func NewOrder(timeInForce TimeInForce, src, dst sdk.Coin, seller sdk.AccAddress, clientOrderId string) (Order, error) {
	if src.Amount.LTE(sdk.ZeroInt()) || dst.Amount.LTE(sdk.ZeroInt()) {
		return Order{}, sdkerrors.Wrapf(ErrInvalidPrice, "Order price is invalid: %s -> %s", src.Amount, dst.Amount)
	}

	o := Order{
		TimeInForce: timeInForce,

		Owner:         seller,
		ClientOrderID: clientOrderId,

		Source:            src,
		SourceRemaining:   src.Amount,
		SourceFilled:      sdk.ZeroInt(),
		Destination:       dst,
		DestinationFilled: sdk.ZeroInt(),
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
	case "fillorkill":
		return TimeInForce_FillOrKill, nil
	case "immediateorcancel":
		return TimeInForce_ImmediateOrCancel, nil
	case "goodtilcancelled":
		return TimeInForce_GoodTilCancel, nil
	}

	return 0, fmt.Errorf("unknown time-in-force value: %v", p)
}

func (tif TimeInForce) String() string {
	switch tif {
	case TimeInForce_ImmediateOrCancel:
		return "ImmediateOrCancel"
	case TimeInForce_GoodTilCancel:
		return "GoodTilCancelled"
	case TimeInForce_FillOrKill:
		return "FillOrKill"
	}
	return "<unknown>"
}
