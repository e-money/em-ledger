// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package bank

import (
	"fmt"
	authexported "github.com/cosmos/cosmos-sdk/x/auth/exported"
	types2 "github.com/e-money/em-ledger/x/authority/types"
	"testing"

	"github.com/stretchr/testify/require"

	emauth "github.com/e-money/em-ledger/hooks/auth"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/params"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"
)

func TestProxySendCoins(t *testing.T) {
	ctx, ak, bk := createTestComponents(t)

	var (
		acc1 = createAccount(ctx, ak, "acc1", "150000gbp, 150000usd, 150000sek")
		acc2 = createAccount(ctx, ak, "acc2", "150000gbp, 150000usd, 150000sek")
		dest = sdk.AccAddress([]byte("dest"))
	)

	bk.rk = restrictedKeeper{
		RestrictedDenoms: []types2.RestrictedDenom{
			{"gbp", []sdk.AccAddress{}},
			{"usd", []sdk.AccAddress{acc1.GetAddress()}},
		},
	}

	var testdata = []struct {
		denom string
		acc   sdk.AccAddress
		valid bool
	}{
		{"gbp", acc2.GetAddress(), false},
		{"usd", acc2.GetAddress(), false},
		{"gbp", acc1.GetAddress(), false},
		{"usd", acc1.GetAddress(), true},
		{"sek", acc1.GetAddress(), true},
		{"sek", acc2.GetAddress(), true},
	}

	for _, d := range testdata {
		c := fmt.Sprintf("1000%s", d.denom)
		err := bk.SendCoins(ctx, d.acc, dest, coins(c))
		if d.valid {
			require.NoError(t, err)
		} else {
			require.True(t, ErrRestrictedDenomination.Is(err), "Actual error \"%s\" (%T)", err.Error(), err)
		}
	}
}

func TestInputOutputCoins(t *testing.T) {
	ctx, ak, bk := createTestComponents(t)

	var (
		acc1 = createAccount(ctx, ak, "acc1", "150000gbp, 150000usd, 150000sek")
		acc2 = createAccount(ctx, ak, "acc2", "150000gbp, 150000usd, 150000sek")
		acc3 = createAccount(ctx, ak, "acc3", "")
	)

	// For simplicity's sake, inputoutput will reject any transaction that includes restricted denominations.

	bk.rk = restrictedKeeper{
		RestrictedDenoms: []types2.RestrictedDenom{
			{"gbp", []sdk.AccAddress{}},
			{"usd", []sdk.AccAddress{acc1.GetAddress()}},
		},
	}

	var testdata = []struct {
		inputs  []bank.Input
		outputs []bank.Output
		valid   bool
	}{
		{[]bank.Input{}, []bank.Output{}, true},
		{
			inputs: []bank.Input{
				{acc1.GetAddress(), coins("1000sek")},
			},
			outputs: []bank.Output{
				{acc2.GetAddress(), coins("500sek")},
				{acc3.GetAddress(), coins("500sek")},
			},
			valid: true,
		},
		{
			inputs: []bank.Input{
				{acc1.GetAddress(), coins("500sek, 1000gbp")},
			},
			outputs: []bank.Output{
				{acc2.GetAddress(), coins("500sek, 500gbp")},
				{acc3.GetAddress(), coins("500gbp")},
			},
			valid: false,
		},
		{
			inputs: []bank.Input{
				{acc1.GetAddress(), coins("1000usd")},
			},
			outputs: []bank.Output{
				{acc2.GetAddress(), coins("1000usd")},
			},
			valid: false,
		},
	}

	for _, d := range testdata {
		err := bk.InputOutputCoins(ctx, d.inputs, d.outputs)
		if d.valid {
			require.NoError(t, err)
		} else {
			require.True(t, ErrRestrictedDenomination.Is(err), "Actual error \"%s\" (%T)", err.Error(), err)
		}
	}

	fmt.Println(ak.GetAccount(ctx, acc3.GetAddress()).GetCoins())
}

func createTestComponents(t *testing.T) (sdk.Context, auth.AccountKeeper, ProxyKeeper) {
	var (
		authCapKey = sdk.NewKVStoreKey("authCapKey")
		keyParams  = sdk.NewKVStoreKey("params")
		tkeyParams = sdk.NewTransientStoreKey("transient_params")

		blacklistedAddrs = make(map[string]bool)
	)

	cdc := createCodec()

	db := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(authCapKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(tkeyParams, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyParams, sdk.StoreTypeIAVL, db)

	err := ms.LoadLatestVersion()
	require.Nil(t, err)

	pk := params.NewKeeper(cdc, keyParams, tkeyParams)

	ctx := sdk.NewContext(ms, abci.Header{ChainID: "test-chain"}, true, log.NewNopLogger())
	accountKeeper := auth.NewAccountKeeper(cdc, authCapKey, pk.Subspace(auth.DefaultParamspace), auth.ProtoBaseAccount)
	accountKeeperWrapped := emauth.Wrap(accountKeeper)

	bankKeeper := bank.NewBaseKeeper(accountKeeperWrapped, pk.Subspace(bank.DefaultParamspace), blacklistedAddrs)

	wrappedBK := Wrap(bankKeeper, restrictedKeeper{})

	return ctx, accountKeeper, wrappedBK
}

type restrictedKeeper struct {
	RestrictedDenoms types2.RestrictedDenoms
}

func (rk restrictedKeeper) GetRestrictedDenoms(sdk.Context) types2.RestrictedDenoms {
	return rk.RestrictedDenoms
}

func createAccount(ctx sdk.Context, ak auth.AccountKeeper, address, balance string) authexported.Account {
	acc := ak.NewAccountWithAddress(ctx, sdk.AccAddress([]byte(address)))
	acc.SetCoins(coins(balance))
	ak.SetAccount(ctx, acc)
	return acc
}

func coins(s string) sdk.Coins {
	coins, err := sdk.ParseCoins(s)
	if err != nil {
		panic(err)
	}
	return coins
}

func createCodec() *codec.Codec {
	cdc := codec.New()
	codec.RegisterCrypto(cdc)
	auth.RegisterCodec(cdc)

	return cdc
}
