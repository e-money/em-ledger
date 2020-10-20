# Working with multisig accounts

```
# Generate a multisig key
emcli keys add myMultiSig --multisig <key1>,<key2>,<key3> --multisig-threshold 2

# Generate the unsigned transaction json.
emcli tx send emoney10lu2cwzlt02k5qey45dk5lkr8kc265wa659sqv emoney1s73cel9vxllx700eaeuqr70663w5f0twzcks3l 500ungm --generate-only > send_tx.json

# Each member of the multisig must then create a signature file.
emcli tx sign send_tx.json --multisig <multi-sig-bech32-address> --from <key-name> > <key-name-sig>.json

# Combine the signatures 
emcli tx multisign send_tx.json.json <multi-sig-key-name> <signature-file1>.json <signature-file2>.json ... <signature-fileN>.json  > final_tx.json

# Broadcast the transaction to your preferred full-node.
emcli tx broadcast final_tx.json
```