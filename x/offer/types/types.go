package types

import (
	"fmt"
	"strings"

	"github.com/Workiva/go-datastructures/queue"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Instrument struct {
	Source, Destination string

	Orders *queue.PriorityQueue
}

type Instruments []Instrument

func (is Instruments) String() string {
	sb := strings.Builder{}

	for _, instr := range is {
		sb.WriteString(fmt.Sprintf("%v/%v - %v\n", instr.Source, instr.Destination, instr.Orders.Len()))
	}

	return sb.String()
}

func (is *Instruments) InsertOrder(order *Order) {
	for _, i := range *is {
		if i.Destination == order.Destination.Denom && i.Source == order.Source.Denom {
			i.Orders.Put(order)
			return
		}
	}

	i := Instrument{
		Source:      order.Source.Denom,
		Destination: order.Destination.Denom,
		Orders:      queue.NewPriorityQueue(1, true),
	}

	*is = append(*is, i)
	i.Orders.Put(order)
}

func (is *Instruments) RemoveInstrument(instr Instrument) {
	for index, v := range *is {
		if instr.Source == v.Source && instr.Destination == v.Destination {
			*is = append((*is)[:index], (*is)[index+1:]...)
			return
		}
	}
}

var _ queue.Item = Order{}

type Order struct {
	ID uint64

	Source, Destination sdk.Coin
	Remaining           sdk.Int

	SourceAccount sdk.AccAddress

	price,
	invertedPrice float64
}

func (o Order) Compare(other queue.Item) int {
	ot := other.(*Order)
	switch {
	case o.price > ot.price:
		return 1
	case o.price < ot.price:
		return -1
	}

	// Prices are equale. The oldest order gets to go first.
	switch {
	case o.ID > ot.ID:
		return 1
	case o.ID < ot.ID:
		return -1
	default:
		return 0
	}
}

func (o Order) InvertedPrice() float64 {
	return o.invertedPrice
}

func (o Order) Price() float64 {
	return o.price
}

func (o Order) String() string {
	return fmt.Sprintf("%d : %v -> %v @ %v/%v (%v remaining)", o.ID, o.Source, o.Destination, o.price, o.invertedPrice, o.Remaining)
}

func NewOrder(src, dst sdk.Coin, seller sdk.AccAddress) *Order {
	return &Order{
		SourceAccount: seller,
		Source:        src,
		Destination:   dst,
		Remaining:     src.Amount,
		price:         float64(dst.Amount.Int64()) / float64(src.Amount.Int64()),
		invertedPrice: float64(src.Amount.Int64()) / float64(dst.Amount.Int64()),
	}
}
