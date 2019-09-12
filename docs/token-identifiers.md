## Token Identifiers

The purpose of this spec is to enable user interfaces to present amounts in a user-friendly manner, e.g. with the proper number of decimals.
This should be possible by relying solely on the token identifier, so that tokens can be issued and carried over IBC without relying on dictionary of token descriptions.

Referencing the [ISO 4217](https://www.iso.org/iso-4217-currency-codes.html) standard for currencies, our token identifiers contain the "Exponent" inserted into the second character of the "Alphabetic Identier": 

Identifier: &lt;first-alpha&gt;&lt;exponent&gt;&lt;subsequent-alphas&gt;

Notice that the exponent can consist of multiple digits (> 9), so a parser must consume digits until it reaches the subsequent-alpha.

The number of minor units is calculated as 10^exponent. Amounts are thus stored and carried around as number of minor units in a integer type.

### Rationale
The exponent was placed at the second position as this is trivial to parse. In addition all known currency identifiers are 3 letters so it is safe to place it at the second position.

Placing the exponent at the beginning or the end of the identifier was ruled out as unsafe. It would be confusing if it was displayed verbatim next to an amount ("EUR2 1000" or "1000 2EUR").

### Examples
The following examples related to tokens issued by e-Money, where our currency-backed tokens are prefixed with "E" for e-Money.


* E2EUR: Euro backed token "EEUR" with exponent 2 and 10^2 minor units (cents).
* E2GBP: Pound sterling backed token "EGBP" with exponent 2 and 10^2 minor units (penny).
* E2USD: US Dollar backed token "EUSD" with exponent 2 and 10^2 minor units (cents).
* E0JPY: Japanese Yen backed token "EJPY" with exponent 0 and 10^0 minor unit.
* N3GM: Staking token "NGM" for the e-Money zone with exponent 3 and 10^3 minor units.

Existing tokens with longer identifiers could also be identified like this:
* A8TOM: [Cosmos Hub](https://cosmos.network) ATOM staking tokens with exponent 8 and 10^8 minor units (uatom).
* I18RIS: [IRISNet](https://irisnet.org) IRIS staking token with exponent 18 and 10^18 minor units (iris-atto).
