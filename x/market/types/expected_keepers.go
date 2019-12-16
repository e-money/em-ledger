package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/exported"
)

type AccountKeeper interface {
	GetAccount(sdk.Context, sdk.AccAddress) exported.Account
	AddAccountListener(func(sdk.Context, exported.Account))
	//SetAccount(ctx sdk.Context, acc exported.Account)
	//AddAccountListener(Listener)
}

//
//type Listener interface {
//	AccountChanged(ctx sdk.Context, acc exported.Account)
//}
