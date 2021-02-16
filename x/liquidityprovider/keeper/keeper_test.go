// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	"github.com/cosmos/cosmos-sdk/std"
	"github.com/cosmos/cosmos-sdk/store"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
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
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/e-money/em-ledger/x/liquidityprovider/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/log"
	tmrand "github.com/tendermint/tendermint/libs/rand"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	dbm "github.com/tendermint/tm-db"
	"testing"
)

var (
	accAddr1 = sdk.AccAddress(tmrand.Bytes(sdk.AddrLen))

	defaultMintable = sdk.NewCoins(
		sdk.NewCoin("eeur", sdk.NewIntWithDecimal(1000, 2)),
	)

	initialBalance = sdk.NewCoins(
		sdk.NewCoin("eeur", sdk.NewIntWithDecimal(50, 2)),
		sdk.NewCoin("ejpy", sdk.NewInt(250)),
	)
)

func TestCreateAndMint(t *testing.T) {
	ctx, ak, bk, keeper := createTestComponents(t, initialBalance)

	assert.Equal(t, initialBalance, bk.GetSupply(ctx).GetTotal())

	acc := accAddr1
	account := ak.NewAccountWithAddress(ctx, acc)
	ak.SetAccount(ctx, account)
	err := bk.SetBalances(ctx, acc, initialBalance)
	require.NoError(t, err)

	// Turn account into a LP
	_, err = keeper.CreateLiquidityProvider(ctx, acc, defaultMintable)
	require.NoError(t, err)
	account = ak.GetAccount(ctx, acc)

	assert.IsType(t, &types.LiquidityProviderAccount{}, account)

	toMint := sdk.NewCoins(sdk.NewCoin("eeur", sdk.NewIntWithDecimal(500, 2)))
	keeper.MintTokens(ctx, acc, toMint)

	balances := bk.GetAllBalances(ctx, acc)
	assert.Equal(t, initialBalance.Add(toMint...), balances)
	assert.Equal(t, initialBalance.Add(toMint...), bk.GetSupply(ctx).GetTotal())

	// Ensure that mintable amount available has been correspondingly reduced
	lpAcc := keeper.GetLiquidityProviderAccount(ctx, acc)
	assert.Equal(t, defaultMintable.Sub(toMint), lpAcc.Mintable)

	allLPs := keeper.GetAllLiquidityProviderAccounts(ctx)
	require.Len(t, allLPs, 1)
}

func TestMintTooMuch(t *testing.T) {
	ctx, ak, bk, keeper := createTestComponents(t, initialBalance)

	acc := accAddr1
	account := ak.NewAccountWithAddress(ctx, acc)
	ak.SetAccount(ctx, account)
	err := bk.SetBalances(ctx, acc, initialBalance)
	require.NoError(t, err)

	// Turn account into a LP
	keeper.CreateLiquidityProvider(ctx, acc, defaultMintable)
	account = ak.GetAccount(ctx, acc)

	toMint := sdk.NewCoins(sdk.NewCoin("eeur", sdk.NewIntWithDecimal(5000, 2)))
	keeper.MintTokens(ctx, acc, toMint)

	balances := bk.GetAllBalances(ctx, acc)
	assert.Equal(t, initialBalance, balances)
	assert.Equal(t, initialBalance, bk.GetSupply(ctx).GetTotal())

	// Ensure that the mintable amount of the account has not been modified by failed attempt to mint.
	lpAcc := keeper.GetLiquidityProviderAccount(ctx, acc)
	assert.Equal(t, defaultMintable, lpAcc.Mintable)
}

func TestMintMultipleDenoms(t *testing.T) {
	ctx, ak, bk, keeper := createTestComponents(t, initialBalance)

	jpy := sdk.NewCoins(sdk.NewCoin("ejpy", sdk.NewInt(1000000)))
	extendedMintable := defaultMintable.Add(jpy...)

	acc := accAddr1
	account := ak.NewAccountWithAddress(ctx, acc)
	ak.SetAccount(ctx, account)
	err := bk.SetBalances(ctx, acc, initialBalance)
	require.NoError(t, err)

	// Turn account into a LP
	keeper.CreateLiquidityProvider(ctx, acc, extendedMintable)
	account = ak.GetAccount(ctx, acc)

	toMint := sdk.NewCoins(
		sdk.NewCoin("eeur", sdk.NewIntWithDecimal(500, 2)),
		sdk.NewCoin("ejpy", sdk.NewInt(500000)),
	)

	keeper.MintTokens(ctx, acc, toMint)
	balances := bk.GetAllBalances(ctx, acc)
	assert.Equal(t, initialBalance.Add(toMint...), balances)
	assert.Equal(t, initialBalance.Add(toMint...), bk.GetSupply(ctx).GetTotal())

	// Ensure that mintable amount available has been correspondingly reduced
	lpAcc := keeper.GetLiquidityProviderAccount(ctx, acc)
	assert.Equal(t, extendedMintable.Sub(toMint), lpAcc.Mintable)
}

