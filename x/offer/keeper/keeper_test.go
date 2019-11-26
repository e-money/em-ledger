package keeper

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/log"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/e-money/em-ledger/x/offer/types"

	abci "github.com/tendermint/tendermint/abci/types"
	dbm "github.com/tendermint/tm-db"
)

func TestName(t *testing.T) {
	ctx, k := createTestComponents(t)

	order := types.NewOrder("EUR", "USD", 100, 120)
	k.ProcessOrder(ctx, order)

	fmt.Println(k.instruments.String())

	order = types.NewOrder("USD", "EUR", 60, 50)
	k.ProcessOrder(ctx, order)

	fmt.Println(k.instruments.String())

	i := k.instruments[0]
	remainingOrder := i.Orders.Peek().(*types.Order)
	fmt.Println(remainingOrder)

}

func createTestComponents(t *testing.T) (sdk.Context, Keeper) {
	var keyOffer = sdk.NewKVStoreKey(types.ModuleName)

	cdc := makeTestCodec()

	db := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(keyOffer, sdk.StoreTypeIAVL, db)

	err := ms.LoadLatestVersion()
	require.Nil(t, err)
	ctx := sdk.NewContext(ms, abci.Header{ChainID: "test-chain"}, true, log.NewNopLogger())

	k := NewKeeper(cdc, keyOffer)

	return ctx, k
}

func makeTestCodec() (cdc *codec.Codec) {
	cdc = codec.New()
	return
}
