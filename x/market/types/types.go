// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package types

import (
	"bytes"
	"fmt"
	"github.com/emirpasic/gods/sets/treeset"
	"github.com/emirpasic/gods/utils"
	"strings"
	"time"

	"github.com/emirpasic/gods/trees/btree"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type (
	Instrument struct {
		Source, Destination string

		Orders *btree.Tree
	}

	Instruments []Instrument

	Order struct {
		ID      uint64    `json:"id" yaml:"id"`
		Created time.Time `json:"created" yaml:"created"`

		Source               sdk.Coin `json:"source" yaml:"source"`
		Destination          sdk.Coin `json:"destination" yaml:"destination"`
		DestinationFilled    sdk.Int  `json:"destination_filled" yaml:"destination_filled"`
		DestinationRemaining sdk.Int  `json:"destination_remaining" yaml:"destination_remaining"`

		Owner         sdk.AccAddress `json:"owner" yaml:"owner"`
		ClientOrderID string         `json:"client_order_id" yaml:"client_order_id"`

		price sdk.Dec
		//invertedPrice sdk.Dec
	}

	ExecutionPlan struct {
		Price sdk.Dec

		FirstOrder,
		SecondOrder *Order
	}
)

func (is Instruments) String() string {
	sb := strings.Builder{}

	for _, instr := range is {
		sb.WriteString(fmt.Sprintf("%v/%v - %v\n", instr.Source, instr.Destination, instr.Orders.Size()))
	}

	return sb.String()
}

func (is *Instruments) InsertOrder(order *Order) {
	for _, i := range *is {
		if i.Destination == order.Destination.Denom && i.Source == order.Source.Denom {
			i.Orders.Put(order, nil)
			return
		}
	}

	i := Instrument{
		Source:      order.Source.Denom,
		Destination: order.Destination.Denom,
		Orders:      btree.NewWith(3, OrderPriorityComparator),
	}

	*is = append(*is, i)
	i.Orders.Put(order, nil)
}

func (is *Instruments) GetInstrument(source, destination string) *Instrument {
	for _, i := range *is {
		if i.Source == source && i.Destination == destination {
			return &i
		}
	}

	return nil
}

func (is *Instruments) RemoveInstrument(instr Instrument) {
	for index, v := range *is {
		if instr.Source == v.Source && instr.Destination == v.Destination {
			*is = append((*is)[:index], (*is)[index+1:]...)
			return
		}
	}
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
	return []interface{}{&o.ID, &o.Created, &o.Source, &o.Destination, &o.DestinationFilled, &o.DestinationRemaining, &o.Owner, &o.ClientOrderID, &o.price}
}

// Should return a number:
//    negative , if a < b
//    zero     , if a == b
//    positive , if a > b
func OrderPriorityComparator(a, b interface{}) int {
	aAsserted := a.(*Order)
	bAsserted := b.(*Order)

	// Price priority
	switch {
	case aAsserted.Price().LT(bAsserted.Price()):
		return -1
	case aAsserted.Price().GT(bAsserted.Price()):
		return 1
	}

	// Time priority
	return int(aAsserted.ID - bAsserted.ID)
}

//func (o Order) InvertedPrice() sdk.Dec {
//	return o.invertedPrice
//}

// Signals whether the order can be meaningfully executed, ie will pay for more than one unit of the destination token.
func (o Order) IsFilled() bool {
	return o.DestinationRemaining.IsZero()
}

func (o Order) IsValid() sdk.Error {
	if o.Source.Amount.LTE(sdk.ZeroInt()) {
		return ErrInvalidPrice(o.Source, o.Destination)
	}

	if o.Destination.Amount.LTE(sdk.ZeroInt()) {
		return ErrInvalidPrice(o.Source, o.Destination)
	}

	if o.Source.Denom == o.Destination.Denom {
		return ErrInvalidInstrument(o.Source.Denom, o.Destination.Denom)
	}

	return nil
}

func (o Order) Price() sdk.Dec {
	return o.price
}

func (o Order) String() string {
	return fmt.Sprintf("%d : %v -> %v @ %v (%v remaining) %v", o.ID, o.Source, o.Destination, o.price, o.DestinationRemaining, o.Owner.String())
}

func (ep ExecutionPlan) DestinationCapacity() sdk.Dec {
	if ep.FirstOrder == nil {
		return sdk.ZeroDec()
	}

	res := ep.FirstOrder.DestinationRemaining.ToDec()

	if ep.SecondOrder != nil {
		res = sdk.MinDec(ep.SecondOrder.DestinationRemaining.ToDec(), res.Mul(ep.SecondOrder.Price()))
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

func NewOrder(src, dst sdk.Coin, seller sdk.AccAddress, created time.Time, clientOrderId string) (Order, sdk.Error) {
	if src.Amount.LTE(sdk.ZeroInt()) || dst.Amount.LTE(sdk.ZeroInt()) {
		return Order{}, ErrInvalidPrice(src, dst)
	}

	o := Order{
		Owner:                seller,
		Created:              created,
		Source:               src,
		Destination:          dst,
		DestinationFilled:    sdk.ZeroInt(),
		DestinationRemaining: dst.Amount,
		ClientOrderID:        clientOrderId,
		price:                dst.Amount.ToDec().Quo(src.Amount.ToDec()),
		//invertedPrice:        src.Amount.ToDec().Quo(dst.Amount.ToDec()),
	}

	if err := o.IsValid(); err != nil {
		return Order{}, err
	}

	return o, nil
}

type Orders struct {
	accountOrders map[string]*treeset.Set
}

func NewOrders() Orders {
	return Orders{make(map[string]*treeset.Set)}
}

func (o Orders) ContainsClientOrderId(owner sdk.AccAddress, clientOrderId string) bool {
	allOrders := o.GetAllOrders(owner)

	order := &Order{ClientOrderID: clientOrderId}
	return allOrders.Contains(order)
}

func (o Orders) GetOrder(owner sdk.AccAddress, clientOrderId string) (res *Order) {
	allOrders := o.GetAllOrders(owner)

	allOrders.Find(func(_ int, value interface{}) bool {
		order := value.(*Order)
		if order.ClientOrderID == clientOrderId {
			res = order
			return true
		}

		return false
	})

	return
}

func (o *Orders) GetAllOrders(owner sdk.AccAddress) *treeset.Set {
	allOrders, found := o.accountOrders[owner.String()]

	if !found {
		// Note that comparator only uses client order id.
		allOrders = treeset.NewWith(OrderClientIdComparator)
		o.accountOrders[owner.String()] = allOrders
	}

	return allOrders
}

func (o *Orders) AddOrder(order *Order) {
	orders := o.GetAllOrders(order.Owner)
	orders.Add(order)
}

func (o *Orders) RemoveOrder(order *Order) {
	orders := o.GetAllOrders(order.Owner)
	orders.Remove(order)
}

func OrderClientIdComparator(a, b interface{}) int {
	aAsserted := a.(*Order)
	bAsserted := b.(*Order)

	return utils.StringComparator(aAsserted.ClientOrderID, bAsserted.ClientOrderID)
}
