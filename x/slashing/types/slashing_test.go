package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPenaltiesEmpty(t *testing.T) {
	specs := map[string]struct {
		src *Penalties
		exp bool
	}{
		"with elements": {
			src: &Penalties{Elements: []Penalty{{
				Validator: "foobar",
				Amounts:   sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(1))),
			}}},
			exp: false,
		},
		"with empty elements": {
			src: &Penalties{Elements: []Penalty{}},
			exp: true,
		},
		"with elements not set": {
			src: &Penalties{},
			exp: true,
		},
		"nil obj": {
			src: &Penalties{},
			exp: true,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			got := spec.src.Empty()
			assert.Equal(t, spec.exp, got)
		})
	}
}

func TestPenaltiesAdd(t *testing.T) {
	specs := map[string]struct {
		src *Penalties
		add []Penalty
		exp []Penalty
	}{
		"merge different coins": {
			src: &Penalties{Elements: []Penalty{{
				Validator: "a",
				Amounts:   sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(1))),
			}}},
			add: []Penalty{{
				Validator: "a",
				Amounts:   sdk.NewCoins(sdk.NewCoin("bar", sdk.NewInt(2))),
			}},
			exp: []Penalty{{
				Validator: "a",
				Amounts:   sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(1)), sdk.NewCoin("bar", sdk.NewInt(2))),
			}},
		},
		"merge same coin denom": {
			src: &Penalties{Elements: []Penalty{{
				Validator: "a",
				Amounts:   sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(1))),
			}}},
			add: []Penalty{{
				Validator: "a",
				Amounts:   sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(2))),
			}},
			exp: []Penalty{{
				Validator: "a",
				Amounts:   sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(3))),
			}},
		},
		"different validator same coin denom": {
			src: &Penalties{Elements: []Penalty{{
				Validator: "a",
				Amounts:   sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(1))),
			}}},
			add: []Penalty{{
				Validator: "b",
				Amounts:   sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(2))),
			}},
			exp: []Penalty{{
				Validator: "a",
				Amounts:   sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(1))),
			}, {
				Validator: "b",
				Amounts:   sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(2))),
			}},
		},
		"with empty elements": {
			src: &Penalties{},
			add: []Penalty{{
				Validator: "a",
				Amounts:   sdk.NewCoins(sdk.NewCoin("bar", sdk.NewInt(2))),
			}},
			exp: []Penalty{{
				Validator: "a",
				Amounts:   sdk.NewCoins(sdk.NewCoin("bar", sdk.NewInt(2))),
			}},
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			src := spec.src
			for _, p := range spec.add {
				src.Add(p.Validator, p.Amounts)
			}
			assert.Equal(t, spec.exp, src.Elements)
		})
	}
}
