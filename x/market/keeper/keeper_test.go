// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package keeper

import (
	"fmt"
	authexported "github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
	"math"
	"testing"
	"time"

	emauth "github.com/e-money/em-ledger/hooks/auth"
	emtypes "github.com/e-money/em-ledger/types"
	"github.com/e-money/em-ledger/x/market/types"

	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/supply"
	"github.com/stretchr/testify/require"
	tmrand "github.com/tendermint/tendermint/libs/rand"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"
)

func init() {
	emtypes.ConfigureSDK()
}

func TestBasicTrade(t *testing.T) {
	ctx, k, ak, _, _ := createTestComponents(t)

	acc1 := createAccount(ctx, ak, "acc1", "5000eur")
	acc2 := createAccount(ctx, ak, "acc2", "7400usd")

	totalSupply := snapshotAccounts(ctx, ak)

	gasmeter := sdk.NewGasMeter(math.MaxUint64)
	order1 := order(acc1, "100eur", "120usd")
	_, err := k.NewOrderSingle(ctx.WithGasMeter(gasmeter), order1)
	require.NoError(t, err)
	require.Equal(t, gasPriceNewOrder, gasmeter.GasConsumed())

	gasmeter = sdk.NewGasMeter(math.MaxUint64)
	order2 := order(acc2, "60usd", "50eur")
	_, err = k.NewOrderSingle(ctx.WithGasMeter(gasmeter), order2)
	require.NoError(t, err)

	// Ensure that gas usage is not higher due to the order being matched.
	require.Equal(t, gasPriceNewOrder, gasmeter.GasConsumed())

	bal1 := ak.GetAccount(ctx, acc1.GetAddress()).GetCoins()
	bal2 := ak.GetAccount(ctx, acc2.GetAddress()).GetCoins()
	require.Len(t, bal1, 2)
	require.Len(t, bal2, 2)

	require.Equal(t, "4950", bal1.AmountOf("eur").String())
	require.Equal(t, "60", bal1.AmountOf("usd").String())

	require.Equal(t, "50", bal2.AmountOf("eur").String())
	require.Equal(t, "7340", bal2.AmountOf("usd").String())

	//require.Len(t, k.instruments, 1)

	//i := k.instruments[0]
	//remainingOrder := i.Orders.LeftKey().(*types.Order)
	//require.Equal(t, int64(50), remainingOrder.SourceRemaining.Int64())

	require.True(t, totalSupply.Sub(snapshotAccounts(ctx, ak)).IsZero())
}

func TestBasicTrade2(t *testing.T) {
	ctx, k, ak, _, _ := createTestComponents(t)

	acc1 := createAccount(ctx, ak, "acc1", "888eur")
	acc2 := createAccount(ctx, ak, "acc2", "1120usd")

	totalSupply := snapshotAccounts(ctx, ak)

	order1 := order(acc1, "888eur", "1121usd")
	_, err := k.NewOrderSingle(ctx, order1)
	require.NoError(t, err)

	order2 := order(acc2, "1120usd", "890eur")
	res, err := k.NewOrderSingle(ctx, order2)
	require.True(t, err == nil, res.Log)

	require.True(t, totalSupply.Sub(snapshotAccounts(ctx, ak)).IsZero())
}

func TestInsufficientGas(t *testing.T) {
	ctx, k, ak, _, _ := createTestComponents(t)

	acc1 := createAccount(ctx, ak, "acc1", "888eur")
	order1 := order(acc1, "888eur", "1121usd")

	gasMeter := sdk.NewGasMeter(gasPriceNewOrder - 5000)

	require.Panics(t, func() {
		k.NewOrderSingle(ctx.WithGasMeter(gasMeter), order1)
	})
}

func TestMultipleOrders(t *testing.T) {
	ctx, k, ak, _, _ := createTestComponents(t)

	acc1 := createAccount(ctx, ak, "acc1", "10000eur")
	acc2 := createAccount(ctx, ak, "acc2", "7400usd")
	acc3 := createAccount(ctx, ak, "acc3", "2200chf")

	totalSupply := snapshotAccounts(ctx, ak)

	// Add two orders that draw on the same balance.
	_, err := k.NewOrderSingle(ctx, order(acc1, "10000eur", "11000usd"))
	require.NoError(t, err)

	_, err = k.NewOrderSingle(ctx, order(acc1, "10000eur", "1400chf"))
	require.NoError(t, err)

	//require.Len(t, k.instruments, 2)

	res, err := k.NewOrderSingle(ctx, order(acc2, "7400usd", "5000eur"))
	require.True(t, err == nil, res.Log)

	res, err = k.NewOrderSingle(ctx, order(acc3, "2200chf", "5000eur"))
	require.True(t, err == nil, res.Log)

	// All acc1's EUR are sold by now. No orders should be on books
	orders := k.GetOrdersByOwner(ctx, acc1.GetAddress())
	require.Len(t, orders, 0)

	// All orders should be filled
	//require.Empty(t, k.instruments)

	require.True(t, totalSupply.Sub(snapshotAccounts(ctx, ak)).IsZero())
}

