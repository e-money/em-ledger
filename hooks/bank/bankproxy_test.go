// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package bank

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	"github.com/cosmos/cosmos-sdk/std"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	emauthtypes "github.com/e-money/em-ledger/x/authority/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/libs/rand"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	dbm "github.com/tendermint/tm-db"
	"testing"
)

func TestProxySendCoins(t *testing.T) {
	ctx, ak, bk := createTestComponents(t)

	var (
		acc1 = createAccount(ctx, ak, bk, randomAddress(), "150000gbp, 150000usd, 150000sek")
		acc2 = createAccount(ctx, ak, bk, randomAddress(), "150000gbp, 150000usd, 150000sek")
		dest = randomAddress()
	)

	bk.rk = restrictedKeeper{
		RestrictedDenoms: emauthtypes.RestrictedDenoms{Denoms: []emauthtypes.RestrictedDenom{
			{"gbp", []string{}},
			{"usd", []string{acc1.GetAddress().String()}},
		},
		}}

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

	for i, d := range testdata {
		t.Run(fmt.Sprintf("test case %d", i), func(t *testing.T) {
			c := fmt.Sprintf("1000%s", d.denom)
			err := bk.SendCoins(ctx, d.acc, dest, coins(c))
			if d.valid {
				require.NoError(t, err)
			} else {
				require.True(t, ErrRestrictedDenomination.Is(err), "Actual error \"%s\" (%T)", err.Error(), err)
			}
		})
	}
}

func TestInputOutputCoins(t *testing.T) {
	ctx, ak, bk := createTestComponents(t)

	var (
		acc1 = createAccount(ctx, ak, bk, randomAddress(), "150000gbp, 150000usd, 150000sek")
		acc2 = createAccount(ctx, ak, bk, randomAddress(), "150000gbp, 150000usd, 150000sek")
		acc3 = createAccount(ctx, ak, bk, randomAddress(), "")
	)

	// For simplicity's sake, inputoutput will reject any transaction that includes restricted denominations.

	bk.rk = restrictedKeeper{
		RestrictedDenoms: emauthtypes.RestrictedDenoms{Denoms: []emauthtypes.RestrictedDenom{
			{"gbp", []string{}},
			{"usd", []string{acc1.GetAddress().String()}},
		},
		}}

	var testdata = []struct {
		inputs  []banktypes.Input
		outputs []banktypes.Output
		valid   bool
	}{
		{[]banktypes.Input{}, []banktypes.Output{}, true},
		{
			inputs: []banktypes.Input{
				{acc1.GetAddress().String(), coins("1000sek")},
			},
			outputs: []banktypes.Output{
				{acc2.GetAddress().String(), coins("500sek")},
				{acc3.GetAddress().String(), coins("500sek")},
			},
			valid: true,
		},
		{
			inputs: []banktypes.Input{
				{acc1.GetAddress().String(), coins("500sek, 1000gbp")},
			},
			outputs: []banktypes.Output{
				{acc2.GetAddress().String(), coins("500sek, 500gbp")},
				{acc3.GetAddress().String(), coins("500gbp")},
			},
			valid: false,
		},
		{
			inputs: []banktypes.Input{
				{acc1.GetAddress().String(), coins("1000usd")},
			},
			outputs: []banktypes.Output{
				{acc2.GetAddress().String(), coins("1000usd")},
			},
			valid: false,
		},
	}

	for i, d := range testdata {
		t.Run(fmt.Sprintf("test case %d", i), func(t *testing.T) {
			err := bk.InputOutputCoins(ctx, d.inputs, d.outputs)
			if d.valid {
				require.NoError(t, err)
			} else {
				require.True(t, ErrRestrictedDenomination.Is(err), "Actual error \"%s\" (%T)", err.Error(), err)
			}
		})
	}

	fmt.Println(bk.GetAllBalances(ctx, acc3.GetAddress()))
}

func TestDeduplicate(t *testing.T) {
	var (
		addr1 = bytes.Repeat([]byte{0x1}, sdk.AddrLen)
		addr2 = bytes.Repeat([]byte{0x2}, sdk.AddrLen)
		addr3 = bytes.Repeat([]byte{0x3}, sdk.AddrLen)
	)
	specs := map[string]struct {
		src []sdk.AccAddress
		exp []sdk.AccAddress
	}{
		"without duplicates": {
			src: []sdk.AccAddress{addr1, addr2, addr3},
			exp: []sdk.AccAddress{addr1, addr2, addr3},
		},
		"duplicates": {
			src: []sdk.AccAddress{addr1, addr1, addr2, addr3},
			exp: []sdk.AccAddress{addr1, addr2, addr3},
		},
		"more duplicates": {
			src: []sdk.AccAddress{addr1, addr2, addr3, addr1, addr2, addr3},
			exp: []sdk.AccAddress{addr1, addr2, addr3},
		},
		"empty": {
			src: []sdk.AccAddress{},
			exp: []sdk.AccAddress{},
		},
		"nil": {
			exp: []sdk.AccAddress{},
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			got := deduplicate(spec.src)
			assert.Equal(t, spec.exp, got)
		})
	}
}

