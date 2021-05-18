// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package types

import (
	"fmt"
	"strings"
)

func (q QueryInstrumentResponse) String() string {
	sb := new(strings.Builder)

	sb.WriteString(fmt.Sprintf("%v => %v\n", q.Source, q.Destination))

	for _, order := range q.Orders {
		sb.WriteString(order.String())
	}

	return sb.String()
}

func (q QueryOrderResponse) String() string {
	return fmt.Sprintf(" - %v %v %v %v\n", q.ID, q.Price, q.SourceRemaining, q.Owner)
}

func (q QueryByAccountResponse) String() string {
	sb := new(strings.Builder)
	for _, order := range q.Orders {
		sb.WriteString(order.String())
	}

	return sb.String()
}

func (q QueryInstrumentsResponse) String() string {
	sb := new(strings.Builder)
	for _, instrument := range q.Instruments {
		sb.WriteString(instrument.String())
	}

	return sb.String()
}

func (q QueryInstrumentsResponse_Element) String() string {
	return fmt.Sprintf("%v => %v", q.Source, q.Destination)
}
