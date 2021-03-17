// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/rand"
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

func TestMarshalUnmarshal(t *testing.T) {
	var myAddr sdk.AccAddress = rand.Bytes(sdk.AddrLen)
	nestedAcc := authtypes.NewBaseAccountWithAddress(myAddr)
	mintableCoins := sdk.NewCoins(sdk.NewCoin("alx", sdk.NewInt(123)))
	src, err := NewLiquidityProviderAccount(nestedAcc, mintableCoins)
	require.NoError(t, err)

	interfaceRegistry := types.NewInterfaceRegistry()
	marshaler := codec.NewProtoCodec(interfaceRegistry)
	authtypes.RegisterInterfaces(interfaceRegistry)
	RegisterInterfaces(interfaceRegistry)

	// encode
	bz, err := marshaler.MarshalBinaryBare(src)
	require.NoError(t, err)
	// and decode to type
	var dest LiquidityProviderAccount
	err = marshaler.UnmarshalBinaryBare(bz, &dest)
	require.NoError(t, err)
	assert.Equal(t, src, &dest)
	assert.Equal(t, myAddr, dest.GetAddress())
}
func TestValidate(t *testing.T) {
	var (
		randomAddress sdk.AccAddress = rand.Bytes(sdk.AddrLen)
		randomPubKey                 = ed25519.GenPrivKey().PubKey()
	)

	specs := map[string]struct {
		srcAccount  authtypes.AccountI
		srcMintable sdk.Coins
		expErr      bool
	}{
		"all good": {
			srcAccount:  authtypes.NewBaseAccountWithAddress(randomAddress),
			srcMintable: sdk.Coins{sdk.NewCoin("foo", sdk.OneInt())},
		},
		"empty coins allowed": {
			srcAccount:  authtypes.NewBaseAccountWithAddress(randomAddress),
			srcMintable: sdk.Coins{},
		},
		"invalid coin rejected": {
			srcAccount:  authtypes.NewBaseAccountWithAddress(randomAddress),
			srcMintable: sdk.Coins{sdk.Coin{Denom: "invalid@#$^", Amount: sdk.OneInt()}},
			expErr:      true,
		},
		"invalid account rejected: non matching pubkey": {
			srcAccount:  authtypes.NewBaseAccount(randomAddress, randomPubKey, 1, 1),
			srcMintable: sdk.Coins{sdk.NewCoin("foo", sdk.OneInt())},
			expErr:      true,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			lp, err := NewLiquidityProviderAccount(spec.srcAccount, spec.srcMintable)
			require.NoError(t, err)
			gotErr := lp.Validate()
			if spec.expErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
		})
	}

}
