package types

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBasic(t *testing.T) {
	priv := ed25519.GenPrivKey()
	addr := sdk.AccAddress(priv.PubKey().Address())

	baseAcc := authtypes.NewBaseAccount(addr, priv.PubKey(), 1, 0)

	mintable := sdk.NewCoin("eeur", sdk.NewIntWithDecimal(1000, 2))
	lpAcc, err := NewLiquidityProviderAccount(baseAcc.GetAddress().String(), sdk.NewCoins(mintable))

	require.NoError(t, err)

	lpAcc.IncreaseMintableAmount(sdk.NewCoins(sdk.NewCoin("ejpy", sdk.NewInt(400))))
	assert.Equal(t, sdk.NewInt(400), lpAcc.Mintable.AmountOf("ejpy"))
}

func TestDecreaseMintable(t *testing.T) {
	priv := ed25519.GenPrivKey()
	addr := sdk.AccAddress(priv.PubKey().Address())

	baseAcc := authtypes.NewBaseAccount(addr, priv.PubKey(), 1, 0)

	mintable := MustParseCoins("100000eeur,700ejpy")
	lpAcc, err := NewLiquidityProviderAccount(baseAcc.GetAddress().String(), mintable)
	require.NoError(t, err)

	reduction := MustParseCoins("200000eeur,300ejpy")
	err = lpAcc.DecreaseMintableAmount(reduction)
	assert.Error(t, err)

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
	var myAddr sdk.AccAddress = sdk.AccAddress("emoney1n5ggspeff4fxc87dvmg0ematr3qzw5l4v20mdv")
	nestedAcc := authtypes.NewBaseAccountWithAddress(myAddr)
	mintableCoins := sdk.NewCoins(sdk.NewCoin("alx", sdk.NewInt(123)))
	src, err := NewLiquidityProviderAccount(nestedAcc.GetAddress().String(), mintableCoins)
	require.NoError(t, err)

	interfaceRegistry := types.NewInterfaceRegistry()
	marshaler := codec.NewProtoCodec(interfaceRegistry)
	authtypes.RegisterInterfaces(interfaceRegistry)
	RegisterInterfaces(interfaceRegistry)

	// encode
	bz, err := marshaler.Marshal(src)
	require.NoError(t, err)
	// and decode to type
	var dest LiquidityProviderAccount
	err = marshaler.Unmarshal(bz, &dest)
	require.NoError(t, err)
	assert.Equal(t, src, &dest)
	assert.Equal(t, myAddr.String(), dest.Address)
}

func TestValidate(t *testing.T) {
	var randomAddress sdk.AccAddress = sdk.AccAddress("emoney1n5ggspeff4fxc87dvmg0ematr3qzw5l4v20mdv")

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
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			lp, err := NewLiquidityProviderAccount(spec.srcAccount.GetAddress().String(), spec.srcMintable)
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
