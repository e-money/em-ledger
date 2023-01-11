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
		{[]string{"E-EUR"}, false, 0},
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
