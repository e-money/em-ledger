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

func TestCirculating(t *testing.T) {
	ctx := sdk.Context{}.WithContext(context.Background())
	enc := simapp.MakeTestEncodingConfig()
	queryHelper := baseapp.NewQueryServerTestHelper(ctx, enc.InterfaceRegistry)
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

	// Test that supply has been initialized as expected
	require.Equal(t, "453", bkMock.GetSupply(ctx, stakingDenom).Amount.String())
	require.Equal(t, "154", bkMock.GetSupply(ctx,"blx").Amount.String())

	types.RegisterQueryServer(queryHelper, NewQuerier(&accountKeeper, bkMock))
	queryClient := types.NewQueryClient(queryHelper)

	gotRsp, err := queryClient.Circulating(sdk.WrapSDKContext(ctx), &types.QueryCirculatingRequest{})
	require.NoError(t, err)
	assert.Equal(t, mustParseCoins("154blx,3"+stakingDenom).String(), gotRsp.Total.String())
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
			Base:        "blx",
		},
		{
			Base:        stakingDenom,
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
	_ AccountKeeper = &accountKeeperMock{}
	_ BankKeeper    = &bankKeeperMock{}
)
