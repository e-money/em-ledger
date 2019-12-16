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
	emauth "github.com/e-money/em-ledger/hooks/auth"
	"github.com/e-money/em-ledger/x/market/types"

	abci "github.com/tendermint/tendermint/abci/types"
	cmn "github.com/tendermint/tendermint/libs/common"
	dbm "github.com/tendermint/tm-db"
)

func TestBasicTrade(t *testing.T) {
	ctx, k, ak, _ := createTestComponents(t)

	acc1 := createAccount(ctx, ak, "acc1", "5000eur")
	acc2 := createAccount(ctx, ak, "acc2", "7400usd")

	order1 := order(acc1, "100eur", "120usd")
	res := k.NewOrderSingle(ctx, order1)
	require.True(t, res.IsOK())

	order2 := order(acc2, "60usd", "50eur")
	res = k.NewOrderSingle(ctx, order2)
	require.True(t, res.IsOK())

	bal1 := ak.GetAccount(ctx, acc1.GetAddress()).GetCoins()
	bal2 := ak.GetAccount(ctx, acc2.GetAddress()).GetCoins()
	require.Len(t, bal1, 2)
	require.Len(t, bal2, 2)

	require.Equal(t, "4950", bal1.AmountOf("eur").String())
	require.Equal(t, "60", bal1.AmountOf("usd").String())

	require.Equal(t, "50", bal2.AmountOf("eur").String())
	require.Equal(t, "7340", bal2.AmountOf("usd").String())

	require.Len(t, k.instruments, 1)

	i := k.instruments[0]
	remainingOrder := i.Orders.LeftKey().(*types.Order)
	require.Equal(t, int64(50), remainingOrder.SourceRemaining.Int64())
}

func TestMultipleOrders(t *testing.T) {
	ctx, k, ak, _ := createTestComponents(t)

	acc1 := createAccount(ctx, ak, "acc1", "10000eur")
	acc2 := createAccount(ctx, ak, "acc2", "7400usd")
	acc3 := createAccount(ctx, ak, "acc3", "2200chf")

	// Add two orders that draw on the same balance.
	res := k.NewOrderSingle(ctx, order(acc1, "10000eur", "11000usd"))
	require.True(t, res.IsOK())

	res = k.NewOrderSingle(ctx, order(acc1, "10000eur", "1400chf"))
	require.True(t, res.IsOK())

	require.Len(t, k.instruments, 2)

	res = k.NewOrderSingle(ctx, order(acc2, "7400usd", "5000eur"))
	require.True(t, res.IsOK(), res.Log)

	// Verify that acc1 still has two orders in the market with the same amount of remaining source tokens in each.
	orders := k.GetOrdersByOwner(acc1.GetAddress())
	require.Len(t, orders, 2)
	require.Equal(t, orders[0].SourceRemaining, orders[1].SourceRemaining)

	res = k.NewOrderSingle(ctx, order(acc3, "2200chf", "5000eur"))
	require.True(t, res.IsOK(), res.Log)

	// All acc1's EUR are sold by now. No orders should be on books
	orders = k.GetOrdersByOwner(acc1.GetAddress())
	require.Len(t, orders, 0)

	// Only a single instrument should remain chf -> eur
	require.Len(t, k.instruments, 1)
}

func TestCancelZeroRemainingOrders(t *testing.T) {
	ctx, k, ak, bk := createTestComponents(t)

	acc := createAccount(ctx, ak, "acc1", "10000eur")
	res := k.NewOrderSingle(ctx, order(acc, "10000eur", "11000usd"))
	require.True(t, res.IsOK())

	err := bk.SendCoins(ctx, acc.GetAddress(), sdk.AccAddress([]byte("void")), coins("10000eur"))
	require.NoError(t, err)

	orders := k.GetOrdersByOwner(acc.GetAddress())
	require.Len(t, orders, 0)
}

func TestInsufficientBalance1(t *testing.T) {
	ctx, k, ak, bk := createTestComponents(t)

	acc1 := createAccount(ctx, ak, "acc1", "500eur")
	acc2 := createAccount(ctx, ak, "acc2", "740usd")
	acc3 := createAccount(ctx, ak, "acc3", "")

	o := order(acc1, "300eur", "360usd")
	k.NewOrderSingle(ctx, o)

	// Modify account balance to be below order source
	bk.SendCoins(ctx, acc1.GetAddress(), acc3.GetAddress(), coins("250eur"))

	o = order(acc2, "360usd", "300eur")
	res := k.NewOrderSingle(ctx, o)
	require.True(t, res.IsOK())

	acc1 = ak.GetAccount(ctx, acc1.GetAddress())
	acc2 = ak.GetAccount(ctx, acc2.GetAddress())
	require.Equal(t, "300usd", acc1.GetCoins().String())
	require.Equal(t, "250eur,440usd", acc2.GetCoins().String())
}

func Test2(t *testing.T) {
	ctx, k, ak, _ := createTestComponents(t)

	acc1 := createAccount(ctx, ak, "acc1", "100eur")
	acc2 := createAccount(ctx, ak, "acc2", "121usd")

	o := order(acc1, "100eur", "120usd")
	res := k.NewOrderSingle(ctx, o)
	require.True(t, res.IsOK())

	o = order(acc2, "121usd", "100eur")
	res = k.NewOrderSingle(ctx, o)
	require.True(t, res.IsOK())

	require.Len(t, k.instruments, 1)

	remainingOrder := k.instruments[0].Orders.LeftKey().(*types.Order)
	require.Equal(t, int64(1), remainingOrder.SourceRemaining.Int64())
}

