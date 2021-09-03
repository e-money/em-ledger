# e-Money Tokens

Before you continue reading, please review the [e-Money Whitepaper](https://e-money.com/documents/e-Money%20Whitepaper.pdf) to understand how our currency-backed tokens work.

Even though our currency-backed tokens have an initial 1.0 exchange rate to their underlying currency, the rate will change over time. The exchange rate change
when interest is accrued on the underlying currency or the number of currency-backed tokens are inflated.

As such, the current exchange rate must be used when displaying the value of a token amount in terms of the underlying currency (e.g. eeur in terms of EUR).

The list of supported tokens and their respective exchange rates are available at through the API.

## API

### Getting Supported Tokens

[https://api.e-money.com/v1/tokens.json](https://api.e-money.com/v1/tokens.json).

```json
{
  "_comment": "See https://github.com/e-money/em-ledger/blob/master/docs/tokens.md",
  "last_updated": "2020-01-01T00:00:00.000000Z",
  "tokens": [
    {
      "token": "eeur",
      "description": "Interest bearing EUR token",
      "exponent": 6,
      "underlying_currency": "EUR"
    },
    {
      "token": "echf",
      "description": "Interest bearing CHF token",
      "exponent": 6,
      "underlying_currency": "CHF"
    },
    {
      "token": "ejpy",
      "description": "Interest bearing JPY token",
      "exponent": 6,
      "underlying_currency": "JPY"
    }
  ],
  "currencies": [
    {
      "currency": "EUR",
      "exponent": 2
    },
    {
      "currency": "CHF",
      "exponent": 2
    },
    {
      "currency": "JPY",
      "exponent": 0
    }
  ]
}
```

The number of decimal places to display for a currency or token is available in `exponent` (see [ISO 4217](https://www.iso.org/iso-4217-currency-codes.html)).

### Getting Exchange Rates

The current exchange rate from a token to underlying currency can be fetched via [https://api.e-money.com/v1/rates](https://api.e-money.com/v1/rates):

```json
[
  {
      "source": "ECHF",
      "destination": "CHF",
      "rate": 0.987021
    },
    {
      "source": "ECHF",
      "destination": "DKK",
      "rate": 6.765752
    },
...
```

As the rates change slowly, it is typically not necessary to query more than once per hour.

## Displaying Amounts

To display a `eeur` amount in terms of `EUR`, the following calculation must be made:  
`EUR_amount = eeur_amount * exchange_rate / 10^eeur_exponent`

The `exponent` for EUR specifies the number of digits to display. For instance, `eeur 1234567890` at an exchange rate of 0.999950 should be displayed as `EUR 123,45` (after rounding down).
