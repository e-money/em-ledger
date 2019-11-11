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

	mintable := sdk.NewCoin("x2eur", sdk.NewIntWithDecimal(1000, 2))
	lpAcc := NewLiquidityProviderAccount(baseAcc, sdk.NewCoins(mintable))

	// The mintable balance is not considered spendable.
	assert.Equal(t, sdk.NewCoins(), lpAcc.SpendableCoins(time.Now()))

	lpAcc.IncreaseMintableAmount(sdk.NewCoins(sdk.NewCoin("x0jpy", sdk.NewInt(400))))
	assert.Equal(t, sdk.NewInt(400), lpAcc.Mintable.AmountOf("x0jpy"))
}

func TestDecreaseMintable(t *testing.T) {
	priv := ed25519.GenPrivKey()
	addr := sdk.AccAddress(priv.PubKey().Address())

	baseAcc := auth.NewBaseAccount(addr, sdk.NewCoins(), priv.PubKey(), 1, 0)

	mintable := MustParseCoins("100000x2eur,700x0jpy")
	lpAcc := NewLiquidityProviderAccount(baseAcc, mintable)

	reduction := MustParseCoins("200000x2eur,300x0jpy")
	assert.Panics(t, func() {
		lpAcc.DecreaseMintableAmount(reduction)
	})

	lpAcc.DecreaseMintableAmount(MustParseCoins("90000x2eur,300x0jpy"))
	assert.Equal(t, MustParseCoins("10000x2eur,400x0jpy"), lpAcc.Mintable)
}

func MustParseCoins(coins string) sdk.Coins {
	result, err := sdk.ParseCoins(coins)
	if err != nil {
		panic(err)
	}

	return result
}
