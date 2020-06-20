// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package types

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/tendermint/tendermint/crypto/ed25519"

	"github.com/stretchr/testify/assert"
)

func TestBasic(t *testing.T) {
	priv := ed25519.GenPrivKey()
	addr := sdk.AccAddress(priv.PubKey().Address())

	baseAcc := auth.NewBaseAccount(addr, sdk.NewCoins(), priv.PubKey(), 1, 0)

	mintable := sdk.NewCoin("eeur", sdk.NewIntWithDecimal(1000, 2))
	lpAcc := NewLiquidityProviderAccount(baseAcc, sdk.NewCoins(mintable))

	// The mintable balance is not considered spendable.
	assert.Equal(t, sdk.NewCoins(), lpAcc.SpendableCoins(time.Now()))

	lpAcc.IncreaseMintableAmount(sdk.NewCoins(sdk.NewCoin("ejpy", sdk.NewInt(400))))
	assert.Equal(t, sdk.NewInt(400), lpAcc.Mintable.AmountOf("ejpy"))
}

func TestDecreaseMintable(t *testing.T) {
	priv := ed25519.GenPrivKey()
	addr := sdk.AccAddress(priv.PubKey().Address())

	baseAcc := auth.NewBaseAccount(addr, sdk.NewCoins(), priv.PubKey(), 1, 0)

	mintable := MustParseCoins("100000eeur,700ejpy")
	lpAcc := NewLiquidityProviderAccount(baseAcc, mintable)

	reduction := MustParseCoins("200000eeur,300ejpy")
	assert.Panics(t, func() {
		lpAcc.DecreaseMintableAmount(reduction)
	})

	lpAcc.DecreaseMintableAmount(MustParseCoins("90000eeur,300ejpy"))
	assert.Equal(t, MustParseCoins("10000eeur,400ejpy"), lpAcc.Mintable)
}

func MustParseCoins(coins string) sdk.Coins {
	result, err := sdk.ParseCoins(coins)
	if err != nil {
		panic(err)
	}

	return result
}
