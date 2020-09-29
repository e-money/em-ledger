import json

# Removes validator set and staking information from emoney2.genesis.json and replaces it with the validator set of testnet.genesis.json.

def mergeSupply(mainnet, testnet):
    total={}
    for balance in mainnet["supply"] + testnet["supply"]:
        amount, denom = balance["amount"], balance["denom"]

        if denom in total:
            newAmount = int(total[denom]) + int(amount)
            total[denom] = str(newAmount)
        else:
            total[denom] = amount

    supply = []

    for denom, balance in total.items():
        supply.append( { "denom" : denom, "amount" : amount })

    mainnet["supply"] = supply

with open('emoney2.export.json') as emoney2g, open('testnet.export.json') as testnetg:
    mainnet = json.load(emoney2g)
    testnet = json.load(testnetg)

    # Transplant testnet staking state to the new net.
    mainnet["app_state"]["auth"]["accounts"].extend(testnet["app_state"]["auth"]["accounts"])
    mainnet["app_state"]["staking"] = testnet["app_state"]["staking"]
    mainnet["app_state"]["distribution"] = testnet["app_state"]["distribution"]
    mainnet["validators"] = testnet["validators"]

    # "supply" module state
    mergeSupply(mainnet["app_state"]["supply"], testnet["app_state"]["supply"])

    with open('output.json', 'w', encoding='utf-8') as f:
        json.dump(mainnet, f, ensure_ascii=False, indent=4)
