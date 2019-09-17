package keeper

import (
	"emoney/x/liquidityprovider/types"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/supply"
	"github.com/tendermint/tendermint/libs/log"
)

type Keeper struct {
	authKeeper   auth.AccountKeeper
	supplyKeeper supply.Keeper
}

func NewKeeper(ak auth.AccountKeeper, sk supply.Keeper) Keeper {
	return Keeper{
		authKeeper:   ak,
		supplyKeeper: sk,
	}
}

func (k Keeper) CreateLiquidityProvider(ctx sdk.Context, address sdk.AccAddress) {
	account := k.authKeeper.GetAccount(ctx, address)
	fmt.Println(" *** Retrieved account\n", account)

	credit := sdk.NewCoins(
		sdk.NewCoin("x2eur", sdk.NewIntWithDecimal(1000, 2)),
	)

	lpAcc := types.NewLiquidityProviderAccount(account, credit)

	k.authKeeper.SetAccount(ctx, lpAcc)
	fmt.Println(" *** Created LP account\n", lpAcc)
}

func (k Keeper) MintTokensFromCredit(ctx sdk.Context, liquidityProvider sdk.AccAddress, amount sdk.Coins) {
	fmt.Println(" *** Mint tokens in keeper")
	logger := k.Logger(ctx)

	a := k.authKeeper.GetAccount(ctx, liquidityProvider)
	fmt.Println(" *** Getting account")

	account, ok := a.(types.LiquidityProviderAccount)
	if !ok {
		logger.Debug(fmt.Sprintf("Account is not a liquidity provider"), "address", liquidityProvider)
		fmt.Println(" *** Account is not a liquidity provider!")
		return
	}

	credit, ok := account.Credit.SafeSub(amount)
	if !ok {
		logger.Debug(fmt.Sprintf("Insufficient credit for minting operation"), "requested", amount, "available", account.Credit)
		fmt.Println(" *** Insufficient credit for minting", amount, account.Credit)
		return
	}

	account.Credit = credit

	balance := account.GetCoins()
	balance = balance.Add(amount)
	account.SetCoins(balance)

	// TODO Ought to trigger the supply module
	fmt.Println(" *** Storing account state in IAVL.")
	k.authKeeper.SetAccount(ctx, account)
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
