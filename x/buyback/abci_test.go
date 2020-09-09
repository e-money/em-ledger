package buyback

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	emauth "github.com/e-money/em-ledger/hooks/auth"
	apptypes "github.com/e-money/em-ledger/types"
	"github.com/e-money/em-ledger/x/buyback/internal/keeper"
	"github.com/e-money/em-ledger/x/market"
	"github.com/e-money/em-ledger/x/market/types"

	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authexported "github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/supply"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmrand "github.com/tendermint/tendermint/libs/rand"
	dbm "github.com/tendermint/tm-db"
)

const (
	stakingDenom = "ungm"
)

func TestBuyback1(t *testing.T) {
	ctx, keeper, market, accountKeeper, bankKeeper, supplyKeeper := createTestComponents(t)

	ctx = ctx.WithBlockHeight(1)
	generateMarketActivity(ctx, market, accountKeeper)

	account := supplyKeeper.GetModuleAccount(ctx, ModuleName)
	bankKeeper.AddCoins(ctx, account.GetAddress(), coins("10000ungm"))

	BeginBlocker(ctx, keeper)

	// Verify that staking tokens are burned
	account = supplyKeeper.GetModuleAccount(ctx, ModuleName)
	require.True(t, account.GetCoins().AmountOf(stakingDenom).IsZero())

	require.Condition(t, func() bool {
		for _, evt := range ctx.EventManager().Events() {
			if evt.Type == EventTypeBuybackBurn {
				return true
			}
		}

		return false
	}, "Burn event not found")

	// Verify that an order was created in the market
	orders := market.GetOrdersByOwner(ctx, account.GetAddress())
	require.Len(t, orders, 1)
	require.Equal(t, sdk.NewCoin("eur", sdk.NewInt(50000)), orders[0].Source)
	require.True(t, strings.HasSuffix(orders[0].ClientOrderID, "1"))

	ctx = ctx.WithBlockHeight(2).WithBlockTime(ctx.BlockTime().Add(2 * time.Hour))
	// Add more stablecoin tokens to the module's buyback account
	bankKeeper.AddCoins(ctx, account.GetAddress(), coins("10000eur,75000chf"))

	// Update account balance information
	account = supplyKeeper.GetModuleAccount(ctx, ModuleName)
	BeginBlocker(ctx, keeper)

	orders = market.GetOrdersByOwner(ctx, account.GetAddress())
	require.Len(t, orders, 2)

	require.Equal(t, account.GetCoins().AmountOf("chf"), orders[0].Source.Amount)
	require.Equal(t, account.GetCoins().AmountOf("eur"), orders[1].Source.Amount)

	require.True(t, strings.HasSuffix(orders[0].ClientOrderID, "2"))

	// Verify prices.
	require.Equal(t, sdk.NewDecWithPrec(25, 2), orders[0].Price())
	require.Equal(t, sdk.NewDecWithPrec(5, 1), orders[1].Price())
}

func TestBuyback2(t *testing.T) {
	// Verify that the module does not update its market positions for every block
	ctx, keeper, market, accountKeeper, bankKeeper, supplyKeeper := createTestComponents(t)

	ctx = ctx.WithBlockHeight(1)
	generateMarketActivity(ctx, market, accountKeeper)

	account := supplyKeeper.GetModuleAccount(ctx, ModuleName)

	orders := market.GetOrdersByOwner(ctx, account.GetAddress())
	require.Empty(t, orders)

	BeginBlocker(ctx, keeper)

	orders = market.GetOrdersByOwner(ctx, account.GetAddress())
	require.Len(t, orders, 1)

	// New balance that should trigger order updates
	bankKeeper.AddCoins(ctx, account.GetAddress(), coins("10000eur,75000chf"))

	// Time since last update is too short to update the module's market positions.
	ctx = ctx.WithBlockHeight(2).WithBlockTime(ctx.BlockTime().Add(30 * time.Minute))
	BeginBlocker(ctx, keeper)

	orders = market.GetOrdersByOwner(ctx, account.GetAddress())
	require.Len(t, orders, 1)
	require.True(t, strings.HasSuffix(orders[0].ClientOrderID, "1"))

	// Time since last update is sufficient. Market positions must be updated.
	ctx = ctx.WithBlockHeight(3).WithBlockTime(ctx.BlockTime().Add(time.Hour))
	BeginBlocker(ctx, keeper)

	orders = market.GetOrdersByOwner(ctx, account.GetAddress())
	require.Len(t, orders, 2)
	require.True(t, strings.HasSuffix(orders[0].ClientOrderID, "3"))
}

func TestBuyback3(t *testing.T) {
	// Test very high NGM price with very low balance
	ctx, keeper, market, accountKeeper, bankKeeper, supplyKeeper := createTestComponents(t)
	account := supplyKeeper.GetModuleAccount(ctx, ModuleName)
	bankKeeper.AddCoins(ctx, account.GetAddress(), coins("50pesos"))

	// Generate some prices for the pesos <-> ungm instrument
	var (
		acc1 = createAccount(ctx, accountKeeper, "acc1", "5000000pesos")
		acc2 = createAccount(ctx, accountKeeper, "acc2", "10000ungm")
	)
	_, err := market.NewOrderSingle(ctx, order(acc1, "4000000pesos", "1ungm"))
	require.NoError(t, err)

	_, err = market.NewOrderSingle(ctx, order(acc2, "1ungm", "4000000pesos"))
	require.NoError(t, err)

	// Attempt to create a position using the meager pesos balance of the module
	BeginBlocker(ctx, keeper)

	account = supplyKeeper.GetModuleAccount(ctx, ModuleName)
	require.Equal(t, "50", account.GetCoins().AmountOf("pesos").String())

	orders := market.GetOrdersByOwner(ctx, account.GetAddress())
	require.Empty(t, orders)
}

