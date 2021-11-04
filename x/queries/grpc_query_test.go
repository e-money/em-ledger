package queries

import (
	"context"
	"testing"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/e-money/em-ledger/x/queries/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/rand"
)

func newQServer() (context.Context, sdk.Context, types.QueryClient, bankKeeperMock) {
	sdkCtx := sdk.Context{}.WithContext(context.Background())
	ctx := sdk.WrapSDKContext(sdkCtx)
	enc := simapp.MakeTestEncodingConfig()
	queryHelper := baseapp.NewQueryServerTestHelper(sdkCtx, enc.InterfaceRegistry)
	accountKeeper := accountKeeperMock{}

	const legAddrLen = 20
	var (
		acc1 sdk.AccAddress = rand.Bytes(legAddrLen)
		acc2 sdk.AccAddress = rand.Bytes(legAddrLen)
		acc3 sdk.AccAddress = rand.Bytes(legAddrLen)
	)

	bkMock := bankKeeperMock{
		balances: map[string]sdk.Coins{
			acc1.String(): mustParseCoins("1" + stakingDenom),
			acc2.String(): mustParseCoins("2" + stakingDenom),
			acc3.String(): mustParseCoins("4blx"),
			authtypes.NewEmptyModuleAccount(stakingtypes.BondedPoolName).Address:    mustParseCoins("100" + stakingDenom),
			authtypes.NewEmptyModuleAccount(stakingtypes.NotBondedPoolName).Address: mustParseCoins("200" + stakingDenom),
		},
		vesting: mustParseCoins("150blx,150" + stakingDenom),
	}

	skMock := slashingKeeperMock{
		missedBlocksMap: map[string]types.MissedBlocksInfo{
			"cosmosvalcons1g0t3yc0twz8d2ex05ek0gsv57edgmx6mnxkzlu": {
				MissedBlocksCounter: 1,
				TotalBlocksCounter:  10,
			},
		},
	}

	types.RegisterQueryServer(
		queryHelper, NewQuerier(&accountKeeper, bkMock, skMock),
	)
	queryClient := types.NewQueryClient(queryHelper)

	return ctx, sdkCtx, queryClient, bkMock
}

func TestCirculating(t *testing.T) {
	ctx, sdkCtx, queryClient, bkMock := newQServer()

	// Test that supply has been initialized as expected
	require.Equal(t, "453", bkMock.GetSupply(sdkCtx, stakingDenom).Amount.String())
	require.Equal(t, "154", bkMock.GetSupply(sdkCtx, "blx").Amount.String())

	gotRsp, err := queryClient.Circulating(ctx, &types.QueryCirculatingRequest{})
	require.NoError(t, err)
	assert.Equal(t, mustParseCoins("154blx,3"+stakingDenom).String(), gotRsp.Total.String())
}

func TestMissedBlocks(t *testing.T) {
	ctx, _, queryClient, _ := newQServer()

	var zero64 int64
	gotMBRsp, err := queryClient.MissedBlocks(
		ctx, &types.QueryMissedBlocksRequest{
			ConsAddress: "cosmosvalcons10e4c5p6qk0sycy9u6u43t7csmlx9fyadr9yxph",
		},
	)

	require.NoError(t, err)
	assert.Equal(t, zero64, gotMBRsp.MissedBlocksInfo.MissedBlocksCounter)
	assert.Equal(t, zero64, gotMBRsp.MissedBlocksInfo.TotalBlocksCounter)

	gotMBRsp, err = queryClient.MissedBlocks(
		ctx, &types.QueryMissedBlocksRequest{
			ConsAddress: "cosmosvalcons1g0t3yc0twz8d2ex05ek0gsv57edgmx6mnxkzlu",
		},
	)

	require.NoError(t, err)
	assert.Equal(t, int64(1), gotMBRsp.MissedBlocksInfo.MissedBlocksCounter)
	assert.Equal(t, int64(10), gotMBRsp.MissedBlocksInfo.TotalBlocksCounter)
}

func mustParseCoins(s string) sdk.Coins {
	if c, err := sdk.ParseCoinsNormalized(s); err == nil {
		return c
	} else {
		panic(err)
	}
}

type accountKeeperMock struct{}

func (a accountKeeperMock) GetModuleAccount(_ sdk.Context, moduleName string) authtypes.ModuleAccountI {
	return authtypes.NewEmptyModuleAccount(moduleName)
}

type slashingKeeperMock struct {
	missedBlocksMap map[string]types.MissedBlocksInfo
}

func (s slashingKeeperMock) GetMissedBlocks(_ sdk.Context, consAddr sdk.ConsAddress) (int64, int64) {
	return s.missedBlocksMap[consAddr.String()].MissedBlocksCounter,
		s.missedBlocksMap[consAddr.String()].TotalBlocksCounter
}

type bankKeeperMock struct {
	balances map[string]sdk.Coins
	vesting  sdk.Coins
}

func (b bankKeeperMock) IterateAllDenomMetaData(
	ctx sdk.Context, cb func(banktypes.Metadata) bool,
) {
	panic("implement me")
}

func (b bankKeeperMock) GetAllDenomMetaData(_ sdk.Context) []banktypes.Metadata {
	return []banktypes.Metadata{
		{
			Base: "blx",
		},
		{
			Base: stakingDenom,
		},
	}
}

func (b bankKeeperMock) SpendableCoins(_ sdk.Context, addr sdk.AccAddress) sdk.Coins {
	return b.balances[addr.String()]
}

func (b bankKeeperMock) GetSupply(_ sdk.Context, denom string) sdk.Coin {
	var supply = sdk.NewCoin(denom, sdk.ZeroInt())
	for _, balance := range b.balances {
		amnt := balance.AmountOfNoDenomValidation(denom)
		supply = supply.Add(sdk.NewCoin(denom, amnt))
	}

	supply = supply.Add(sdk.NewCoin(denom, b.vesting.AmountOfNoDenomValidation(denom)))
	return supply
}

func (b bankKeeperMock) IterateAllBalances(_ sdk.Context, cb func(sdk.AccAddress, sdk.Coin) bool) {
	for address, balance := range b.balances {
		addr, err := sdk.AccAddressFromBech32(address)
		if err != nil {
			panic(err)
		}

		for _, coin := range balance {
			if cb(addr, coin) {
				return
			}
		}
	}
}

var (
	_ AccountKeeper  = &accountKeeperMock{}
	_ BankKeeper     = &bankKeeperMock{}
	_ SlashingKeeper = &slashingKeeperMock{}
)
