// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package bank

import sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

var ErrRestrictedDenomination = sdkerrors.Register("embank", 1, "Denomination is restricted")