func Test3(t *testing.T) {
	ctx, k, ak, _ := createTestComponents(t)

	acc1 := createAccount(ctx, ak, "acc1", "100eur")
	acc2 := createAccount(ctx, ak, "acc2", "120usd")

	o := order(acc1, "100eur", "120usd")
	k.NewOrderSingle(ctx, o)

	for i := 0; i < 4; i++ {
		o = order(acc2, "30usd", "25eur")
		k.NewOrderSingle(ctx, o)
	}

	require.Len(t, k.instruments, 0)
	acc1 = ak.GetAccount(ctx, acc1.GetAddress())
	acc2 = ak.GetAccount(ctx, acc2.GetAddress())
	require.Equal(t, coins("120usd"), acc1.GetCoins())
	require.Equal(t, coins("100eur"), acc2.GetCoins())
}

func TestDeleteOrder(t *testing.T) {
	ctx, k, ak, _ := createTestComponents(t)
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
	ctx, k, ak, _ := createTestComponents(t)
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
	ctx, k, ak, _ := createTestComponents(t)
	acc := createAccount(ctx, ak, "acc1", "100eur")

	res := k.CancelOrder(ctx, acc.GetAddress(), "abcde")
	require.False(t, res.IsOK())
}

func TestCancelReplaceOrder(t *testing.T) {
	ctx, k, ak, _ := createTestComponents(t)
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
	ctx, k, ak, bk := createTestComponents(t)
	acc := createAccount(ctx, ak, "acc1", "15000eur")
	acc2 := createAccount(ctx, ak, "acc2", "11000chf,100000eur")

	order := types.NewOrder(coin("10000eur"), coin("1000usd"), acc.GetAddress(), cid())
	res := k.NewOrderSingle(ctx, order)
	require.True(t, res.IsOK())

	{
		// Partially fill the order above
		acc2 := createAccount(ctx, ak, "acc2", "900000usd")
		order2 := types.NewOrder(coin("400usd"), coin("4000eur"), acc2.GetAddress(), cid())
		res = k.NewOrderSingle(ctx, order2)
		require.True(t, res.IsOK())
	}

	err := bk.SendCoins(ctx, acc.GetAddress(), acc2.GetAddress(), coins("8000eur"))
	require.Nil(t, err)

	// Seller's account balance drops, remaining should be adjusted accordingly.
	orders := k.GetOrdersByOwner(acc.GetAddress())
	require.Len(t, orders, 1)
	require.Equal(t, coin("10000eur"), orders[0].Source)
	require.Equal(t, "3000", orders[0].SourceRemaining.String())
	require.Equal(t, "4000", orders[0].SourceFilled.String())

	// Seller's account balance is restored. Order should be adjusted, but take into consideration that the order has already been partially filled.
	err = bk.SendCoins(ctx, acc2.GetAddress(), acc.GetAddress(), coins("12000eur"))
	require.Nil(t, err)

	orders = k.GetOrdersByOwner(acc.GetAddress())
	require.Equal(t, "6000", orders[0].SourceRemaining.String())
	require.Equal(t, "4000", orders[0].SourceFilled.String())

	// Account balance dips below original sales amount, but can still fill the remaining order.
	err = bk.SendCoins(ctx, acc.GetAddress(), acc2.GetAddress(), coins("6000eur"))
	require.Nil(t, err)

	orders = k.GetOrdersByOwner(acc.GetAddress())
	require.Equal(t, "6000", orders[0].SourceRemaining.String())
}

func createTestComponents(t *testing.T) (sdk.Context, *Keeper, auth.AccountKeeper, bank.Keeper) {
	var (
		keyMarket  = sdk.NewKVStoreKey(types.ModuleName)
		authCapKey = sdk.NewKVStoreKey("authCapKey")
		keyParams  = sdk.NewKVStoreKey("params")
		tkeyParams = sdk.NewTransientStoreKey("transient_params")

		blacklistedAddrs = make(map[string]bool)
	)

	cdc := makeTestCodec()

	db := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(authCapKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyMarket, sdk.StoreTypeIAVL, db)

	err := ms.LoadLatestVersion()
	require.Nil(t, err)

	pk := params.NewKeeper(cdc, keyParams, tkeyParams, params.DefaultCodespace)

	ctx := sdk.NewContext(ms, abci.Header{ChainID: "test-chain"}, true, log.NewNopLogger())
	accountKeeper := auth.NewAccountKeeper(cdc, authCapKey, pk.Subspace(auth.DefaultParamspace), auth.ProtoBaseAccount)
	accountKeeperWrapped := emauth.Wrap(accountKeeper)

	bankKeeper := bank.NewBaseKeeper(accountKeeperWrapped, pk.Subspace(bank.DefaultParamspace), bank.DefaultCodespace, blacklistedAddrs)
	marketKeeper := NewKeeper(cdc, keyMarket, accountKeeperWrapped, bankKeeper)

	return ctx, marketKeeper, accountKeeper, bankKeeper
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

func order(account exported.Account, src, dst string) *types.Order {
	return types.NewOrder(coin(src), coin(dst), account.GetAddress(), cid())
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
