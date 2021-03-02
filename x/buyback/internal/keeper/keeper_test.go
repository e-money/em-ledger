package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	dbm "github.com/tendermint/tm-db"
	"testing"
	"time"
)

func TestUpdateBuybackMarket(t *testing.T) {
	ctx, keeper := setupKeeper(t)
	keeper.SetUpdateInterval(ctx, time.Second)

	now := time.Now()
	ok := keeper.UpdateBuybackMarket(ctx.WithBlockTime(now))
	require.True(t, ok) // empty state
	ok = keeper.UpdateBuybackMarket(ctx.WithBlockTime(now.Add(time.Second)))
	require.True(t, ok) // after interval
	ok = keeper.UpdateBuybackMarket(ctx.WithBlockTime(now.Add(time.Second - time.Nanosecond)))
	require.False(t, ok) // before interval ends
}

func setupKeeper(t *testing.T) (sdk.Context, Keeper) {
	var (
		buybackKey = sdk.NewKVStoreKey("buyback")
	)

	db := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(buybackKey, sdk.StoreTypeIAVL, db)

	err := ms.LoadLatestVersion()
	require.Nil(t, err)

	ctx := sdk.NewContext(ms, tmproto.Header{ChainID: "test-chain"}, true, log.NewNopLogger())
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	marshaler := codec.NewProtoCodec(interfaceRegistry)

	keeper := NewKeeper(marshaler, buybackKey, nil, nil, nil, nil)
	return ctx, keeper
}
