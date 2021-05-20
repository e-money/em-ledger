// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.
package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDenoms(t *testing.T) {
	assert.True(t, ValidateDenom("eeur"))
	// todo (reviewer): the regexp pattern for valid denoms was changed for IBC to `[a-zA-Z][a-zA-Z0-9/]{2,127}`
	// assert.False(t, ValidateDenom("EEUR"))
	assert.False(t, ValidateDenom("123456"))
}

func TestParseDenominations(t *testing.T) {
	testdata := []struct {
		denoms string
		valid  bool
		count  int
	}{
		{"eeur,ejpy", true, 2},
		{"  eeur, ejpy ", true, 2},
		{" eeur,,ejpy ", true, 2},
		{"", true, 0},
		{"  E-EUR, ejpy ", false, -1},
	}

	for _, d := range testdata {
		denoms, error := ParseDenominations(d.denoms)
		if error != nil {
			if d.valid {
				assert.NoError(t, error)
			}
			continue
		}

		assert.Len(t, denoms, d.count)
	}
}
