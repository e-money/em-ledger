// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var ErrAccountDoesNotExist = sdkerrors.Register(ModuleName, 1, "account does not exist")