func TestSendCoinsBalanceUpdateNotification(t *testing.T) {
	var (
		ctx   sdk.Context
		addr1 = randomAddress()
		addr2 = randomAddress()
	)

	specs := map[string]struct {
		srcSender          sdk.AccAddress
		srcRecvr           sdk.AccAddress
		listenerCount      int
		nestedKeeperResult error

		expAddr []sdk.AccAddress
		expErr  bool
	}{
		"one listener called": {
			srcSender:     addr1,
			srcRecvr:      addr2,
			listenerCount: 1,
			expAddr:       []sdk.AccAddress{addr1, addr2},
		},
		"multiple listener called": {
			srcSender:     addr1,
			srcRecvr:      addr2,
			listenerCount: 2,
			expAddr:       []sdk.AccAddress{addr1, addr2},
		},
		"no listener called on error": {
			srcSender:          addr1,
			srcRecvr:           addr2,
			listenerCount:      2,
			nestedKeeperResult: errors.New("test, ignore"),
			expErr:             true,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			receivedAddr := make([][][]sdk.AccAddress, spec.listenerCount)
			newListener := func(listenerNb int) func(sdk.Context, []sdk.AccAddress) {
				return func(_ sdk.Context, addrs []sdk.AccAddress) {
					receivedAddr[listenerNb] = append(receivedAddr[listenerNb], addrs)
				}
			}

			nestedBk := senderBankKeeperMock{
				SendCoinsFn: func(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error {
					return spec.nestedKeeperResult
				},
			}
			wrappedBankKeeper := Wrap(nestedBk, allDenomsAllowed)

			// register listeners
			for i := 0; i < spec.listenerCount; i++ {
				wrappedBankKeeper.AddBalanceListener(newListener(i))
			}

			// when
			gotErr := wrappedBankKeeper.SendCoins(ctx, spec.srcSender, spec.srcRecvr, coins("1token"))

			// then
			if spec.expErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
			for i := 0; i < spec.listenerCount; i++ {
				require.Len(t, receivedAddr[i], 1)
				assert.Equal(t, spec.expAddr, receivedAddr[i][0])
			}
		})
	}
}

func TestInputOutputCoinsBalanceUpdateNotification(t *testing.T) {
	var (
		ctx   sdk.Context
		addr1 = randomAddress()
		addr2 = randomAddress()
	)

	specs := map[string]struct {
		srcInput           []banktypes.Input
		srcOutput          []banktypes.Output
		listenerCount      int
		nestedKeeperResult error

		expAddr []sdk.AccAddress
		expErr  bool
	}{
		"one listener called": {
			srcInput:      []banktypes.Input{{Address: addr1.String()}},
			srcOutput:     []banktypes.Output{{Address: addr2.String()}},
			listenerCount: 1,
			expAddr:       []sdk.AccAddress{addr1, addr2},
		},
		"multiple listener called": {
			srcInput:      []banktypes.Input{{Address: addr1.String()}},
			srcOutput:     []banktypes.Output{{Address: addr2.String()}},
			listenerCount: 2,
			expAddr:       []sdk.AccAddress{addr1, addr2},
		},
		"no listener called on error": {
			srcInput:           []banktypes.Input{{Address: addr1.String()}},
			srcOutput:          []banktypes.Output{{Address: addr2.String()}},
			listenerCount:      2,
			nestedKeeperResult: errors.New("test, ignore"),
			expErr:             true,
		},
		"deduplicated": {
			srcInput:      []banktypes.Input{{Address: addr1.String()}, {Address: addr1.String()}, {Address: addr2.String()}},
			srcOutput:     []banktypes.Output{{Address: addr2.String()}, {Address: addr2.String()}, {Address: addr1.String()}},
			listenerCount: 2,
			expAddr:       []sdk.AccAddress{addr1, addr2},
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			receivedAddr := make([][][]sdk.AccAddress, spec.listenerCount)
			newListener := func(listenerNb int) func(sdk.Context, []sdk.AccAddress) {
				return func(_ sdk.Context, addrs []sdk.AccAddress) {
					receivedAddr[listenerNb] = append(receivedAddr[listenerNb], addrs)
				}
			}

			nestedBk := senderBankKeeperMock{
				InputOutputCoinsFn: func(ctx sdk.Context, in []banktypes.Input, out []banktypes.Output) error {
					return spec.nestedKeeperResult
				},
			}
			wrappedBankKeeper := Wrap(nestedBk, allDenomsAllowed)

			// register listeners
			for i := 0; i < spec.listenerCount; i++ {
				wrappedBankKeeper.AddBalanceListener(newListener(i))
			}

			// when
			gotErr := wrappedBankKeeper.InputOutputCoins(ctx, spec.srcInput, spec.srcOutput)

			// then
			if spec.expErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
			for i := 0; i < spec.listenerCount; i++ {
				require.Len(t, receivedAddr[i], 1)
				assert.Equal(t, spec.expAddr, receivedAddr[i][0])
			}
		})
	}
}

