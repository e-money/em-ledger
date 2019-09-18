package keeper

import (
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/supply"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"testing"

	"emoney/x/liquidityprovider/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"

	abci "github.com/tendermint/tendermint/abci/types"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"
)

var (
	accAddr1 = sdk.AccAddress([]byte("account1"))

	defaultCredit = sdk.NewCoins(
		sdk.NewCoin("x2eur", sdk.NewIntWithDecimal(1000, 2)),
	)

	initialBalance = sdk.NewCoins(
		sdk.NewCoin("x2eur", sdk.NewIntWithDecimal(50, 2)),
		sdk.NewCoin("x0jpy", sdk.NewInt(250)),
	)
)

func TestCreateAndMint(t *testing.T) {
	ctx, ak, sk, _, keeper := createTestComponents(t, initialBalance)

	assert.Equal(t, initialBalance, sk.GetSupply(ctx).Total)

	acc := accAddr1
	account := ak.NewAccountWithAddress(ctx, acc)
	_ = account.SetCoins(initialBalance)
	ak.SetAccount(ctx, account)

	// Turn account into a LP
	keeper.CreateLiquidityProvider(ctx, acc, defaultCredit)
	account = ak.GetAccount(ctx, acc)

	assert.IsType(t, types.LiquidityProviderAccount{}, account)

	toMint := sdk.NewCoins(sdk.NewCoin("x2eur", sdk.NewIntWithDecimal(500, 2)))
	keeper.MintTokensFromCredit(ctx, acc, toMint)

	account = ak.GetAccount(ctx, acc)
	assert.Equal(t, initialBalance.Add(toMint), account.GetCoins())
	assert.Equal(t, initialBalance.Add(toMint), sk.GetSupply(ctx).Total)

	// Ensure that credit available has been correspondingly reduced
	lpAcc := keeper.getLiquidityProviderAccount(ctx, acc)
	assert.Equal(t, defaultCredit.Sub(toMint), lpAcc.Credit)
}

func TestMintTooMuch(t *testing.T) {
	ctx, ak, sk, _, keeper := createTestComponents(t, initialBalance)

	acc := accAddr1
	account := ak.NewAccountWithAddress(ctx, acc)
	_ = account.SetCoins(initialBalance)
	ak.SetAccount(ctx, account)

	// Turn account into a LP
	keeper.CreateLiquidityProvider(ctx, acc, defaultCredit)
	account = ak.GetAccount(ctx, acc)

	toMint := sdk.NewCoins(sdk.NewCoin("x2eur", sdk.NewIntWithDecimal(5000, 2)))
	keeper.MintTokensFromCredit(ctx, acc, toMint)

	account = ak.GetAccount(ctx, acc)
	assert.Equal(t, initialBalance, account.GetCoins())
	assert.Equal(t, initialBalance, sk.GetSupply(ctx).Total)

	// Ensure that credit of account has not been modified by failed attempt to mint.
	lpAcc := keeper.getLiquidityProviderAccount(ctx, acc)
	assert.Equal(t, defaultCredit, lpAcc.Credit)
}

func TestMintMultipleDenoms(t *testing.T) {
	ctx, ak, sk, _, keeper := createTestComponents(t, initialBalance)

	jpy := sdk.NewCoins(sdk.NewCoin("x0jpy", sdk.NewInt(1000000)))
	extendedCredit := defaultCredit.Add(jpy)

	acc := accAddr1
	account := ak.NewAccountWithAddress(ctx, acc)
	_ = account.SetCoins(initialBalance)
	ak.SetAccount(ctx, account)

	// Turn account into a LP
	keeper.CreateLiquidityProvider(ctx, acc, extendedCredit)
	account = ak.GetAccount(ctx, acc)

	toMint := sdk.NewCoins(
		sdk.NewCoin("x2eur", sdk.NewIntWithDecimal(500, 2)),
		sdk.NewCoin("x0jpy", sdk.NewInt(500000)),
	)

	keeper.MintTokensFromCredit(ctx, acc, toMint)
	account = ak.GetAccount(ctx, acc)
	assert.Equal(t, initialBalance.Add(toMint), account.GetCoins())
	assert.Equal(t, initialBalance.Add(toMint), sk.GetSupply(ctx).Total)

	// Ensure that credit available has been correspondingly reduced
	lpAcc := keeper.getLiquidityProviderAccount(ctx, acc)
	assert.Equal(t, extendedCredit.Sub(toMint), lpAcc.Credit)
}

func TestMintWithoutLPAccount(t *testing.T) {
	ctx, ak, sk, _, keeper := createTestComponents(t, initialBalance)

	acc := accAddr1
	account := ak.NewAccountWithAddress(ctx, acc)
	_ = account.SetCoins(initialBalance)
	ak.SetAccount(ctx, account)

	toMint := sdk.NewCoins(sdk.NewCoin("x2eur", sdk.NewIntWithDecimal(500, 2)))
	keeper.MintTokensFromCredit(ctx, acc, toMint)

	account = ak.GetAccount(ctx, acc)
	assert.IsType(t, &auth.BaseAccount{}, account)
	assert.Equal(t, initialBalance, sk.GetSupply(ctx).Total)
	assert.Equal(t, initialBalance, account.GetCoins())
}

func makeTestCodec() (cdc *codec.Codec) {
	cdc = codec.New()

	bank.RegisterCodec(cdc)
	auth.RegisterCodec(cdc)
	types.RegisterCodec(cdc)
	sdk.RegisterCodec(cdc)
	supply.RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)

	return
}

func createTestComponents(t *testing.T, initialSupply sdk.Coins) (sdk.Context, auth.AccountKeeper, supply.Keeper, bank.Keeper, Keeper) {
	cdc := makeTestCodec()

	var (
		keyAcc     = sdk.NewKVStoreKey(auth.StoreKey)
		keyParams  = sdk.NewKVStoreKey(params.StoreKey)
		keySupply  = sdk.NewKVStoreKey(supply.StoreKey)
		tkeyParams = sdk.NewTransientStoreKey(params.TStoreKey)
	)

	db := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(keyAcc, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keySupply, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyParams, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(tkeyParams, sdk.StoreTypeTransient, db)
	err := ms.LoadLatestVersion()
	require.Nil(t, err)

	ctx := sdk.NewContext(ms, abci.Header{ChainID: "supply-chain"}, true, log.NewNopLogger())

	maccPerms := map[string][]string{
		types.ModuleName: {supply.Minter},
	}

	pk := params.NewKeeper(cdc, keyParams, tkeyParams, params.DefaultCodespace)
	ak := auth.NewAccountKeeper(cdc, keyAcc, pk.Subspace(auth.DefaultParamspace), auth.ProtoBaseAccount)
	bk := bank.NewBaseKeeper(ak, pk.Subspace(bank.DefaultParamspace), bank.DefaultCodespace)
	sk := supply.NewKeeper(cdc, keySupply, ak, bk, supply.DefaultCodespace, maccPerms)

	// Empty supply
	sk.SetSupply(ctx, supply.NewSupply(initialSupply))

	keeper := NewKeeper(ak, sk)

	return ctx, ak, sk, bk, keeper
}
