// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.
package util

import (
	"fmt"
	"testing"

	"github.com/e-money/em-ledger/x/authority/types"

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

func TestParseDenominations2(t *testing.T) {
	testdata := []struct {
		denoms []types.Denomination
		valid  bool
		count  int
	}{
		{
			denoms: []types.Denomination{
				{Base: "eeur"},
				{Base: "ejpy"},
			},
			valid: true,
			count: 2,
		},
	}

	for _, d := range testdata {
		sd := fmt.Sprintf("%s", d.denoms)
		fmt.Println(sd)
	}
}
