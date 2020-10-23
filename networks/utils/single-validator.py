import json
import sys

# Injects the specified validator into a genesis file and allocates nearly all voting power to that validator.

# Gensis file to modify
inputfile = "genesis-modified.json"

# The validator that replaces the entire set of the inputfile
validator = "emoneyvaloper1fj76mrfpcwm7yqlmfqkqlx9kgvjns97cnffp3u"
consensus_pubkey = "emoneyvalconspub1zcjduepq4jkprfqcfx3txy34lqp3wsanseyzcxtxvh4lhrgmnwqlshfr75rqmdp64x"

removeUnusedValidators = True

with open(inputfile) as f:
    genesis = json.load(f)

    del genesis["validators"]

    genesis["genesis_time"] = "2020-07-25T12:00:00Z"

    app_state = genesis["app_state"]

    staking = app_state["staking"]
    distr = app_state["distribution"]

    # Replacing the first existing validator
    replaced_val = staking["validators"][0]["operator_address"]

    # ----- Replace references to a validator with the new one.
    for v in staking["validators"]:
        if v["operator_address"] == replaced_val:
            v["operator_address"] = validator
            v["consensus_pubkey"] = consensus_pubkey


    for outstanding_reward in distr["outstanding_rewards"]:
        if outstanding_reward["validator_address"] == replaced_val:
            outstanding_reward["validator_address"] = validator

    for outstanding_commission in distr["validator_accumulated_commissions"]:
        if outstanding_commission["validator_address"] == replaced_val:
            outstanding_commission["validator_address"] = validator

    for current_reward in distr["validator_current_rewards"]:
        if current_reward["validator_address"] == replaced_val:
            current_reward["validator_address"] = validator

    for historical_reward in distr["validator_historical_rewards"]:
        if historical_reward["validator_address"] == replaced_val:
            historical_reward["validator_address"] = validator

    # ----- Reset all delegations, rewards etc.


    # Change all delegator_starting_infos to the validator
    for delegation_start in distr["delegator_starting_infos"]:
        delegation_start["validator_address"] = validator


    for delegation in staking["delegations"]:
        delegation["validator_address"] = validator

    for last_power in staking["last_validator_powers"]:
        if last_power["Address"] == replaced_val:
            last_power["Address"] = validator

    for last_power in staking["last_validator_powers"]:
        if last_power["Address"] != validator:
            last_power["Power"] = "1"
        else:
            last_power["Power"] = staking["last_total_power"]

    if staking["redelegations"] != None:
        for redelegation in staking["redelegations"]:
            redelegation["validator_dst_address"] = validator

    totalTokens = 0
    delegator_shares = 0
    for v in staking["validators"]:
        totalTokens += int(v["tokens"])
        delegator_shares += float(v["delegator_shares"])

        v["tokens"] = "1"
        v["delegator_shares"] = "1"


    for v in staking["validators"]:
        if v["operator_address"] == validator:
            v["tokens"] = str(totalTokens)
            v["delegator_shares"] = str(delegator_shares)

    # Strip the remaining validators entirely.
    if removeUnusedValidators:
        staking["validators"] = [v for v in staking["validators"] if v["operator_address"] == validator]
        staking["last_validator_powers"] = [v for v in staking["last_validator_powers"] if v["Address"] == validator]

        distr["outstanding_rewards"] = [v for v in distr["outstanding_rewards"] if v["validator_address"] == validator]
        distr["validator_accumulated_commissions"] = [v for v in distr["validator_accumulated_commissions"] if v["validator_address"] == validator]
        distr["validator_current_rewards"] = [v for v in distr["validator_current_rewards"] if v["validator_address"] == validator]
        distr["validator_historical_rewards"] = [v for v in distr["validator_historical_rewards"] if v["validator_address"] == validator]



    print(json.dumps(genesis))