func TestCancelZeroRemainingOrders(t *testing.T) {
	ctx, k, ak, bk, _ := createTestComponents(t)

	acc := createAccount(ctx, ak, "acc1", "10000eur")
	_, err := k.NewOrderSingle(ctx, order(acc, "10000eur", "11000usd"))
	require.NoError(t, err)

	err = bk.SendCoins(ctx, acc.GetAddress(), sdk.AccAddress([]byte("void")), coins("10000eur"))
	require.NoError(t, err)

	orders := k.GetOrdersByOwner(ctx, acc.GetAddress())
	require.Len(t, orders, 0)
}

func TestInsufficientBalance1(t *testing.T) {
	ctx, k, ak, bk, _ := createTestComponents(t)

	acc1 := createAccount(ctx, ak, "acc1", "500eur")
	acc2 := createAccount(ctx, ak, "acc2", "740usd")
	acc3 := createAccount(ctx, ak, "acc3", "")

	totalSupply := snapshotAccounts(ctx, ak)

	o := order(acc1, "300eur", "360usd")
	k.NewOrderSingle(ctx, o)

	// Modify account balance to be below order source
	bk.SendCoins(ctx, acc1.GetAddress(), acc3.GetAddress(), coins("250eur"))

	o = order(acc2, "360usd", "300eur")
	_, err := k.NewOrderSingle(ctx, o)
	require.NoError(t, err)

	acc1 = ak.GetAccount(ctx, acc1.GetAddress())
	acc2 = ak.GetAccount(ctx, acc2.GetAddress())
	require.Equal(t, "300usd", acc1.GetCoins().String())
	require.Equal(t, "250eur,440usd", acc2.GetCoins().String())

	require.True(t, totalSupply.Sub(snapshotAccounts(ctx, ak)).IsZero())
}

func Test2(t *testing.T) {
	ctx, k, ak, _, _ := createTestComponents(t)

	acc1 := createAccount(ctx, ak, "acc1", "100eur")
	acc2 := createAccount(ctx, ak, "acc2", "121usd")

	totalSupply := snapshotAccounts(ctx, ak)

	o := order(acc1, "100eur", "120usd")
	_, err := k.NewOrderSingle(ctx, o)
	require.NoError(t, err)

	o = order(acc2, "121usd", "100eur")
	_, err = k.NewOrderSingle(ctx, o)
	require.NoError(t, err)

	//require.Empty(t, k.instruments)
	require.Equal(t, coins("120usd"), ak.GetAccount(ctx, acc1.GetAddress()).GetCoins())
	require.Equal(t, coins("100eur,1usd"), ak.GetAccount(ctx, acc2.GetAddress()).GetCoins())

	require.True(t, totalSupply.Sub(snapshotAccounts(ctx, ak)).IsZero())
}

func Test3(t *testing.T) {
	ctx, k, ak, _, _ := createTestComponents(t)

	acc1 := createAccount(ctx, ak, "acc1", "100eur")
	acc2 := createAccount(ctx, ak, "acc2", "120usd")

	totalSupply := snapshotAccounts(ctx, ak)

	o := order(acc1, "100eur", "120usd")
	k.NewOrderSingle(ctx, o)

	gasMeter := sdk.NewGasMeter(math.MaxUint64)
	for i := 0; i < 4; i++ {
		o = order(acc2, "30usd", "25eur")
		k.NewOrderSingle(ctx.WithGasMeter(gasMeter), o)
	}
	require.Equal(t, 4*gasPriceNewOrder, gasMeter.GasConsumed())

	//require.Len(t, k.instruments, 0)
	acc1 = ak.GetAccount(ctx, acc1.GetAddress())
	acc2 = ak.GetAccount(ctx, acc2.GetAddress())
	require.Equal(t, coins("120usd"), acc1.GetCoins())
	require.Equal(t, coins("100eur"), acc2.GetCoins())

	require.True(t, totalSupply.Sub(snapshotAccounts(ctx, ak)).IsZero())
}

