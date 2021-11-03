package keeper

import (
	"context"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	ctypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/e-money/em-ledger/x/buyback/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	dbm "github.com/tendermint/tm-db"
	"testing"
	"time"
)

func TestQueryBalance(t *testing.T) {
	var (
		myBalance          = sdk.NewCoins(sdk.NewCoin("alx", sdk.OneInt()))
		myModuleAddress, _ = sdk.AccAddressFromBech32("emoney1kt0vh0ttget0xx77g6d3ttnvq2lnxx6vp3uyl0")
		interfaceRegistry  = codectypes.NewInterfaceRegistry()
	)

	ctx := sdk.Context{}.WithContext(context.Background())
	queryHelper := baseapp.NewQueryServerTestHelper(ctx, interfaceRegistry)

	bankMock := bankMock{
		balance: myBalance,
	}
	keeper := &Keeper{
		bankKeeper:     &bankMock,
		acccountKeeper: &accountKeeperMock{addr: myModuleAddress},
	}
	types.RegisterQueryServer(queryHelper, keeper)
	queryClient := types.NewQueryClient(queryHelper)

	specs := map[string]struct {
		req        *types.QueryBalanceRequest
		expBalance sdk.Coins
	}{
		"all good": {
			req:        &types.QueryBalanceRequest{},
			expBalance: myBalance,
		},
		"nil param": {
			expBalance: myBalance,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			gotRsp, gotErr := queryClient.Balance(sdk.WrapSDKContext(ctx), spec.req)
			require.NoError(t, gotErr)
			require.NotNil(t, gotRsp)
			assert.Equal(t, spec.expBalance, gotRsp.Balance)
			assert.Equal(t, myModuleAddress, bankMock.lastRecordedReqAddr)
		})
	}
}

func TestQueryBuybackTime(t *testing.T) {
	storeKey := sdk.NewKVStoreKey(types.ModuleName)
	db := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(storeKey, sdk.StoreTypeIAVL, db)
	ms.LoadLatestVersion()

	now := time.Now().UTC()
	updateInterval := 24 * time.Hour

	ctx := sdk.NewContext(ms, tmproto.Header{ChainID: "test-chain"}, true, log.NewNopLogger())
	ctx = ctx.WithBlockTime(now)

	keeper := &Keeper{
		cdc:      codec.NewProtoCodec(ctypes.NewInterfaceRegistry()),
		storeKey: storeKey,
	}

	keeper.SetUpdateInterval(ctx, updateInterval)
	keeper.UpdateBuybackMarket(ctx)

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, codectypes.NewInterfaceRegistry())
	types.RegisterQueryServer(queryHelper, keeper)

	queryClient := types.NewQueryClient(queryHelper)

	response, err := queryClient.BuybackTime(sdk.WrapSDKContext(ctx), &types.QueryBuybackTimeRequest{})
	require.NoError(t, err)
	require.NotNil(t, response)
	require.Equal(t, now, response.LastRunTime)
	require.Equal(t, now.Add(updateInterval), response.NextRunTime)
}

type bankMock struct {
	balance             sdk.Coins
	lastRecordedReqAddr sdk.AccAddress
}

func (b *bankMock) GetAllBalances(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins {
	b.lastRecordedReqAddr = addr
	return b.balance
}

func (b bankMock) GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	panic("not expected to be called")
}

func (b bankMock) BurnCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error {
	panic("not expected to be called")
}

type accountKeeperMock struct {
	addr sdk.AccAddress
}

func (a accountKeeperMock) GetModuleAddress(name string) sdk.AccAddress {
	return a.addr
}
