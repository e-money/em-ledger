package types

import "github.com/cosmos/cosmos-sdk/types"

func (a *Penalties) Empty() bool {
	return a == nil || len(a.Elements) == 0
}

func (a *Penalties) Add(validator string, amounts types.Coins) {
	// todo (reviewer) : merge with duplicates
	for i := range a.Elements {
		if a.Elements[i].Validator == validator {
			a.Elements[i].Amounts = a.Elements[i].Amounts.Add(amounts...)
			return
		}
	}
	a.Elements = append(a.Elements, Penalty{
		Validator: validator,
		Amounts:   amounts,
	})
}