func TestDeleteOrder(t *testing.T) {
	ctx, k, ak, _, _ := createTestComponents(t)

	acc1 := createAccount(ctx, ak, "acc1", "100eur")
	totalSupply := snapshotAccounts(ctx, ak)

	cid := cid()

	order1, _ := types.NewOrder(coin("100eur"), coin("120usd"), acc1.GetAddress(), time.Now(), cid)
	_, err := k.NewOrderSingle(ctx, order1)
	require.NoError(t, err)

	order2, _ := types.NewOrder(coin("100eur"), coin("77chf"), acc1.GetAddress(), time.Now(), cid)
	_, err = k.NewOrderSingle(ctx, order2)
	require.Error(t, err) // Verify that client order ids cannot be duplicated.

	//require.Len(t, k.instruments, 1) // Ensure that the eur->chf pair was not added.

	//k.deleteOrder(ctx, &order1)
	//require.Len(t, k.instruments, 0) // Removing the only eur->usd order should have removed instrument

	require.True(t, totalSupply.Sub(snapshotAccounts(ctx, ak)).IsZero())
}

func TestGetOrdersByOwnerAndCancel(t *testing.T) {
	ctx, k, ak, _, _ := createTestComponents(t)
	acc1 := createAccount(ctx, ak, "acc1", "100eur")
	acc2 := createAccount(ctx, ak, "acc2", "120usd")

	for i := 0; i < 5; i++ {
		order, _ := types.NewOrder(coin("5eur"), coin("12usd"), acc1.GetAddress(), time.Now(), cid())
		_, err := k.NewOrderSingle(ctx, order)
		require.NoError(t, err)
	}

	for i := 0; i < 5; i++ {
		order, _ := types.NewOrder(coin("7usd"), coin("3chf"), acc2.GetAddress(), time.Now(), cid())
		res, err := k.NewOrderSingle(ctx, order)
		require.True(t, err == nil, res.Log)
	}

	allOrders1 := k.GetOrdersByOwner(ctx, acc1.GetAddress())
	require.Len(t, allOrders1, 5)

	{
		order, _ := types.NewOrder(coin("12usd"), coin("5eur"), acc2.GetAddress(), time.Now(), cid())
		res, err := k.NewOrderSingle(ctx, order)
		require.True(t, err == nil, res.Log)
	}

	allOrders2 := k.GetOrdersByOwner(ctx, acc1.GetAddress())
	require.Len(t, allOrders2, 4)

	cid := allOrders2[2].ClientOrderID
	gasMeter := sdk.NewGasMeter(math.MaxUint64)
	_, err := k.CancelOrder(ctx.WithGasMeter(gasMeter), acc1.GetAddress(), cid)
	require.NoError(t, err)

	_, err = k.CancelOrder(ctx.WithGasMeter(gasMeter), acc1.GetAddress(), cid)
	require.Error(t, err)

	require.Equal(t, 2*gasPriceCancelOrder, gasMeter.GasConsumed())

	allOrders3 := k.GetOrdersByOwner(ctx, acc1.GetAddress())
	require.Len(t, allOrders3, 3)

	found := false
	for _, e := range ctx.EventManager().Events() {
		found = found || (e.Type == types.EventTypeCancel)
	}

	require.True(t, found)
}

func TestCancelOrders1(t *testing.T) {
	// Cancel a non-existing order by an account with no orders in the system.
	ctx, k, ak, _, _ := createTestComponents(t)
	acc := createAccount(ctx, ak, "acc1", "100eur")

	_, err := k.CancelOrder(ctx, acc.GetAddress(), "abcde")
	require.Error(t, err)
}

