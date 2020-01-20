package bank

import (
	"github.com/e-money/em-ledger/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type RestrictedKeeper interface {
	GetRestrictedDenoms(sdk.Context) types.RestrictedDenoms
}
