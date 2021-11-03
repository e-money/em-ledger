# em-ledger v1.0.0 Release Notes

v1.0.0 is a major release, which brings IBC support to the e-money chain.

## Major new features

- Upgrade to Cosmos SDK Stargate (v0.42.x)
- Support for  IBC connections
- The Cosmos SDK upgrade module is now available for future upgrades


## Minor changes
- Fees paid in stablecoin tokens are now sent to the NGM buyback module, rather than the distribution module. NGM fees are still sent to the distribution module.
- NGM buyback now creates market orders, rather than use last traded price.
- The *restricted denominations* concept and functionality has been removed.

# Migration changes
The `emd migrate` command makes the following changes to the genesis file it outputs:
- Size of validator set is increased to 100.
- Jailing for downtime is increased to 24 hours.
- NGM buyback is set to happen every 24 hours.