func TestCancelReplaceOrder(t *testing.T) {
	ctx, k, ak, _, _ := createTestComponents(t)
	acc1 := createAccount(ctx, ak, "acc1", "20000eur")
	acc2 := createAccount(ctx, ak, "acc2", "45000usd")

	totalSupply := snapshotAccounts(ctx, ak)

	order1cid := cid()
	order1, _ := types.NewOrder(coin("500eur"), coin("1200usd"), acc1.GetAddress(), time.Now(), order1cid)
	_, err := k.NewOrderSingle(ctx, order1)
	require.NoError(t, err)

	gasMeter := sdk.NewGasMeter(math.MaxUint64)
	order2cid := cid()
	order2, _ := types.NewOrder(coin("5000eur"), coin("17000usd"), acc1.GetAddress(), time.Now(), order2cid)
	res, err := k.CancelReplaceOrder(ctx.WithGasMeter(gasMeter), order2, order1cid)
	require.True(t, err == nil, res.Log)
	require.Equal(t, gasPriceCancelReplaceOrder, gasMeter.GasConsumed())

	{
		orders := k.GetOrdersByOwner(ctx, acc1.GetAddress())
		require.Len(t, orders, 1)
		require.Equal(t, order2cid, orders[0].ClientOrderID)
		require.Equal(t, coin("5000eur"), orders[0].Source)
		require.Equal(t, coin("17000usd"), orders[0].Destination)
		require.Equal(t, sdk.NewInt(5000), orders[0].SourceRemaining)
	}

	order3, _ := types.NewOrder(coin("500chf"), coin("1700usd"), acc1.GetAddress(), time.Now(), cid())
	// Wrong client order id for previous order submitted.
	_, err = k.CancelReplaceOrder(ctx, order3, order1cid)
	require.True(t, types.ErrClientOrderIdNotFound.Is(err))
	//require.Equal(t, types.CodeClientOrderIdNotFound, res.Code)

	// Changing instrument of order
	gasMeter = sdk.NewGasMeter(math.MaxUint64)
	_, err = k.CancelReplaceOrder(ctx.WithGasMeter(gasMeter), order3, order2cid)
	require.True(t, types.ErrOrderInstrumentChanged.Is(err))
	//require.Equal(t, types.CodeOrderInstrumentChanged, res.Code)
	require.Equal(t, gasPriceCancelReplaceOrder, gasMeter.GasConsumed())

	o := order(acc2, "2600usd", "300eur")
	_, err = k.NewOrderSingle(ctx, o)
	require.NoError(t, err)

	acc1 = ak.GetAccount(ctx, acc1.GetAddress())
	acc2 = ak.GetAccount(ctx, acc2.GetAddress())

	require.Equal(t, int64(300), acc2.GetCoins().AmountOf("eur").Int64())
	require.Equal(t, int64(1020), acc1.GetCoins().AmountOf("usd").Int64())

	filled := sdk.ZeroInt()
	{
		orders := k.GetOrdersByOwner(ctx, acc1.GetAddress())
		require.Len(t, orders, 1)
		filled = orders[0].Source.Amount.Sub(orders[0].SourceRemaining)
	}

	// CancelReplace and verify that previously filled amount is subtracted from the resulting order
	order4cid := cid()
	order4, _ := types.NewOrder(coin("10000eur"), coin("35050usd"), acc1.GetAddress(), time.Now(), order4cid)
	res, err = k.CancelReplaceOrder(ctx, order4, order2cid)
	require.True(t, err == nil, res.Log)

	{
		orders := k.GetOrdersByOwner(ctx, acc1.GetAddress())
		require.Len(t, orders, 1)
		require.Equal(t, order4cid, orders[0].ClientOrderID)
		require.Equal(t, coin("10000eur"), orders[0].Source)
		require.Equal(t, coin("35050usd"), orders[0].Destination)
		require.Equal(t, sdk.NewInt(10000).Sub(filled), orders[0].SourceRemaining)
	}

	// CancelReplace with an order that asks for a larger source than the replaced order has remaining
	order5 := order(acc2, "42000usd", "8000eur")
	k.NewOrderSingle(ctx, order5)
	require.True(t, err == nil, res.Log)

	order6 := order(acc1, "8000eur", "30000usd")
	_, err = k.CancelReplaceOrder(ctx, order6, order4cid)
	require.True(t, types.ErrNoSourceRemaining.Is(err))
	//require.Equal(t, types.CodeNoSourceRemaining, res.Code)

	require.True(t, totalSupply.Sub(snapshotAccounts(ctx, ak)).IsZero())
}

