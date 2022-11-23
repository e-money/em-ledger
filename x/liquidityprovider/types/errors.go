package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var ErrAccountDoesNotExist = sdkerrors.Register(ModuleName, 1, "account does not exist")
