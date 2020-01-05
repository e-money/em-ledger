// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package keeper

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/store"
	"github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/supply"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/log"
	"math"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	emauth "github.com/e-money/em-ledger/hooks/auth"
	"github.com/e-money/em-ledger/x/market/types"

	abci "github.com/tendermint/tendermint/abci/types"
	cmn "github.com/tendermint/tendermint/libs/common"
	dbm "github.com/tendermint/tm-db"
)

func TestBasicTrade(t *testing.T) {
	ctx, k, ak, _, _ := createTestComponents(t)

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

func TestBasicTrade2(t *testing.T) {
	ctx, k, ak, _, _ := createTestComponents(t)

	acc1 := createAccount(ctx, ak, "acc1", "888eur")
	acc2 := createAccount(ctx, ak, "acc2", "1120usd")

	order1 := order(acc1, "888eur", "1121usd")
	res := k.NewOrderSingle(ctx, order1)
	require.True(t, res.IsOK())

	order2 := order(acc2, "1120usd", "890eur")
	res = k.NewOrderSingle(ctx, order2)
	require.True(t, res.IsOK(), res.Log)

	bal1 := ak.GetAccount(ctx, acc1.GetAddress()).GetCoins()
	bal2 := ak.GetAccount(ctx, acc2.GetAddress()).GetCoins()
	fmt.Println("acc1", bal1)
	fmt.Println("acc2", bal2)
}

func TestMultipleOrders(t *testing.T) {
	ctx, k, ak, _, _ := createTestComponents(t)

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

	res = k.NewOrderSingle(ctx, order(acc3, "2200chf", "5000eur"))
	require.True(t, res.IsOK(), res.Log)

	// All acc1's EUR are sold by now. No orders should be on books
	orders := k.GetOrdersByOwner(acc1.GetAddress())
	require.Len(t, orders, 0)

	// Only a single instrument should remain chf -> eur
	require.Len(t, k.instruments, 1)
}

func TestCancelZeroRemainingOrders(t *testing.T) {
	ctx, k, ak, bk, _ := createTestComponents(t)

	acc := createAccount(ctx, ak, "acc1", "10000eur")
	res := k.NewOrderSingle(ctx, order(acc, "10000eur", "11000usd"))
	require.True(t, res.IsOK())

	err := bk.SendCoins(ctx, acc.GetAddress(), sdk.AccAddress([]byte("void")), coins("10000eur"))
	require.NoError(t, err)

	orders := k.GetOrdersByOwner(acc.GetAddress())
	require.Len(t, orders, 0)
}

func TestInsufficientBalance1(t *testing.T) {
	ctx, k, ak, bk, _ := createTestComponents(t)

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
	ctx, k, ak, _, _ := createTestComponents(t)

	acc1 := createAccount(ctx, ak, "acc1", "100eur")
	acc2 := createAccount(ctx, ak, "acc2", "121usd")

	o := order(acc1, "100eur", "120usd")
	res := k.NewOrderSingle(ctx, o)
	require.True(t, res.IsOK())

	o = order(acc2, "121usd", "100eur")
	res = k.NewOrderSingle(ctx, o)
	require.True(t, res.IsOK())

	require.Empty(t, k.instruments)
	require.Equal(t, coins("120usd"), ak.GetAccount(ctx, acc1.GetAddress()).GetCoins())
	require.Equal(t, coins("100eur,1usd"), ak.GetAccount(ctx, acc2.GetAddress()).GetCoins())
}

func Test3(t *testing.T) {
	ctx, k, ak, _, _ := createTestComponents(t)

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
	ctx, k, ak, _, _ := createTestComponents(t)

	acc1 := createAccount(ctx, ak, "acc1", "100eur")

	cid := cid()

	order1, _ := types.NewOrder(coin("100eur"), coin("120usd"), acc1.GetAddress(), time.Now(), cid)
	res := k.NewOrderSingle(ctx, order1)
	require.True(t, res.IsOK())

	order2, _ := types.NewOrder(coin("100eur"), coin("77chf"), acc1.GetAddress(), time.Now(), cid)
	res = k.NewOrderSingle(ctx, order2)
	require.False(t, res.IsOK()) // Verify that client order ids cannot be duplicated.

	require.Len(t, k.instruments, 1) // Ensure that the eur->chf pair was not added.

	k.deleteOrder(ctx, &order1)
	require.Len(t, k.instruments, 0) // Removing the only eur->usd order should have removed instrument
}

func TestGetOrdersByOwnerAndCancel(t *testing.T) {
	ctx, k, ak, _, _ := createTestComponents(t)
	acc1 := createAccount(ctx, ak, "acc1", "100eur")
	acc2 := createAccount(ctx, ak, "acc2", "120usd")

	for i := 0; i < 5; i++ {
		order, _ := types.NewOrder(coin("5eur"), coin("12usd"), acc1.GetAddress(), time.Now(), cid())
		res := k.NewOrderSingle(ctx, order)
		require.True(t, res.IsOK())
	}

	for i := 0; i < 5; i++ {
		order, _ := types.NewOrder(coin("7usd"), coin("3chf"), acc2.GetAddress(), time.Now(), cid())
		res := k.NewOrderSingle(ctx, order)
		require.True(t, res.IsOK(), res.Log)
	}

	allOrders1 := k.GetOrdersByOwner(acc1.GetAddress())
	require.Len(t, allOrders1, 5)

	{
		order, _ := types.NewOrder(coin("12usd"), coin("5eur"), acc2.GetAddress(), time.Now(), cid())
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
	ctx, k, ak, _, _ := createTestComponents(t)
	acc := createAccount(ctx, ak, "acc1", "100eur")

	res := k.CancelOrder(ctx, acc.GetAddress(), "abcde")
	require.False(t, res.IsOK())
}

func TestCancelReplaceOrder(t *testing.T) {
	ctx, k, ak, _, _ := createTestComponents(t)
	acc1 := createAccount(ctx, ak, "acc1", "20000eur")
	acc2 := createAccount(ctx, ak, "acc2", "45000usd")

	order1cid := cid()
	order1, _ := types.NewOrder(coin("500eur"), coin("1200usd"), acc1.GetAddress(), time.Now(), order1cid)
	res := k.NewOrderSingle(ctx, order1)
	require.True(t, res.IsOK())

	order2cid := cid()
	order2, _ := types.NewOrder(coin("5000eur"), coin("17000usd"), acc1.GetAddress(), time.Now(), order2cid)
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

	order3, _ := types.NewOrder(coin("500chf"), coin("1700usd"), acc1.GetAddress(), time.Now(), cid())
	// Wrong client order id for previous order submitted.
	res = k.CancelReplaceOrder(ctx, order3, order1cid)
	require.Equal(t, types.CodeClientOrderIdNotFound, res.Code)

	// Changing instrument of order
	res = k.CancelReplaceOrder(ctx, order3, order2cid)
	require.Equal(t, types.CodeOrderInstrumentChanged, res.Code)

	o, _ := types.NewOrder(coin("2600usd"), coin("300eur"), acc2.GetAddress(), time.Now(), cid())
	res = k.NewOrderSingle(ctx, o)
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
	order4, _ := types.NewOrder(coin("10000eur"), coin("35050usd"), acc1.GetAddress(), time.Now(), order4cid)
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
	ctx, k, ak, bk, _ := createTestComponents(t)
	acc := createAccount(ctx, ak, "acc1", "15000eur")
	acc2 := createAccount(ctx, ak, "acc2", "11000chf,100000eur")

	order, _ := types.NewOrder(coin("10000eur"), coin("1000usd"), acc.GetAddress(), time.Now(), cid())
	res := k.NewOrderSingle(ctx, order)
	require.True(t, res.IsOK())

	{
		// Partially fill the order above
		acc2 := createAccount(ctx, ak, "acc2", "900000usd")
		order2, _ := types.NewOrder(coin("400usd"), coin("4000eur"), acc2.GetAddress(), time.Now(), cid())
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
	require.Equal(t, "400", orders[0].DestinationFilled.String())

	// Seller's account balance is restored. Order should be adjusted, but take into consideration that the order has already been partially filled.
	err = bk.SendCoins(ctx, acc2.GetAddress(), acc.GetAddress(), coins("12000eur"))
	require.Nil(t, err)

	orders = k.GetOrdersByOwner(acc.GetAddress())
	require.Equal(t, "6000", orders[0].SourceRemaining.String())
	require.Equal(t, "400", orders[0].DestinationFilled.String())

	// Account balance dips below original sales amount, but can still fill the remaining order.
	err = bk.SendCoins(ctx, acc.GetAddress(), acc2.GetAddress(), coins("6000eur"))
	require.Nil(t, err)

	orders = k.GetOrdersByOwner(acc.GetAddress())
	require.Equal(t, "6000", orders[0].SourceRemaining.String())
}

func TestUnknownAsset(t *testing.T) {
	ctx, k1, ak, _, _ := createTestComponents(t)

	acc1 := createAccount(ctx, ak, "acc1", "5000eur")

	// Make an order with a destination that is not known by the supply module
	o := order(acc1, "1000eur", "1200nok")
	res := k1.NewOrderSingle(ctx, o)
	require.False(t, res.IsOK())
	require.Equal(t, types.Codespace, res.Codespace)
	require.Equal(t, types.CodeUnknownAsset, res.Code)
}

func TestLoadFromStore(t *testing.T) {
	// Create order book with a number of passive orders.
	ctx, k1, ak, _, _ := createTestComponents(t)

	acc1 := createAccount(ctx, ak, "acc1", "5000eur")
	acc2 := createAccount(ctx, ak, "acc2", "7400usd")

	o := order(acc1, "1000eur", "1200usd")
	require.True(t, k1.NewOrderSingle(ctx, o).IsOK())

	o = order(acc2, "5000usd", "3500chf")
	require.True(t, k1.NewOrderSingle(ctx, o).IsOK())

	_, k2, _, _, _ := createTestComponents(t)

	k2.key = k1.key
	// Create new keeper and let it inherit the store of the previous keeper
	k2.initializeFromStore(ctx)

	// Verify that all orders are loaded correctly into the book
	require.Len(t, k2.instruments, len(k1.instruments))

	require.Equal(t, 1, k2.accountOrders.GetAllOrders(acc1.GetAddress()).Size())
	require.Equal(t, 1, k2.accountOrders.GetAllOrders(acc2.GetAddress()).Size())
}

func TestVestingAccount(t *testing.T) {
	ctx, keeper, ak, _, _ := createTestComponents(t)
	account := createAccount(ctx, ak, "acc1", "110000eur")

	vestingAcc := auth.NewDelayedVestingAccount(account.(*auth.BaseAccount), math.MaxInt64)
	ak.SetAccount(ctx, vestingAcc)

	res := keeper.NewOrderSingle(ctx, order(vestingAcc, "5000eur", "4700chf"))
	require.False(t, res.IsOK())
}

func TestInvalidInstrument(t *testing.T) {
	ctx, k, ak, _, _ := createTestComponents(t)

	acc1 := createAccount(ctx, ak, "acc1", "5000eur")

	// Ensure that an order cannot contain the same denomination in source and destination
	o := types.Order{
		ID:                124,
		Source:            coin("125eur"),
		Destination:       coin("250eur"),
		DestinationFilled: sdk.ZeroInt(),
		Owner:             acc1.GetAddress(),
		ClientOrderID:     "abcddeg",
	}

	res := k.NewOrderSingle(ctx, o)
	require.False(t, res.IsOK())
}

func TestSyntheticPrice1(t *testing.T) {
	ctx, k, ak, _, _ := createTestComponents(t)
	acc1 := createAccount(ctx, ak, "acc1", "5000eur")
	acc2 := createAccount(ctx, ak, "acc2", "6500usd")
	acc3 := createAccount(ctx, ak, "acc3", "4500chf")

	printTotalBalance(acc1, acc2, acc3)

	fmt.Println("acc1:", ak.GetAccount(ctx, acc1.GetAddress()).GetCoins(), acc1.GetAddress().String())
	fmt.Println("acc2:", ak.GetAccount(ctx, acc2.GetAddress()).GetCoins(), acc2.GetAddress().String())
	fmt.Println("acc3:", ak.GetAccount(ctx, acc3.GetAddress()).GetCoins(), acc3.GetAddress().String())

	o := order(acc1, "1000eur", "1114usd")
	require.True(t, k.NewOrderSingle(ctx, o).IsOK())

	//o = order(acc1, "1000eur", "1084chf")
	o = order(acc1, "500eur", "542chf")
	require.True(t, k.NewOrderSingle(ctx, o).IsOK())

	o = order(acc3, "1000chf", "1028usd")
	require.True(t, k.NewOrderSingle(ctx, o).IsOK())

	o = order(acc2, "5000usd", "4490eur")
	require.True(t, k.NewOrderSingle(ctx, o).IsOK())

	fmt.Println("acc1:", ak.GetAccount(ctx, acc1.GetAddress()).GetCoins())
	fmt.Println("acc2:", ak.GetAccount(ctx, acc2.GetAddress()).GetCoins())
	fmt.Println("acc3:", ak.GetAccount(ctx, acc3.GetAddress()).GetCoins())

	printTotalBalance(ak.GetAccount(ctx, acc1.GetAddress()), ak.GetAccount(ctx, acc2.GetAddress()), ak.GetAccount(ctx, acc3.GetAddress()))

}

func printTotalBalance(accs ...exported.Account) {
	sum := sdk.NewCoins()

	for _, acc := range accs {
		sum = sum.Add(acc.GetCoins())
	}

	fmt.Println(sum)
}

func createTestComponents(t *testing.T) (sdk.Context, *Keeper, auth.AccountKeeper, bank.Keeper, supply.Keeper) {
	var (
		keyMarket  = sdk.NewKVStoreKey(types.ModuleName)
		authCapKey = sdk.NewKVStoreKey("authCapKey")
		keyParams  = sdk.NewKVStoreKey("params")
		supplyKey  = sdk.NewKVStoreKey("supply")
		tkeyParams = sdk.NewTransientStoreKey("transient_params")

		blacklistedAddrs = make(map[string]bool)
	)

	db := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(authCapKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyMarket, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(supplyKey, sdk.StoreTypeIAVL, db)

	err := ms.LoadLatestVersion()
	require.Nil(t, err)

	pk := params.NewKeeper(types.ModuleCdc, keyParams, tkeyParams, params.DefaultCodespace)

	ctx := sdk.NewContext(ms, abci.Header{ChainID: "test-chain"}, true, log.NewNopLogger())
	accountKeeper := auth.NewAccountKeeper(types.ModuleCdc, authCapKey, pk.Subspace(auth.DefaultParamspace), auth.ProtoBaseAccount)
	accountKeeperWrapped := emauth.Wrap(accountKeeper)

	bankKeeper := bank.NewBaseKeeper(accountKeeperWrapped, pk.Subspace(bank.DefaultParamspace), bank.DefaultCodespace, blacklistedAddrs)

	maccPerms := map[string][]string{}

	supplyKeeper := supply.NewKeeper(types.ModuleCdc, supplyKey, accountKeeper, bankKeeper, maccPerms)
	supplyKeeper.SetSupply(ctx, supply.NewSupply(coins("1eur,1usd,1chf,1jpy")))

	marketKeeper := NewKeeper(types.ModuleCdc, keyMarket, accountKeeperWrapped, bankKeeper, supplyKeeper)

	return ctx, marketKeeper, accountKeeper, bankKeeper, supplyKeeper
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

func order(account exported.Account, src, dst string) types.Order {
	o, err := types.NewOrder(coin(src), coin(dst), account.GetAddress(), time.Now(), cid())
	if err != nil {
		panic(err)
	}

	return o
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
