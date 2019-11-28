package keeper

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store"
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

func Test1(t *testing.T) {
	ctx, k, ak := createTestComponents(t)

	acc1 := ak.NewAccountWithAddress(ctx, sdk.AccAddress([]byte("acc1")))
	acc2 := ak.NewAccountWithAddress(ctx, sdk.AccAddress([]byte("acc2")))

	src := sdk.NewCoin("eur", sdk.NewInt(100))
	dst := sdk.NewCoin("usd", sdk.NewInt(120))
	order := types.NewOrder(src, dst, acc1.GetAddress())
	k.ProcessOrder(ctx, order)

	fmt.Println(k.instruments.String())

	src = sdk.NewCoin("usd", sdk.NewInt(60))
	dst = sdk.NewCoin("eur", sdk.NewInt(50))
	order = types.NewOrder(src, dst, acc2.GetAddress())
	k.ProcessOrder(ctx, order)

	fmt.Println(k.instruments.String())

	i := k.instruments[0]
	remainingOrder := i.Orders.Peek().(*types.Order)
	fmt.Println(remainingOrder)
}

func Test2(t *testing.T) {
	ctx, k, ak := createTestComponents(t)

	acc1 := ak.NewAccountWithAddress(ctx, sdk.AccAddress([]byte("acc1")))
	acc2 := ak.NewAccountWithAddress(ctx, sdk.AccAddress([]byte("acc2")))

	src := sdk.NewCoin("eur", sdk.NewInt(100))
	dst := sdk.NewCoin("usd", sdk.NewInt(120))
	order := types.NewOrder(src, dst, acc1.GetAddress())
	k.ProcessOrder(ctx, order)

	fmt.Println(k.instruments.String())

	src = sdk.NewCoin("usd", sdk.NewInt(121))
	dst = sdk.NewCoin("eur", sdk.NewInt(100))
	order = types.NewOrder(src, dst, acc2.GetAddress())
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
