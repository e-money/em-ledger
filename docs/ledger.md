# Ledger Device Support

Using a hardware wallet to store your keys greatly improves the security of your crypto assets. The Ledger Nano devices acts as an enclave of the seed and private keys, and the process of signing transaction takes place within it. No private information ever leaves the Ledger device.

At the core of a Ledger device there is a [24 word seed phrase](https://www.ledger.com/academy/crypto/what-is-a-recovery-phrase) that is used to generate private keys. This seed phrase is generated when you initialize you Ledger and can be used to create accounts on the e-Money network.

**Do not lose or share your seed phrase with anyone. To prevent theft or loss of funds, it is best to keep multiple copies of your seed phrase stored in safe, secure places. If someone is able to gain access to your seed phrase, they will fully control the accounts associated with them.**

The following is a short tutorial on using Ledger Nano with the em-ledger command line interface (CLI). If using a CLI tool is unfamiliar to you, please use the [Lunie](https://lunie.io) web wallet instead.

## Before You Begin

The tool used to generate addresses and transactions on the e-Money network is `emcli`. Here is how to get started:  

- Initialize your Ledger Nano device and securely store the seed phrase. Do not share it with anyone!
- Install the [Ledger Live](https://www.ledger.com/ledger-live/) application to manage the Ledger Nano.
- Use Ledger Live to install the "Cosmos" application onto your Ledger Nano.
- Install `emcli` using the [build instructions](https://github.com/e-money/em-ledger#build-instructions).
- Verify that emcli is installed correctly with the following command:

```bash
emcli version --long

name: e-money
server_name: emd
client_name: emcli
version: 0.7.0-rc4-5-g2f3ae1b
commit: 2f3ae1b5f676c2db366d3cbcd1f90f1327074e35
build_tags: netgo,ledger
go: go version go1.13.8 darwin/amd64
```

## Add Your Ledger Key

- Connect and unlock your Ledger device.
- Open the Cosmos app on your Ledger.
- Create an account in emcli from your ledger key.

Be sure to change the _key_name_ parameter to be a meaningful name. The `ledger` flag tells `emcli` to use your Ledger to seed the account.

```bash
emcli keys add <key_name> --ledger

NAME: TYPE: ADDRESS:     PUBKEY:
<key_name> ledger emoney1... emoneypub1...
```

e-Money uses HD Wallets. This means you can setup many accounts using the same Ledger seed. To create another account from your Ledger device, run:

```bash
emcli keys add <secondKeyName> --ledger
```

## Confirm Your Address on the Ledger Device

Run this command to display your address on the device. Use the `key_name` you gave your ledger key.

```bash
emcli keys show <key_name> -d
```

Confirm that the address displayed on the device matches that displayed when you added the key.


## Connect to a Full Node

Next, you need to configure emcli with the URL of a e-Money full node and the appropriate `chain-id`. In this example we connect to the public load balanced full node operated by validator.network on the `emoney-1` chain. You can point your `emcli` to any full node you like, as long as the `chain-id` is set to the same as the full node.

```bash
emcli config node https://emoney.validator.network:443
emcli config chain-id emoney-1
```

Test your connection with a query such as:

``` bash
emcli query staking validators
```

## Sign a Transaction

You are now ready to start signing and sending transactions. Send a transaction with emcli using the `tx send` command.

Be sure to unlock your device with the PIN and open the Cosmos app before trying to run these commands.

``` bash
emcli tx send --help # to see all available options.
```

Use the `key_name` you set for your Ledger key and gaia will connect with the Cosmos Ledger app to then sign your transaction.

```bash
emcli tx send <key_name> <destination_address> <amount><denomination>
```

When prompted with `confirm transaction before signing`, Answer `Y`.

Next you will be prompted to review and approve the transaction on your Ledger device. Be sure to inspect the transaction information displayed on the screen. You can scroll through each field and each message.

## Receive Funds

To receive funds to the e-Money account on your Ledger device, retrieve the address for your Ledger account (the ones with `TYPE ledger`) with this command:

```bash
emcli keys list

NAME: TYPE: ADDRESS:     PUBKEY:
<key_name> ledger emoney1... emoneypub1...
```

### Further Documentation

Not sure what `emcli` can do? See [emcli.md](emcli.md) for further information.
