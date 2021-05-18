package types

import (
	"fmt"
	"strings"
)

func (q QueryGasPricesResponse) String() string {
	sb := new(strings.Builder)
	sb.WriteString("Minimum gas prices\n")
	for _, gp := range q.MinGasPrices {
		sb.WriteString(fmt.Sprintf(" - %v : %v\n", gp.Denom, gp.Amount.String()))
	}

	return sb.String()
}