func TestOrdersChangeWithAccountBalance(t *testing.T) {
	ctx, k, ak, bk, _ := createTestComponents(t)
	acc := createAccount(ctx, ak, "acc1", "15000eur")
	acc2 := createAccount(ctx, ak, "acc2", "11000chf,100000eur")

	order, _ := types.NewOrder(coin("10000eur"), coin("1000usd"), acc.GetAddress(), time.Now(), cid())
	_, err := k.NewOrderSingle(ctx, order)
	require.NoError(t, err)

	{
		// Partially fill the order above
		acc2 := createAccount(ctx, ak, "acc2", "900000usd")
		order2, _ := types.NewOrder(coin("400usd"), coin("4000eur"), acc2.GetAddress(), time.Now(), cid())
		_, err = k.NewOrderSingle(ctx, order2)
		require.NoError(t, err)
	}

	totalSupply := snapshotAccounts(ctx, ak)

	err = bk.SendCoins(ctx, acc.GetAddress(), acc2.GetAddress(), coins("8000eur"))
	require.Nil(t, err)

	// Seller's account balance drops, remaining should be adjusted accordingly.
	orders := k.GetOrdersByOwner(ctx, acc.GetAddress())
	require.Len(t, orders, 1)
	require.Equal(t, coin("10000eur"), orders[0].Source)
	require.Equal(t, "3000", orders[0].SourceRemaining.String())
	require.Equal(t, "400", orders[0].DestinationFilled.String())

	// Seller's account balance is restored. Order should be adjusted, but take into consideration that the order has already been partially filled.
	err = bk.SendCoins(ctx, acc2.GetAddress(), acc.GetAddress(), coins("12000eur"))
	require.Nil(t, err)

	orders = k.GetOrdersByOwner(ctx, acc.GetAddress())
	require.Equal(t, "6000", orders[0].SourceRemaining.String())
	require.Equal(t, "400", orders[0].DestinationFilled.String())

	// Account balance dips below original sales amount, but can still fill the remaining order.
	err = bk.SendCoins(ctx, acc.GetAddress(), acc2.GetAddress(), coins("6000eur"))
	require.Nil(t, err)

	orders = k.GetOrdersByOwner(ctx, acc.GetAddress())
	require.Equal(t, "6000", orders[0].SourceRemaining.String())

	require.True(t, totalSupply.Sub(snapshotAccounts(ctx, ak)).IsZero())
}

func TestUnknownAsset(t *testing.T) {
	ctx, k1, ak, _, _ := createTestComponents(t)

	acc1 := createAccount(ctx, ak, "acc1", "5000eur")

	gasMeter := sdk.NewGasMeter(math.MaxUint64)

	// Make an order with a destination that is not known by the supply module
	o := order(acc1, "1000eur", "1200nok")
	_, err := k1.NewOrderSingle(ctx.WithGasMeter(gasMeter), o)
	require.True(t, types.ErrUnknownAsset.Is(err))
	//require.Equal(t, types.Codespace, res.Codespace)
	//require.Equal(t, types.CodeUnknownAsset, res.Code)
	require.Equal(t, gasPriceNewOrder, gasMeter.GasConsumed())
}

func TestLoadFromStore(t *testing.T) {
	// Create order book with a number of passive orders.
	ctx, k1, ak, _, _ := createTestComponents(t)

	acc1 := createAccount(ctx, ak, "acc1", "5000eur")
	acc2 := createAccount(ctx, ak, "acc2", "7400usd")

	o := order(acc1, "1000eur", "1200usd")
	_, err := k1.NewOrderSingle(ctx, o)
	require.NoError(t, err)

	o = order(acc2, "5000usd", "3500chf")
	_, err = k1.NewOrderSingle(ctx, o)
	require.NoError(t, err)

	_, k2, _, _, _ := createTestComponents(t)

	k2.key = k1.key
	// Create new keeper and let it inherit the store of the previous keeper
	k2.initializeFromStore(ctx)

	// Verify that all orders are loaded correctly into the book
	//require.Len(t, k2.instruments, len(k1.instruments))

	//require.Equal(t, 1, k2.accountOrders.GetAllOrders(acc1.GetAddress()).Size())
	//require.Equal(t, 1, k2.accountOrders.GetAllOrders(acc2.GetAddress()).Size())
}

func TestVestingAccount(t *testing.T) {
	ctx, keeper, ak, _, _ := createTestComponents(t)
	account := createAccount(ctx, ak, "acc1", "110000eur")

	vestingAcc := vesting.NewDelayedVestingAccount(account.(*auth.BaseAccount), math.MaxInt64)
	ak.SetAccount(ctx, vestingAcc)

	_, err := keeper.NewOrderSingle(ctx, order(vestingAcc, "5000eur", "4700chf"))
	require.True(t, types.ErrAccountBalanceInsufficient.Is(err))
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

	_, err := k.NewOrderSingle(ctx, o)
	require.True(t, types.ErrInvalidInstrument.Is(err))
}

