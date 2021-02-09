// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package types

import (
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestBasic(t *testing.T) {
	priv := ed25519.GenPrivKey()
	addr := sdk.AccAddress(priv.PubKey().Address())

	baseAcc := authtypes.NewBaseAccount(addr, priv.PubKey(), 1, 0)

	mintable := sdk.NewCoin("eeur", sdk.NewIntWithDecimal(1000, 2))
	lpAcc, err := NewLiquidityProviderAccount(baseAcc, sdk.NewCoins(mintable))

	require.NoError(t, err)

	lpAcc.IncreaseMintableAmount(sdk.NewCoins(sdk.NewCoin("ejpy", sdk.NewInt(400))))
	assert.Equal(t, sdk.NewInt(400), lpAcc.Mintable.AmountOf("ejpy"))
}

func TestDecreaseMintable(t *testing.T) {
	priv := ed25519.GenPrivKey()
	addr := sdk.AccAddress(priv.PubKey().Address())

	baseAcc := authtypes.NewBaseAccount(addr, priv.PubKey(), 1, 0)

	mintable := MustParseCoins("100000eeur,700ejpy")
	lpAcc, err := NewLiquidityProviderAccount(baseAcc, mintable)
	require.NoError(t, err)

	reduction := MustParseCoins("200000eeur,300ejpy")
	assert.Panics(t, func() {
		lpAcc.DecreaseMintableAmount(reduction)
	})

	lpAcc.DecreaseMintableAmount(MustParseCoins("90000eeur,300ejpy"))
	assert.Equal(t, MustParseCoins("10000eeur,400ejpy"), lpAcc.Mintable)
}

func MustParseCoins(coins string) sdk.Coins {
	result, err := sdk.ParseCoinsNormalized(coins)
	if err != nil {
		panic(err)
	}

	return result
}
