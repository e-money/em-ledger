package types

import (
	"github.com/stretchr/testify/require"
	"testing"
)

import sdk "github.com/cosmos/cosmos-sdk/types"

func TestValidation(t *testing.T) {
	coin1 := sdk.Coin{
		Denom:  "eur",
		Amount: sdk.NewInt(-100),
	}
	coin2 := sdk.NewCoin("chf", sdk.NewInt(500))

	msg1 := MsgMintTokens{
		Amount:            []sdk.Coin{coin1, coin2},
		LiquidityProvider: sdk.AccAddress("abcd"),
	}

	msg2 := MsgBurnTokens{
		Amount:            []sdk.Coin{coin1, coin2},
		LiquidityProvider: sdk.AccAddress("abcd"),
	}

	require.NotNil(t, msg1.ValidateBasic())
	require.NotNil(t, msg2.ValidateBasic())
}