var allDenomsAllowed = RestrictedKeeperFunc(func(ctx sdk.Context) emauthtypes.RestrictedDenoms {
	return emauthtypes.RestrictedDenoms{} // allow all
})

type senderBankKeeperMock struct {
	bankkeeper.Keeper
	InputOutputCoinsFn func(ctx sdk.Context, inputs []banktypes.Input, outputs []banktypes.Output) error
	SendCoinsFn        func(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error
}

func (pk senderBankKeeperMock) SendCoins(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error {
	if pk.SendCoinsFn == nil {
		panic("unexpected call")
	}
	return pk.SendCoinsFn(ctx, fromAddr, toAddr, amt)
}

func (pk senderBankKeeperMock) InputOutputCoins(ctx sdk.Context, inputs []banktypes.Input, outputs []banktypes.Output) error {
	if pk.InputOutputCoinsFn == nil {
		panic("unexpected call")
	}
	return pk.InputOutputCoinsFn(ctx, inputs, outputs)
}

func randomAddress() sdk.AccAddress {
	return rand.Bytes(sdk.AddrLen)
}

func createTestComponents(t *testing.T) (sdk.Context, authkeeper.AccountKeeper, *ProxyKeeper) {
	t.Helper()
	encConfig := MakeTestEncodingConfig()
	var (
		bankKey    = sdk.NewKVStoreKey(banktypes.ModuleName)
		authCapKey = sdk.NewKVStoreKey("authCapKey")
		keyParams  = sdk.NewKVStoreKey("params")
		tkeyParams = sdk.NewTransientStoreKey("transient_params")

		blacklistedAddrs = make(map[string]bool)
	)

	db := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(authCapKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(bankKey, sdk.StoreTypeIAVL, db)

	err := ms.LoadLatestVersion()
	require.NoError(t, err)

	maccPerms := map[string][]string{}

	pk := paramskeeper.NewKeeper(encConfig.Marshaler, encConfig.Amino, keyParams, tkeyParams)

	ctx := sdk.NewContext(ms, tmproto.Header{ChainID: "test-chain"}, true, log.NewNopLogger())

	accountKeeper := authkeeper.NewAccountKeeper(
		encConfig.Marshaler, authCapKey, pk.Subspace(authtypes.ModuleName), authtypes.ProtoBaseAccount, maccPerms,
	)

	bankKeeper := bankkeeper.NewBaseKeeper(
		encConfig.Marshaler, bankKey, accountKeeper, pk.Subspace(banktypes.ModuleName), blacklistedAddrs,
	)

	wrappedBK := Wrap(bankKeeper, restrictedKeeper{})
	return ctx, accountKeeper, wrappedBK
}

type restrictedKeeper struct {
	RestrictedDenoms emauthtypes.RestrictedDenoms
}

func (rk restrictedKeeper) GetRestrictedDenoms(sdk.Context) emauthtypes.RestrictedDenoms {
	return rk.RestrictedDenoms
}

func createAccount(ctx sdk.Context, ak authkeeper.AccountKeeper, bk bankkeeper.SendKeeper, address sdk.AccAddress, balance string) authtypes.AccountI {
	acc := ak.NewAccountWithAddress(ctx, address)
	if err := bk.SetBalances(ctx, address, coins(balance)); err != nil {
		panic(err)
	}
	ak.SetAccount(ctx, acc)
	return acc
}

func coins(s string) sdk.Coins {
	coins, err := sdk.ParseCoinsNormalized(s)
	if err != nil {
		panic(err)
	}
	return coins
}

func MakeTestEncodingConfig() simappparams.EncodingConfig {
	cdc := codec.NewLegacyAmino()
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	marshaler := codec.NewProtoCodec(interfaceRegistry)

	encodingConfig := simappparams.EncodingConfig{
		InterfaceRegistry: interfaceRegistry,
		Marshaler:         marshaler,
		TxConfig:          tx.NewTxConfig(marshaler, tx.DefaultSignModes),
		Amino:             cdc,
	}

	std.RegisterLegacyAminoCodec(encodingConfig.Amino)
	std.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	ModuleBasics := module.NewBasicManager(
		bank.AppModuleBasic{},
		auth.AppModuleBasic{},
	)

	ModuleBasics.RegisterLegacyAminoCodec(encodingConfig.Amino)
	ModuleBasics.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	return encodingConfig
}
