package types

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"testing"
)

func TestNewOrder(t *testing.T) {
	var (
		src    = sdk.NewCoin("eur", sdk.NewInt(100))
		dst    = sdk.NewCoin("usd", sdk.NewInt(120))
		seller = sdk.AccAddress([]byte("acc1"))
	)
	order := NewOrder(src, dst, seller, "A")
	fmt.Println(order.String())
}
