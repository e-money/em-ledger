package types

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
)

var _ auth.Account = LiquidityProviderAccount{}

type LiquidityProviderAccount struct {
	auth.Account

	Credit sdk.Coins
}

func NewLiquidityProviderAccount(baseAccount auth.Account, credit sdk.Coins) *LiquidityProviderAccount {
	return &LiquidityProviderAccount{
		Account: baseAccount,
		Credit:  credit,
	}
}

func (acc LiquidityProviderAccount) String() string {
	var pubkey string

	if acc.GetPubKey() != nil {
		pubkey = sdk.MustBech32ifyAccPub(acc.GetPubKey())
	}

	return fmt.Sprintf(`Account:
  Address:       %s
  Pubkey:        %s
  Credit:        %s
  Coins:         %s
  AccountNumber: %d
  Sequence:      %d`,
		acc.GetAddress(), pubkey, acc.Credit, acc.GetCoins(), acc.GetAccountNumber(), acc.GetSequence(),
	)
}
