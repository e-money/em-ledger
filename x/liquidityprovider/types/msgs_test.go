// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestValidation(t *testing.T) {
	coin1 := sdk.Coin{
		Denom:  "eur",
		Amount: sdk.NewInt(-100),
	}
	coin2 := sdk.NewCoin("chf", sdk.NewInt(500))

	msg1 := MsgMintTokens{
		Amount:            []sdk.Coin{coin1, coin2},
		LiquidityProvider: "invalidAddress",
	}

	msg2 := MsgBurnTokens{
		Amount:            []sdk.Coin{coin1, coin2},
		LiquidityProvider: "invalidAddress",
	}

	require.NotNil(t, msg1.ValidateBasic())
	require.NotNil(t, msg2.ValidateBasic())
}
