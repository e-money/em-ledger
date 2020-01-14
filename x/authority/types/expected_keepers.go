package types

type GasPricesKeeper interface {
	SetMinimumGasPrices(gasPricesStr string) error
}