func TestRestrictedDenominations1(t *testing.T) {
	ctx, k, ak, _, _ := createTestComponents(t)

	acc1 := createAccount(ctx, ak, "acc1", "5000gbp, 10000eur")
	acc2 := createAccount(ctx, ak, "acc2", "6500usd,1200gbp")

	// Restrict trading of gbp
	k.authorityk = dummyAuthority{
		RestrictedDenoms: []emtypes.RestrictedDenom{
			{"gbp", []sdk.AccAddress{acc1.GetAddress()}},
		}}

	k.initializeFromStore(ctx)

	{ // Verify that acc2 can't create a passive gbp order
		o := order(acc2, "500gbp", "542eur")
		_, err := k.NewOrderSingle(ctx, o)
		require.NoError(t, err)
		//require.Empty(t, k.instruments)

		o = order(acc2, "542usd", "500gbp")
		_, err = k.NewOrderSingle(ctx, o)
		require.NoError(t, err)
		//require.Empty(t, k.instruments)
	}

	{ // Verify that acc1 can create a passive gbp order
		o := order(acc1, "542eur", "500gbp")
		_, err := k.NewOrderSingle(ctx, o)
		require.NoError(t, err)
		//require.Len(t, k.instruments, 1)

		o = order(acc1, "200gbp", "333usd")
		_, err = k.NewOrderSingle(ctx, o)
		require.NoError(t, err)
		//require.Len(t, k.instruments, 2)
	}

	{ // Verify that acc2 managed to sell its gbp to a passive order
		o := order(acc2, "500gbp", "542eur")
		_, err := k.NewOrderSingle(ctx, o)
		require.NoError(t, err)

		balance := ak.GetAccount(ctx, acc2.GetAddress()).GetCoins()
		require.Equal(t, "542", balance.AmountOf("eur").String())

		o = order(acc2, "333usd", "200gbp")
		_, err = k.NewOrderSingle(ctx, o)
		require.NoError(t, err)
		balance = ak.GetAccount(ctx, acc2.GetAddress()).GetCoins()
		require.Equal(t, "900", balance.AmountOf("gbp").String())
	}
}

func TestRestrictedDenominations2(t *testing.T) {
	// Two instruments are restricted.
	ctx, k, ak, _, _ := createTestComponents(t)

	acc1 := createAccount(ctx, ak, "acc1", "5000gbp, 10000usd")

	// Restrict trading of gbp and usd
	k.authorityk = dummyAuthority{
		RestrictedDenoms: []emtypes.RestrictedDenom{
			{"gbp", []sdk.AccAddress{}},
			{"usd", []sdk.AccAddress{acc1.GetAddress()}},
		}}

	k.initializeFromStore(ctx)

	gasMeter := sdk.NewGasMeter(math.MaxUint64)
	// Ensure that no orders can be created, even though acc1 is allowed to create usd orders
	o := order(acc1, "542usd", "500gbp")
	_, err := k.NewOrderSingle(ctx.WithGasMeter(gasMeter), o)
	require.NoError(t, err)
	//require.Empty(t, k.instruments)
	require.Equal(t, gasPriceNewOrder, gasMeter.GasConsumed())

	o = order(acc1, "500gbp", "542usd")
	_, err = k.NewOrderSingle(ctx, o)
	require.NoError(t, err)
	//require.Empty(t, k.instruments)
}

func TestSyntheticInstruments1(t *testing.T) {
	ctx, k, ak, _, _ := createTestComponents(t)
	acc1 := createAccount(ctx, ak, "acc1", "5000eur")
	acc2 := createAccount(ctx, ak, "acc2", "6500usd")
	acc3 := createAccount(ctx, ak, "acc3", "4500chf")

	totalSupply := snapshotAccounts(ctx, ak)

	o := order(acc1, "1000eur", "1114usd")
	_, err := k.NewOrderSingle(ctx, o)
	require.NoError(t, err)

	o = order(acc1, "500eur", "542chf")
	_, err = k.NewOrderSingle(ctx, o)
	require.NoError(t, err)

	o = order(acc3, "1000chf", "1028usd")
	_, err = k.NewOrderSingle(ctx, o)
	require.NoError(t, err)

	gasMeter := sdk.NewGasMeter(math.MaxUint64)
	o = order(acc2, "5000usd", "4485eur")
	_, err = k.NewOrderSingle(ctx.WithGasMeter(gasMeter), o)
	require.NoError(t, err)
	require.Equal(t, gasMeter.GasConsumed(), gasPriceNewOrder) // Matches several orders, but should pay only the fixed fee

	// Ensure acc2 received at least some euro
	acc2Balance := ak.GetAccount(ctx, acc2.GetAddress()).GetCoins()
	require.True(t, acc2Balance.AmountOf("eur").IsPositive())

	// Ensure acc2 did not receive any CHF, which is used in the synthetic instrument
	require.True(t, acc2Balance.AmountOf("chf").IsZero())

	// Ensure that acc2 filled all the eur sale orders in the market.
	require.True(t, acc2Balance.AmountOf("eur").Equal(sdk.NewInt(1500)))

	// Ensure that all tokens are accounted for.
	require.True(t, totalSupply.Sub(snapshotAccounts(ctx, ak)).IsZero())
}

