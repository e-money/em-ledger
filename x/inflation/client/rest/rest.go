package rest

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/gorilla/mux"
)

// RegisterRoutes registers minting module REST handlers on the provided router.
func RegisterRoutes(cliCtx client.Context, r *mux.Router) {
	r.HandleFunc(
		"/inflation/current", queryInflationHandlerFn(cliCtx),
	).Methods("GET")
}
