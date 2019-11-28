package keeper

import (
	"fmt"
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
	k.ProcessOrder(ctx, order)

	order = types.NewOrder(coin("60usd"), coin("50eur"), acc2.GetAddress())
	k.ProcessOrder(ctx, order)

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
	// TODO Test: Add order to queue. Reduce balance below order level. Ensure trade doesn't #€%! up.
	ctx, k, ak := createTestComponents(t)

	acc1 := createAccount(ctx, ak, "acc1", "500eur")
	acc2 := createAccount(ctx, ak, "acc2", "740usd")

	order := types.NewOrder(coin("300eur"), coin("360usd"), acc1.GetAddress())
	k.ProcessOrder(ctx, order)

	acc1.SetCoins(coins("250eur"))
	k.ak.SetAccount(ctx, acc1)

	// TODO Can order still be partially filled?

	order = types.NewOrder(coin("360usd"), coin("300eur"), acc2.GetAddress())
	k.ProcessOrder(ctx, order)

}

func Test2(t *testing.T) {
	ctx, k, ak := createTestComponents(t)

	acc1 := ak.NewAccountWithAddress(ctx, sdk.AccAddress([]byte("acc1")))
	acc2 := ak.NewAccountWithAddress(ctx, sdk.AccAddress([]byte("acc2")))

	order := types.NewOrder(coin("100eur"), coin("120usd"), acc1.GetAddress())
	k.ProcessOrder(ctx, order)

	fmt.Println(k.instruments.String())

	order = types.NewOrder(coin("121usd"), coin("100eur"), acc2.GetAddress())
	k.ProcessOrder(ctx, order)

	fmt.Println(k.instruments.String())

	i := k.instruments[0]
	remainingOrder := i.Orders.Peek().(*types.Order)
	fmt.Println(remainingOrder)
}

func Test3(t *testing.T) {
	ctx, k, ak := createTestComponents(t)

	acc1 := ak.NewAccountWithAddress(ctx, sdk.AccAddress([]byte("acc1")))
	acc2 := ak.NewAccountWithAddress(ctx, sdk.AccAddress([]byte("acc2")))

	src := sdk.NewCoin("eur", sdk.NewInt(100))
	dst := sdk.NewCoin("usd", sdk.NewInt(120))
	order := types.NewOrder(src, dst, acc1.GetAddress())
	k.ProcessOrder(ctx, order)

	fmt.Println(k.instruments.String())

	for i := 0; i < 4; i++ {
		src = sdk.NewCoin("usd", sdk.NewInt(30))
		dst = sdk.NewCoin("eur", sdk.NewInt(25))
		order = types.NewOrder(src, dst, acc2.GetAddress())
		k.ProcessOrder(ctx, order)
	}

	fmt.Println(k.instruments.String())

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
