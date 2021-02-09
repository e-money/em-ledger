// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package types

import (
	"github.com/stretchr/testify/require"
	"testing"
)

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestSimple(t *testing.T) {
	var (
		addr1 = sdk.AccAddress([]byte("addr1"))
		addr2 = sdk.AccAddress([]byte("addr2"))
	)

	state := []RestrictedDenom{
		{"ngm", []string{addr1.String()}},
		{"evilt", []string{}},
		{"othergovtoken", []string{addr2.String()}},
	}

	rs := RestrictedDenoms(state)

	if ngm, found := rs.Find("ngm"); found {
		require.True(t, ngm.IsAnyAllowed(addr1, addr2))
		require.True(t, ngm.IsAnyAllowed(addr1))
		require.False(t, ngm.IsAnyAllowed(addr2))
		require.False(t, ngm.IsAnyAllowed())
	} else {
		require.Fail(t, "ngm token not found")
	}

	if evilt, found := rs.Find("evilt"); found {
		require.False(t, evilt.IsAnyAllowed(addr1, addr2))
		require.False(t, evilt.IsAnyAllowed())
	} else {
		require.Fail(t, "evilt token not found")
	}

	if othergovtoken, found := rs.Find("othergovtoken"); found {
		require.True(t, othergovtoken.IsAnyAllowed(addr1, addr2))
		require.True(t, othergovtoken.IsAnyAllowed(addr2))
		require.False(t, othergovtoken.IsAnyAllowed(addr1))
		require.False(t, othergovtoken.IsAnyAllowed())
	} else {
		require.Fail(t, "othergovtoken token not found")
	}
}
