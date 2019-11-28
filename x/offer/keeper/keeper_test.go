package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store"
	"github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/log"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/e-money/em-ledger/x/offer/types"

	abci "github.com/tendermint/tendermint/abci/types"
	dbm "github.com/tendermint/tm-db"
)

func TestBasicTrade(t *testing.T) {
	ctx, k, ak := createTestComponents(t)

	acc1 := createAccount(ctx, ak, "acc1", "5000eur")
	acc2 := createAccount(ctx, ak, "acc2", "7400usd")

	order := types.NewOrder(coin("100eur"), coin("120usd"), acc1.GetAddress())
	res := k.ProcessOrder(ctx, order)
	require.True(t, res.IsOK())

	order = types.NewOrder(coin("60usd"), coin("50eur"), acc2.GetAddress())
	res = k.ProcessOrder(ctx, order)
	require.True(t, res.IsOK())

	bal1 := ak.GetAccount(ctx, acc1.GetAddress()).GetCoins()
	bal2 := ak.GetAccount(ctx, acc2.GetAddress()).GetCoins()
	require.Len(t, bal1, 2)
	require.Len(t, bal2, 2)

	require.Equal(t, int64(4950), bal1.AmountOf("eur").Int64())
	require.Equal(t, int64(60), bal1.AmountOf("usd").Int64())

	require.Equal(t, int64(50), bal2.AmountOf("eur").Int64())
	require.Equal(t, int64(7340), bal2.AmountOf("usd").Int64())

	require.Len(t, k.instruments, 1)

	i := k.instruments[0]
	remainingOrder := i.Orders.Peek().(*types.Order)
	require.Equal(t, int64(50), remainingOrder.Remaining.Int64())
}

func TestInsufficientBalance1(t *testing.T) {
	// TODO This test will have to heavily modified or deleted once orders are removed when account balances drop below the order's source amount.
	ctx, k, ak := createTestComponents(t)

	acc1 := createAccount(ctx, ak, "acc1", "500eur")
	acc2 := createAccount(ctx, ak, "acc2", "740usd")

	order := types.NewOrder(coin("300eur"), coin("360usd"), acc1.GetAddress())
	k.ProcessOrder(ctx, order)

	// Modify account balance to be below order source
	acc1.SetCoins(coins("250eur"))
	k.ak.SetAccount(ctx, acc1)

	order = types.NewOrder(coin("360usd"), coin("300eur"), acc2.GetAddress())
	res := k.ProcessOrder(ctx, order)
	require.False(t, res.IsOK())

	acc1 = ak.GetAccount(ctx, acc1.GetAddress())
	acc2 = ak.GetAccount(ctx, acc2.GetAddress())
	require.Equal(t, coins("250eur"), acc1.GetCoins()) // Still holds the updated amount
	require.Equal(t, coins("740usd"), acc2.GetCoins())

	// TODO This is a very bad situation. The new, legit order is being blocked by the passive order not having the correct balance.

	order = types.NewOrder(coin("180usd"), coin("150eur"), acc2.GetAddress())
	res = k.ProcessOrder(ctx, order)
	require.True(t, res.IsOK())

	// Verify that the smaller order was executed
	acc1 = ak.GetAccount(ctx, acc1.GetAddress())
	acc2 = ak.GetAccount(ctx, acc2.GetAddress())
	require.Equal(t, coins("100eur,180usd"), acc1.GetCoins()) // Still holds the updated amount
	require.Equal(t, coins("560usd,150eur"), acc2.GetCoins())
}

func Test2(t *testing.T) {
	ctx, k, ak := createTestComponents(t)

	acc1 := createAccount(ctx, ak, "acc1", "100eur")
	acc2 := createAccount(ctx, ak, "acc2", "121usd")

	order := types.NewOrder(coin("100eur"), coin("120usd"), acc1.GetAddress())
	res := k.ProcessOrder(ctx, order)
	require.True(t, res.IsOK())

	order = types.NewOrder(coin("121usd"), coin("100eur"), acc2.GetAddress())
	res = k.ProcessOrder(ctx, order)
	require.True(t, res.IsOK())

	require.Len(t, k.instruments, 1)

	remainingOrder := k.instruments[0].Orders.Peek().(*types.Order)
	require.Equal(t, int64(1), remainingOrder.Remaining.Int64())
}

func Test3(t *testing.T) {
	ctx, k, ak := createTestComponents(t)

	acc1 := createAccount(ctx, ak, "acc1", "100eur")
	acc2 := createAccount(ctx, ak, "acc2", "120usd")

	order := types.NewOrder(coin("100eur"), coin("120usd"), acc1.GetAddress())
	k.ProcessOrder(ctx, order)

	for i := 0; i < 4; i++ {
		order = types.NewOrder(coin("30usd"), coin("25eur"), acc2.GetAddress())
		k.ProcessOrder(ctx, order)
	}

	// TODO There really ought to be zero instruments left now.
	i := k.instruments[0]
	require.Equal(t, 0, i.Orders.Len())
}

func createTestComponents(t *testing.T) (sdk.Context, Keeper, auth.AccountKeeper) {
	var (
		keyOffer   = sdk.NewKVStoreKey(types.ModuleName)
		authCapKey = sdk.NewKVStoreKey("authCapKey")
		keyParams  = sdk.NewKVStoreKey("params")
		tkeyParams = sdk.NewTransientStoreKey("transient_params")

		blacklistedAddrs = make(map[string]bool)
	)

	cdc := makeTestCodec()

	db := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(authCapKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyOffer, sdk.StoreTypeIAVL, db)

	err := ms.LoadLatestVersion()
	require.Nil(t, err)

	pk := params.NewKeeper(cdc, keyParams, tkeyParams, params.DefaultCodespace)

	ctx := sdk.NewContext(ms, abci.Header{ChainID: "test-chain"}, true, log.NewNopLogger())
	ak := auth.NewAccountKeeper(cdc, authCapKey, pk.Subspace(auth.DefaultParamspace), auth.ProtoBaseAccount)
	bk := bank.NewBaseKeeper(ak, pk.Subspace(bank.DefaultParamspace), bank.DefaultCodespace, blacklistedAddrs)

	k := NewKeeper(cdc, keyOffer, ak, bk)

	return ctx, k, ak
}

func makeTestCodec() (cdc *codec.Codec) {
	cdc = codec.New()

	auth.RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)

	return
}

func coin(s string) sdk.Coin {
	coin, err := sdk.ParseCoin(s)
	if err != nil {
		panic(err)
	}
	return coin
}

func coins(s string) sdk.Coins {
	coins, err := sdk.ParseCoins(s)
	if err != nil {
		panic(err)
	}
	return coins
}

func createAccount(ctx sdk.Context, ak auth.AccountKeeper, address, balance string) exported.Account {
	acc := ak.NewAccountWithAddress(ctx, sdk.AccAddress([]byte(address)))
	acc.SetCoins(coins(balance))
	ak.SetAccount(ctx, acc)
	return acc
}
