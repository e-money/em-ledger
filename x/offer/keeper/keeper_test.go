// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

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
	cmn "github.com/tendermint/tendermint/libs/common"
	dbm "github.com/tendermint/tm-db"
)

func TestBasicTrade(t *testing.T) {
	ctx, k, ak := createTestComponents(t)

	acc1 := createAccount(ctx, ak, "acc1", "5000eur")
	acc2 := createAccount(ctx, ak, "acc2", "7400usd")

	order := types.NewOrder(coin("100eur"), coin("120usd"), acc1.GetAddress(), cid())
	res := k.NewOrderSingle(ctx, order)
	require.True(t, res.IsOK())

	order = types.NewOrder(coin("60usd"), coin("50eur"), acc2.GetAddress(), cid())
	res = k.NewOrderSingle(ctx, order)
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
	remainingOrder := i.Orders.LeftKey().(*types.Order)
	require.Equal(t, int64(50), remainingOrder.SourceRemaining.Int64())
}

func TestInsufficientBalance1(t *testing.T) {
	// TODO This test will have to heavily modified or deleted once orders are removed when account balances drop below the order's source amount.
	ctx, k, ak := createTestComponents(t)

	acc1 := createAccount(ctx, ak, "acc1", "500eur")
	acc2 := createAccount(ctx, ak, "acc2", "740usd")

	order := types.NewOrder(coin("300eur"), coin("360usd"), acc1.GetAddress(), cid())
	k.NewOrderSingle(ctx, order)

	// Modify account balance to be below order source
	acc1.SetCoins(coins("250eur"))
	k.ak.SetAccount(ctx, acc1)

	order = types.NewOrder(coin("360usd"), coin("300eur"), acc2.GetAddress(), cid())
	res := k.NewOrderSingle(ctx, order)
	require.False(t, res.IsOK())

	acc1 = ak.GetAccount(ctx, acc1.GetAddress())
	acc2 = ak.GetAccount(ctx, acc2.GetAddress())
	require.Equal(t, coins("250eur"), acc1.GetCoins()) // Still holds the updated amount
	require.Equal(t, coins("740usd"), acc2.GetCoins())

	// TODO This is a very bad situation. The new, legit order is being blocked by the passive order not having the correct balance.

	order = types.NewOrder(coin("180usd"), coin("150eur"), acc2.GetAddress(), cid())
	res = k.NewOrderSingle(ctx, order)
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

	order := types.NewOrder(coin("100eur"), coin("120usd"), acc1.GetAddress(), cid())
	res := k.NewOrderSingle(ctx, order)
	require.True(t, res.IsOK())

	order = types.NewOrder(coin("121usd"), coin("100eur"), acc2.GetAddress(), cid())
	res = k.NewOrderSingle(ctx, order)
	require.True(t, res.IsOK())

	require.Len(t, k.instruments, 1)

	remainingOrder := k.instruments[0].Orders.LeftKey().(*types.Order)
	require.Equal(t, int64(1), remainingOrder.SourceRemaining.Int64())
}

func Test3(t *testing.T) {
	ctx, k, ak := createTestComponents(t)

	acc1 := createAccount(ctx, ak, "acc1", "100eur")
	acc2 := createAccount(ctx, ak, "acc2", "120usd")

	order := types.NewOrder(coin("100eur"), coin("120usd"), acc1.GetAddress(), cid())
	k.NewOrderSingle(ctx, order)

	for i := 0; i < 4; i++ {
		order = types.NewOrder(coin("30usd"), coin("25eur"), acc2.GetAddress(), cid())
		k.NewOrderSingle(ctx, order)
	}

	require.Len(t, k.instruments, 0)
	acc1 = ak.GetAccount(ctx, acc1.GetAddress())
	acc2 = ak.GetAccount(ctx, acc2.GetAddress())
	require.Equal(t, coins("120usd"), acc1.GetCoins())
	require.Equal(t, coins("100eur"), acc2.GetCoins())
}