func TestNonMatchingOrders(t *testing.T) {
	ctx, k, ak, _, _ := createTestComponents(t)
	acc1 := createAccount(ctx, ak, "acc1", "100000usd")
	acc2 := createAccount(ctx, ak, "acc2", "240000eur")

	_, err := k.NewOrderSingle(ctx, order(acc1, "20000usd", "20000eur"))
	require.NoError(t, err)
	_, err = k.NewOrderSingle(ctx, order(acc2, "20000eur", "50000usd"))
	require.NoError(t, err)

	acc1Orders := k.GetOrdersByOwner(ctx, acc1.GetAddress())
	require.Len(t, acc1Orders, 1)
	require.Equal(t, sdk.ZeroInt(), acc1Orders[0].DestinationFilled)
	require.Equal(t, sdk.ZeroInt(), acc1Orders[0].SourceFilled)

	acc2Orders := k.GetOrdersByOwner(ctx, acc2.GetAddress())
	require.Len(t, acc2Orders, 1)
	require.Equal(t, sdk.ZeroInt(), acc2Orders[0].DestinationFilled)
	require.Equal(t, sdk.ZeroInt(), acc2Orders[0].SourceFilled)
}

func TestSyntheticInstruments2(t *testing.T) {
	ctx, k, ak, _, _ := createTestComponents(t)
	acc1 := createAccount(ctx, ak, "acc1", "972000chf,5000000usd")
	acc2 := createAccount(ctx, ak, "acc2", "765000gbp,108000000jpy")

	acc3 := createAccount(ctx, ak, "acc3", "3700000eur")

	totalSupply := snapshotAccounts(ctx, ak)

	passiveOrders := []types.Order{
		order(acc1, "1000000usd", "896000eur"),

		order(acc1, "1000000usd", "972000chf"),
		order(acc1, "972000chf", "897000eur"),

		order(acc1, "1000000usd", "108000000jpy"),
		order(acc2, "40000000jpy", "331000eur"),
		order(acc2, "68000000jpy", "563000eur"),

		order(acc1, "400000usd", "306000gbp"),
		order(acc1, "600000usd", "459000gbp"),
		order(acc2, "765000gbp", "896000eur"),
	}

	for _, o := range passiveOrders {
		res, err := k.NewOrderSingle(ctx, o)
		require.NoError(t, err, res.Log)
	}

	gasMeter := sdk.NewGasMeter(math.MaxUint64)
	monsterOrder := order(acc3, "3700000eur", "4000000usd")
	res, err := k.NewOrderSingle(ctx.WithGasMeter(gasMeter), monsterOrder)
	require.NoError(t, err, res.Log)
	require.Equal(t, gasPriceNewOrder, gasMeter.GasConsumed())

	//require.Len(t, k.instruments, 0)

	acc3bal := ak.GetAccount(ctx, acc3.GetAddress()).GetCoins()
	require.Equal(t, "4000000", acc3bal.AmountOf("usd").String())

	// Ensure that all tokens are accounted for.
	require.True(t, totalSupply.Sub(snapshotAccounts(ctx, ak)).IsZero())
}

func TestDestinationCapacity(t *testing.T) {
	ctx, k, ak, _, _ := createTestComponents(t)

	acc1 := createAccount(ctx, ak, "acc1", "900000000usd")
	acc2 := createAccount(ctx, ak, "acc2", "500000000000eur")

	order1 := order(acc1, "800000000usd", "720000000eur")
	order1.SourceRemaining = sdk.NewInt(182000000)
	order1.SourceFilled = sdk.NewInt(618000000)
	order1.DestinationFilled = sdk.NewInt(645161290)

	_, err := k.NewOrderSingle(ctx, order1)
	require.NoError(t, err)

	order2 := order(acc2, "471096868463eur", "500182000000usd")
	_, err = k.NewOrderSingle(ctx, order2)
	require.NoError(t, err)
}

