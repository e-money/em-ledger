// This software is Copyright (c) 2019 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package types

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDenoms(t *testing.T) {
	assert.True(t, ValidateDenom("x2eur"))
	assert.False(t, ValidateDenom("X2EUR"))
	assert.False(t, ValidateDenom("123456"))
}

func TestParseDenominations(t *testing.T) {
	var testdata = []struct {
		denoms string
		valid  bool
		count  int
	}{
		{"x2eur,x0jpy", true, 2},
		{"  x2eur, x0jpy ", true, 2},
		{" x2eur,,x0jpy ", true, 2},
		{"", true, 0},
		{"  x2EUR, x0jpy ", false, -1},
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