func TestDeleteOrder(t *testing.T) {
	ctx, k, ak := createTestComponents(t)
	acc1 := createAccount(ctx, ak, "acc1", "100eur")

	cid := cid()

	order1 := types.NewOrder(coin("100eur"), coin("120usd"), acc1.GetAddress(), cid)
	res := k.NewOrderSingle(ctx, order1)
	require.True(t, res.IsOK())

	order2 := types.NewOrder(coin("100eur"), coin("77chf"), acc1.GetAddress(), cid)
	res = k.NewOrderSingle(ctx, order2)
	require.False(t, res.IsOK()) // Verify that client order ids cannot be duplicated.

	require.Len(t, k.instruments, 1) // Ensure that the eur->chf pair was not added.

	k.deleteOrder(order1)
	require.Len(t, k.instruments, 0) // Removing the only eur->usd order should have removed instrument
}

func TestGetOrdersByOwnerAndCancel(t *testing.T) {
	ctx, k, ak := createTestComponents(t)
	acc1 := createAccount(ctx, ak, "acc1", "100eur")
	acc2 := createAccount(ctx, ak, "acc2", "120usd")

	for i := 0; i < 5; i++ {
		order := types.NewOrder(coin("5eur"), coin("12usd"), acc1.GetAddress(), cid())
		res := k.NewOrderSingle(ctx, order)
		require.True(t, res.IsOK())
	}

	for i := 0; i < 5; i++ {
		order := types.NewOrder(coin("7usd"), coin("3chf"), acc2.GetAddress(), cid())
		res := k.NewOrderSingle(ctx, order)
		require.True(t, res.IsOK(), res.Log)
	}

	allOrders1 := k.GetOrdersByOwner(acc1.GetAddress())
	require.Len(t, allOrders1, 5)

	{
		order := types.NewOrder(coin("12usd"), coin("5eur"), acc2.GetAddress(), cid())
		res := k.NewOrderSingle(ctx, order)
		require.True(t, res.IsOK(), res.Log)
	}

	allOrders2 := k.GetOrdersByOwner(acc1.GetAddress())
	require.Len(t, allOrders2, 4)

	cid := allOrders2[2].ClientOrderID
	require.True(t, k.CancelOrder(ctx, acc1.GetAddress(), cid).IsOK())
	require.False(t, k.CancelOrder(ctx, acc1.GetAddress(), cid).IsOK())

	allOrders3 := k.GetOrdersByOwner(acc1.GetAddress())
	require.Len(t, allOrders3, 3)
}

func TestCancelOrders1(t *testing.T) {
	// Cancel a non-existing order by an account with no orders in the system.
	ctx, k, ak := createTestComponents(t)
	acc := createAccount(ctx, ak, "acc1", "100eur")

	res := k.CancelOrder(ctx, acc.GetAddress(), "abcde")
	require.False(t, res.IsOK())
}

