// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	auth "github.com/cosmos/cosmos-sdk/x/auth/exported"
)

var _ auth.Account = LiquidityProviderAccount{}

type LiquidityProviderAccount struct {
	auth.Account

	Mintable sdk.Coins `json:"mintable" yaml:"mintable"`
}

func NewLiquidityProviderAccount(baseAccount auth.Account, mintable sdk.Coins) *LiquidityProviderAccount {
	return &LiquidityProviderAccount{
		Account:  baseAccount,
		Mintable: mintable,
	}
}

func (acc *LiquidityProviderAccount) IncreaseMintableAmount(increase sdk.Coins) {
	acc.Mintable = acc.Mintable.Add(increase...)
}

// Function panics if resulting mintable amount is negative. Should be checked prior to invocation for cleaner handling.
func (acc *LiquidityProviderAccount) DecreaseMintableAmount(decrease sdk.Coins) {
	if mintable, anyNegative := acc.Mintable.SafeSub(decrease); !anyNegative {
		acc.Mintable = mintable
		return
	}

	panic(fmt.Errorf("mintable amount cannot be negative"))
}

func (acc LiquidityProviderAccount) String() string {
	var pubkey string

	if acc.GetPubKey() != nil {
		pubkey = sdk.MustBech32ifyPubKey(sdk.Bech32PubKeyTypeAccPub, acc.GetPubKey())
	}

	return fmt.Sprintf(`Account:
  Address:       %s
  Pubkey:        %s
  Mintable:      %s
  Coins:         %s
  AccountNumber: %d
  Sequence:      %d`,
		acc.GetAddress(), pubkey, acc.Mintable, acc.GetCoins(), acc.GetAccountNumber(), acc.GetSequence(),
	)
}