func TestGroupMarketData(t *testing.T) {
	price := func(i int64) *sdk.Dec {
		r := sdk.NewDec(i)
		return &r
	}

	md := []types.MarketData{
		{
			Source:      "eur",
			Destination: "ungm",
		},
		{
			Source:      "chf",
			Destination: "ungm",
		},
		{
			Source:      "eur",
			Destination: "chf",
			LastPrice:   price(5),
		},
		{
			Source:      "ungm",
			Destination: "chf",
			LastPrice:   price(8),
		},
	}

	groupedOrders := groupMarketDataBySource(md, "ungm")

	require.Len(t, groupedOrders, 2)
	require.Contains(t, groupedOrders, "eur")
	require.Contains(t, groupedOrders, "chf")

}

// Create some basic pricing information that market orders can be made from
func generateMarketActivity(ctx sdk.Context, marketKeeper *market.Keeper, accounts auth.AccountKeeper) {
	var (
		acc1 = createAccount(ctx, accounts, "acc1", "50000ungm")
		acc2 = createAccount(ctx, accounts, "acc2", "150000eur,290000chf")
	)

	marketKeeper.NewOrderSingle(ctx, order(acc1, "5000ungm", "10000eur"))
	marketKeeper.NewOrderSingle(ctx, order(acc1, "5000ungm", "20000chf"))
	marketKeeper.NewOrderSingle(ctx, order(acc2, "10000eur", "5000ungm"))
	marketKeeper.NewOrderSingle(ctx, order(acc2, "20000chf", "5000ungm"))
}

func order(account authexported.Account, src, dst string) types.Order {
	s, _ := sdk.ParseCoin(src)
	d, _ := sdk.ParseCoin(dst)
	o, err := types.NewOrder(types.TimeInForce_GoodTilCancel, s, d, account.GetAddress(), tmrand.Str(10))
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

func createTestComponents(t *testing.T) (sdk.Context, keeper.Keeper, *market.Keeper, auth.AccountKeeper, bank.Keeper, supply.Keeper) {
	var (
		keyMarket  = sdk.NewKVStoreKey(types.ModuleName)
		keyIndices = sdk.NewKVStoreKey(types.StoreKeyIdx)
		authCapKey = sdk.NewKVStoreKey("authCapKey")
		keyParams  = sdk.NewKVStoreKey("params")
		supplyKey  = sdk.NewKVStoreKey("supply")
		buybackKey = sdk.NewKVStoreKey("buyback")

		tkeyParams = sdk.NewTransientStoreKey("transient_params")

		blacklistedAddrs = make(map[string]bool)
	)

	db := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(authCapKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyMarket, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyIndices, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(supplyKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(buybackKey, sdk.StoreTypeIAVL, db)

	err := ms.LoadLatestVersion()
	require.Nil(t, err)

	pk := params.NewKeeper(types.ModuleCdc, keyParams, tkeyParams)

	ctx := sdk.NewContext(ms, abci.Header{ChainID: "test-chain", Time: time.Now()}, true, log.NewNopLogger())
	accountKeeper := auth.NewAccountKeeper(types.ModuleCdc, authCapKey, pk.Subspace(auth.DefaultParamspace), auth.ProtoBaseAccount)
	accountKeeperWrapped := emauth.Wrap(accountKeeper)

	bankKeeper := bank.NewBaseKeeper(accountKeeperWrapped, pk.Subspace(bank.DefaultParamspace), blacklistedAddrs)

	maccPerms := map[string][]string{
		AccountName: {supply.Burner},
	}

	supplyKeeper := supply.NewKeeper(types.ModuleCdc, supplyKey, accountKeeper, bankKeeper, maccPerms)

	initialSupply := coins(fmt.Sprintf("1000000eur,1000000usd,1000000chf,1000000jpy,1000000gbp,1000000%v,500000000pesos", stakingDenom))
	supplyKeeper.SetSupply(ctx, supply.NewSupply(initialSupply))

	marketKeeper := market.NewKeeper(types.ModuleCdc, keyMarket, keyIndices, accountKeeperWrapped, bankKeeper, supplyKeeper, mockAuthority{})

	keeper := NewKeeper(types.ModuleCdc, buybackKey, marketKeeper, supplyKeeper, mockStakingKeeper{})

	// Deposit a working balance on the buyback module account.
	buybackAccount := supplyKeeper.GetModuleAccount(ctx, ModuleName)
	bankKeeper.SetCoins(ctx, buybackAccount.GetAddress(), coins("50000eur"))

	return ctx, keeper, marketKeeper, accountKeeper, bankKeeper, supplyKeeper
}

func coins(c string) sdk.Coins {
	coins, err := sdk.ParseCoins(c)
	if err != nil {
		panic(err)
	}

	return coins
}

var (
	_ types.RestrictedKeeper = (*mockAuthority)(nil)
	_ keeper.StakingKeeper   = (*mockStakingKeeper)(nil)
)

type (
	mockAuthority     struct{}
	mockStakingKeeper struct{}
)

func (mockStakingKeeper) BondDenom(sdk.Context) string {
	return stakingDenom
}

func (mockAuthority) GetRestrictedDenoms(sdk.Context) apptypes.RestrictedDenoms {
	return nil
}