func TestCancelReplaceOrder(t *testing.T) {
	ctx, k, ak := createTestComponents(t)
	acc1 := createAccount(ctx, ak, "acc1", "20000eur")
	acc2 := createAccount(ctx, ak, "acc2", "45000usd")

	order1cid := cid()
	order1 := types.NewOrder(coin("500eur"), coin("1200usd"), acc1.GetAddress(), order1cid)
	res := k.NewOrderSingle(ctx, order1)
	require.True(t, res.IsOK())

	order2cid := cid()
	order2 := types.NewOrder(coin("5000eur"), coin("17000usd"), acc1.GetAddress(), order2cid)
	res = k.CancelReplaceOrder(ctx, order2, order1cid)
	require.True(t, res.IsOK())

	{
		orders := k.GetOrdersByOwner(acc1.GetAddress())
		require.Len(t, orders, 1)
		require.Equal(t, order2cid, orders[0].ClientOrderID)
		require.Equal(t, coin("5000eur"), orders[0].Source)
		require.Equal(t, coin("17000usd"), orders[0].Destination)
		require.Equal(t, sdk.NewInt(5000), orders[0].SourceRemaining)
	}

	order3 := types.NewOrder(coin("500chf"), coin("1700usd"), acc1.GetAddress(), cid())
	// Wrong client order id for previous order submitted.
	res = k.CancelReplaceOrder(ctx, order3, order1cid)
	require.Equal(t, types.CodeClientOrderIdNotFound, res.Code)

	// Changing instrument of order
	res = k.CancelReplaceOrder(ctx, order3, order2cid)
	require.Equal(t, types.CodeOrderInstrumentChanged, res.Code)

	res = k.NewOrderSingle(ctx,
		types.NewOrder(coin("2600usd"), coin("300eur"), acc2.GetAddress(), cid()),
	)
	require.True(t, res.IsOK())

	acc1 = ak.GetAccount(ctx, acc1.GetAddress())
	acc2 = ak.GetAccount(ctx, acc2.GetAddress())

	require.Equal(t, int64(765), acc2.GetCoins().AmountOf("eur").Int64())
	require.Equal(t, int64(2600), acc1.GetCoins().AmountOf("usd").Int64())

	//fmt.Println("acc1", acc1.GetCoins())
	//fmt.Println("acc2", acc2.GetCoins())
	//fmt.Println("Total supply:", acc1.GetCoins().Add(acc2.GetCoins()))

	filled := sdk.ZeroInt()
	{
		orders := k.GetOrdersByOwner(acc1.GetAddress())
		require.Len(t, orders, 1)
		filled = orders[0].Source.Amount.Sub(orders[0].SourceRemaining)
	}

	// CancelReplace and verify that previously filled amount is subtracted from the resulting order
	order4cid := cid()
	order4 := types.NewOrder(coin("10000eur"), coin("35050usd"), acc1.GetAddress(), order4cid)
	res = k.CancelReplaceOrder(ctx, order4, order2cid)
	require.True(t, res.IsOK(), res.Log)

	{
		orders := k.GetOrdersByOwner(acc1.GetAddress())
		require.Len(t, orders, 1)
		require.Equal(t, order4cid, orders[0].ClientOrderID)
		require.Equal(t, coin("10000eur"), orders[0].Source)
		require.Equal(t, coin("35050usd"), orders[0].Destination)
		require.Equal(t, sdk.NewInt(10000).Sub(filled), orders[0].SourceRemaining)
	}
}

func TestOrdersChangeWithAccountBalance(t *testing.T) {
	ctx, k, ak := createTestComponents(t)
	acc := createAccount(ctx, ak, "acc1", "15000eur")

	order := types.NewOrder(coin("10000eur"), coin("1000usd"), acc.GetAddress(), cid())
	res := k.NewOrderSingle(ctx, order)
	require.True(t, res.IsOK())

	{
		// Partially fill the order above
		acc2 := createAccount(ctx, ak, "acc2", "900000usd")
		order2 := types.NewOrder(coin("400usd"), coin("4000eur"), acc2.GetAddress(), cid())
		res = k.NewOrderSingle(ctx, order2)
		require.True(t, res.IsOK())

		//fmt.Println(ak.GetAccount(ctx, acc2.GetAddress()))
	}

	acc.SetCoins(coins("3000eur"))
	k.accountChanged(ctx, acc)

	// Seller's account balance drops, remaining should be adjusted accordingly.
	orders := k.GetOrdersByOwner(acc.GetAddress())
	require.Len(t, orders, 1)
	require.Equal(t, coin("10000eur"), orders[0].Source)
	require.Equal(t, "3000", orders[0].SourceRemaining.String())
	require.Equal(t, "4000", orders[0].SourceFilled.String())

	// Seller's account balance is restored. Order should be adjusted, but take into consideration that the order has already been partially filled.
	acc.SetCoins(coins("15000eur"))
	k.accountChanged(ctx, acc)

	orders = k.GetOrdersByOwner(acc.GetAddress())
	require.Equal(t, "6000", orders[0].SourceRemaining.String())
	require.Equal(t, "4000", orders[0].SourceFilled.String())

	// Account balance dips below original sales amount, but can still fill the remaining order.
	acc.SetCoins(coins("9000eur"))
	k.accountChanged(ctx, acc)
	orders = k.GetOrdersByOwner(acc.GetAddress())
	require.Equal(t, "6000", orders[0].SourceRemaining.String())
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

// Generate a random string to use as a client order id
func cid() string {
	return cmn.RandStr(10)
}