func TestDestinationCapacity2(t *testing.T) {
	ctx, k, ak, _, _ := createTestComponents(t)

	acc1 := createAccount(ctx, ak, "acc1", "900000000usd")
	acc2 := createAccount(ctx, ak, "acc2", "500000000000eur")
	acc3 := createAccount(ctx, ak, "acc3", "140000000000chf")

	// chf -> usd -> eur

	order1 := order(acc1, "800000000usd", "720000000eur")
	order1.SourceRemaining = sdk.NewInt(182000000)
	order1.SourceFilled = sdk.NewInt(618000000)
	order1.DestinationFilled = sdk.NewInt(645161290)

	_, err := k.NewOrderSingle(ctx, order1)
	require.NoError(t, err)

	order2 := order(acc3, "130000000000chf", "800000000usd")
	_, err = k.NewOrderSingle(ctx, order2)
	require.NoError(t, err)

	aggressiveOrder := order(acc2, "471096868463eur", "120000000000chf")
	_, err = k.NewOrderSingle(ctx, aggressiveOrder)
	require.NoError(t, err)
}

func TestPreventPhantomLiquidity(t *testing.T) {
	ctx, k, ak, _, _ := createTestComponents(t)

	acc1 := createAccount(ctx, ak, "acc1", "10000eur")

	order1 := order(acc1, "8000eur", "9000usd")
	_, err := k.NewOrderSingle(ctx, order1)
	require.NoError(t, err)

	// Cannot sell more than the balance in the same instrument
	order2 := order(acc1, "8000eur", "9000usd")
	_, err = k.NewOrderSingle(ctx, order2)
	require.Error(t, err)

	// Can sell the balance in another instrument
	order3 := order(acc1, "8000eur", "6000chf")
	_, err = k.NewOrderSingle(ctx, order3)
	require.NoError(t, err)
}

func printTotalBalance(accs ...authexported.Account) {
	sum := sdk.NewCoins()

	for _, acc := range accs {
		sum = sum.Add(acc.GetCoins()...)
	}

	fmt.Println(sum)
}

func createTestComponents(t *testing.T) (sdk.Context, *Keeper, auth.AccountKeeper, bank.Keeper, supply.Keeper) {
	var (
		keyMarket  = sdk.NewKVStoreKey(types.ModuleName)
		keyIndices = sdk.NewKVStoreKey(types.StoreKeyIdx)
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

	pk := params.NewKeeper(types.ModuleCdc, keyParams, tkeyParams)

	ctx := sdk.NewContext(ms, abci.Header{ChainID: "test-chain"}, true, log.NewNopLogger())
	accountKeeper := auth.NewAccountKeeper(types.ModuleCdc, authCapKey, pk.Subspace(auth.DefaultParamspace), auth.ProtoBaseAccount)
	accountKeeperWrapped := emauth.Wrap(accountKeeper)

	bankKeeper := bank.NewBaseKeeper(accountKeeperWrapped, pk.Subspace(bank.DefaultParamspace), blacklistedAddrs)

	maccPerms := map[string][]string{}

	supplyKeeper := supply.NewKeeper(types.ModuleCdc, supplyKey, accountKeeper, bankKeeper, maccPerms)
	supplyKeeper.SetSupply(ctx, supply.NewSupply(coins("1eur,1usd,1chf,1jpy,1gbp")))

	marketKeeper := NewKeeper(types.ModuleCdc, keyMarket, keyIndices, accountKeeperWrapped, bankKeeper, supplyKeeper, dummyAuthority{})

	return ctx, marketKeeper, accountKeeper, bankKeeper, supplyKeeper
}

type dummyAuthority struct {
	RestrictedDenoms emtypes.RestrictedDenoms
}

func (da dummyAuthority) GetRestrictedDenoms(sdk.Context) emtypes.RestrictedDenoms {
	return da.RestrictedDenoms
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

func order(account authexported.Account, src, dst string) types.Order {
	o, err := types.NewOrder(coin(src), coin(dst), account.GetAddress(), time.Now(), cid())
	if err != nil {
		panic(err)
	}

	return o
}

func createAccount(ctx sdk.Context, ak auth.AccountKeeper, address, balance string) authexported.Account {
	acc := ak.NewAccountWithAddress(ctx, sdk.AccAddress([]byte(address)))
	acc.SetCoins(coins(balance))
	ak.SetAccount(ctx, acc)
	return acc
}

// Generate a random string to use as a client order id
func cid() string {
	return tmrand.Str(10)
}

func dumpEvents(events sdk.Events) {
	fmt.Println("Number of events:", len(events))
	for _, evt := range events {
		fmt.Println(evt.Type)
		for _, kv := range evt.Attributes {
			fmt.Println(" - ", string(kv.Key), string(kv.Value))
		}
	}

}

func snapshotAccounts(ctx sdk.Context, ak auth.AccountKeeper) (totalBalance sdk.Coins) {
	ak.IterateAccounts(ctx, func(acc authexported.Account) (stop bool) {
		totalBalance = totalBalance.Add(acc.GetCoins()...)
		return
	})
	return
}
