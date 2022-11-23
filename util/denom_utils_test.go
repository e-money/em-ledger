package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseDenominations(t *testing.T) {
	testdata := []struct {
		denoms []string
		valid  bool
		count  int
	}{
		{[]string{"eeur", "ejpy"}, true, 2},
		{[]string{"  eeur", "ejpy "}, true, 2},
		{[]string{"  eeur ", " ejpy "}, true, 2},
		{[]string{"  eeur,EEUR,Euro stablecoin ", " ejpy "}, true, 2},
		{[]string{""}, false, 0},
		// - allowed now
		// https://github.com/cosmos/cosmos-sdk/pull/9699/commits/9f5fe1a9c37d2b446b90dc97740e77b0d7855763
		{[]string{"E-EUR"}, true, 1},
	}

	for _, d := range testdata {
		denoms, error := ParseDenominations(d.denoms, "e-Money EUR stablecoin")
		if error != nil {
			if d.valid {
				assert.NoError(t, error)
			}
			continue
		}

		assert.Len(t, denoms, d.count)
	}
}
