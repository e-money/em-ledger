// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package types

import (
	"fmt"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	auth "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/gogo/protobuf/proto"
)

var _ auth.AccountI = &LiquidityProviderAccount{}

func NewLiquidityProviderAccount(account auth.AccountI, mintable sdk.Coins) (*LiquidityProviderAccount, error) {
	msg, ok := account.(proto.Message)
	if !ok {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidType, "can't proto marshal %T", msg)
	}
	any, err := codectypes.NewAnyWithValue(msg)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrPackAny, err.Error())
	}
	return &LiquidityProviderAccount{
		Account:  any,
		Mintable: mintable,
	}, nil
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
	nestedAccount := acc.GetNestedAccount()
	if nestedAccount.GetPubKey() != nil {
		pubkey = sdk.MustBech32ifyPubKey(sdk.Bech32PubKeyTypeAccPub, nestedAccount.GetPubKey())
	}

	return fmt.Sprintf(`Account:
  Address:       %s
  Pubkey:        %s
  Mintable:      %s
  AccountNumber: %d
  Sequence:      %d`,
		nestedAccount.GetAddress(), pubkey, acc.Mintable, nestedAccount.GetAccountNumber(), nestedAccount.GetSequence(),
	)
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (m LiquidityProviderAccount) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	var account auth.AccountI
	if err := unpacker.UnpackAny(m.Account, &account); err != nil {
		return err
	}
	return codectypes.UnpackInterfaces(account, unpacker)
}

func (m *LiquidityProviderAccount) GetNestedAccount() auth.AccountI {
	content, ok := m.Account.GetCachedValue().(auth.AccountI)
	if !ok {
		panic("nested account was nil")
	}
	return content
}

func (acc *LiquidityProviderAccount) GetAddress() sdk.AccAddress {
	return acc.GetNestedAccount().GetAddress()
}

func (acc *LiquidityProviderAccount) SetAddress(address sdk.AccAddress) error {
	return acc.GetNestedAccount().SetAddress(address)

}

func (acc *LiquidityProviderAccount) GetPubKey() cryptotypes.PubKey {
	return acc.GetNestedAccount().GetPubKey()
}

func (acc *LiquidityProviderAccount) SetPubKey(key cryptotypes.PubKey) error {
	nestedAccount := acc.GetNestedAccount()
	err := nestedAccount.SetPubKey(key)
	if err != nil {
		return err
	}
	any, err := codectypes.NewAnyWithValue(nestedAccount)
	if err != nil {
		return err
	}
	acc.Account = any
	return nil
}

func (acc *LiquidityProviderAccount) GetAccountNumber() uint64 {
	return acc.GetNestedAccount().GetAccountNumber()
}

func (acc *LiquidityProviderAccount) SetAccountNumber(u uint64) error {
	return acc.GetNestedAccount().SetAccountNumber(u)
}

func (acc *LiquidityProviderAccount) GetSequence() uint64 {
	return acc.GetNestedAccount().GetSequence()
}

func (acc *LiquidityProviderAccount) SetSequence(u uint64) error {
	return acc.GetNestedAccount().SetSequence(u)
}