func TestMintWithoutLPAccount(t *testing.T) {
	ctx, ak, bk, keeper := createTestComponents(t, initialBalance)

	acc := accAddr1
	account := ak.NewAccountWithAddress(ctx, acc)
	ak.SetAccount(ctx, account)
	err := bk.SetBalances(ctx, acc, initialBalance)
	require.NoError(t, err)

	toMint := sdk.NewCoins(sdk.NewCoin("eeur", sdk.NewIntWithDecimal(500, 2)))
	keeper.MintTokens(ctx, acc, toMint)

	account = ak.GetAccount(ctx, acc)
	assert.IsType(t, &authtypes.BaseAccount{}, account)
	assert.Equal(t, initialBalance, bk.GetSupply(ctx).GetTotal())
	balances := bk.GetAllBalances(ctx, acc)
	assert.Equal(t, initialBalance, balances)

	allLPs := keeper.GetAllLiquidityProviderAccounts(ctx)
	require.Empty(t, allLPs)
}

func TestCreateAndRevoke(t *testing.T) {
	ctx, ak, bk, keeper := createTestComponents(t, initialBalance)
	acc := accAddr1

	account := ak.NewAccountWithAddress(ctx, acc)
	ak.SetAccount(ctx, account)
	err := bk.SetBalances(ctx, acc, initialBalance)
	require.NoError(t, err)

	// Turn account into a LP
	keeper.CreateLiquidityProvider(ctx, acc, defaultMintable)
	account = ak.GetAccount(ctx, acc)

	assert.IsType(t, &types.LiquidityProviderAccount{}, account)

	keeper.RevokeLiquidityProviderAccount(ctx, account)
	account = ak.GetAccount(ctx, acc)
	assert.IsType(t, &authtypes.BaseAccount{}, account)
}

func TestLiquidityProviderIO(t *testing.T) {
	ctx, ak, _, keeper := createTestComponents(t, initialBalance)
	_, pub, acc := testdata.KeyTestPubAddr()

	account := ak.NewAccountWithAddress(ctx, acc)
	err := account.SetPubKey(pub)
	require.NoError(t, err)
	ak.SetAccount(ctx, account)

	// when serialize
	_, err = keeper.CreateLiquidityProvider(ctx, acc, defaultMintable)
	require.NoError(t, err)

	// then deserialize
	account = ak.GetAccount(ctx, acc)
	require.Equal(t, pub, account.GetPubKey())

	// and when updated
	_, otherPub, _ := testdata.KeyTestPubAddr()
	account.SetPubKey(otherPub)
	ak.SetAccount(ctx, account)

	// then
	account = ak.GetAccount(ctx, acc)
	require.Equal(t, otherPub, account.GetPubKey())
}

func TestAccountNotFound(t *testing.T) {
	ctx, ak, _, keeper := createTestComponents(t, initialBalance)

	acc := accAddr1
	keeper.CreateLiquidityProvider(ctx, acc, defaultMintable)
	assert.Nil(t, ak.GetAccount(ctx, acc))
}

func createTestComponents(t *testing.T, initialSupply sdk.Coins) (sdk.Context, authkeeper.AccountKeeper, bankkeeper.Keeper, Keeper) {
	t.Helper()
	encConfig := MakeTestEncodingConfig()
	var (
		keyInflation = sdk.NewKVStoreKey(types.ModuleName)
		bankKey      = sdk.NewKVStoreKey(banktypes.ModuleName)
		authCapKey   = sdk.NewKVStoreKey("authCapKey")
		keyParams    = sdk.NewKVStoreKey("params")
		stakingKey   = sdk.NewKVStoreKey("staking")
		authKey      = sdk.NewKVStoreKey(authtypes.StoreKey)
		tkeyParams   = sdk.NewTransientStoreKey("transient_params")

		blockedAddrs = make(map[string]bool)
	)

	db := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(keyInflation, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(authCapKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(stakingKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(authKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(bankKey, sdk.StoreTypeIAVL, db)

	err := ms.LoadLatestVersion()
	require.NoError(t, err)

	maccPerms := map[string][]string{
		types.ModuleName:               {authtypes.Minter},
		authtypes.FeeCollectorName:     nil,
		"buyback":                      {authtypes.Burner},
		stakingtypes.NotBondedPoolName: {authtypes.Burner, authtypes.Staking},
		stakingtypes.BondedPoolName:    {authtypes.Burner, authtypes.Staking},
	}

	pk := paramskeeper.NewKeeper(encConfig.Marshaler, encConfig.Amino, keyParams, tkeyParams)

	ctx := sdk.NewContext(ms, tmproto.Header{ChainID: "test-chain"}, true, log.NewNopLogger())

	ak := authkeeper.NewAccountKeeper(
		encConfig.Marshaler, authCapKey, pk.Subspace(authtypes.ModuleName), authtypes.ProtoBaseAccount, maccPerms,
	)

	bk := bankkeeper.NewBaseKeeper(
		encConfig.Marshaler, bankKey, ak, pk.Subspace(banktypes.ModuleName), blockedAddrs,
	)

	bk.SetSupply(ctx, banktypes.NewSupply(initialSupply))

	keeper := NewKeeper(ak, bk)

	return ctx, ak, bk, keeper
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
	types.RegisterLegacyAminoCodec(encodingConfig.Amino)
	types.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	return encodingConfig
}
