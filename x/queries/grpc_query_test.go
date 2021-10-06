package queries

import (
	"context"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/e-money/em-ledger/x/queries/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/rand"
	"testing"
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
			acc1.String(): sdk.NewCoins(sdk.NewCoin("alx", sdk.OneInt())),
			acc2.String(): sdk.NewCoins(sdk.NewCoin("alx", sdk.NewInt(2))),
			acc3.String(): sdk.NewCoins(sdk.NewCoin("blx", sdk.NewInt(4))),
			authtypes.NewEmptyModuleAccount(stakingtypes.BondedPoolName).Address:    sdk.NewCoins(sdk.NewCoin("alx", sdk.NewInt(100))),
			authtypes.NewEmptyModuleAccount(stakingtypes.NotBondedPoolName).Address: sdk.NewCoins(sdk.NewCoin("alx", sdk.NewInt(200))),
		},
	}
	types.RegisterQueryServer(queryHelper, NewQuerier(&accountKeeper, bkMock))
	queryClient := types.NewQueryClient(queryHelper)

	specs := map[string]struct {
		accts    []authtypes.AccountI
		expErr   bool
		expState sdk.Coins
	}{
		"sum accounts of same denom": {
			accts: []authtypes.AccountI{
				authtypes.NewBaseAccountWithAddress(acc1),
				authtypes.NewBaseAccountWithAddress(acc2),
			},
			expState: sdk.NewCoins(sdk.NewCoin("alx", sdk.NewInt(3))),
		},
		"do not sum different denoms": {
			accts: []authtypes.AccountI{
				authtypes.NewBaseAccountWithAddress(acc1),
				authtypes.NewBaseAccountWithAddress(acc3),
			},
			expState: sdk.NewCoins(sdk.NewCoin("alx", sdk.NewInt(1)), sdk.NewCoin("blx", sdk.NewInt(4))),
		},
		"do not sum bounded and unbounded pools": {
			accts: []authtypes.AccountI{
				authtypes.NewEmptyModuleAccount(stakingtypes.BondedPoolName),
				authtypes.NewEmptyModuleAccount(stakingtypes.NotBondedPoolName),
				authtypes.NewBaseAccountWithAddress(acc1),
			},
			expState: sdk.NewCoins(sdk.NewCoin("alx", sdk.NewInt(1))),
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			accountKeeper.accts = spec.accts
			gotRsp, gotErr := queryClient.Circulating(sdk.WrapSDKContext(ctx), &types.QueryCirculatingRequest{})
			if spec.expErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
			assert.Equal(t, spec.expState, gotRsp.Total)
		})
	}
}

type accountKeeperMock struct {
	accts []authtypes.AccountI
}

func (a accountKeeperMock) IterateAccounts(ctx sdk.Context, process func(authtypes.AccountI) bool) {
	for _, a := range a.accts {
		if ok := process(a); ok {
			return
		}
	}
}

type bankKeeperMock struct {
	balances map[string]sdk.Coins
}

func (b bankKeeperMock) SpendableCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins {
	return b.balances[addr.String()]
}
