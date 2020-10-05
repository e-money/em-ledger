import json
import sys

# Injects the specified validator into a genesis file and allocates nearly all voting power to that validator.

# Gensis file to modify
inputfile = "genesis-modified.json"

# The validator that replaces the entire set of the inputfile
validator = "emoneyvaloper1x54upxhjrlqxfujmp9p27ezr2gufs34478glx5"
consensus_pubkey = "emoneyvalconspub1zcjduepqw089yduq3ccrevvxlsncwgus8xckc4h8pswnz088aszackq2w5dqv4qg7r"

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


    # TODO Adding up delegator_shares in this manner probably doesn't make sense.
    for v in staking["validators"]:
        if v["operator_address"] == validator:
            v["tokens"] = str(totalTokens)
            v["delegator_shares"] = str(delegator_shares)

    print(json.dumps(genesis))


    # print(json.dumps(staking["validators"], indent=4))


    # print(json.dumps(genesis, indent=4))





#  [X] Ret i [distribution][delegator_starting_infos]
#  [X] Ret til samme validator i alle [staking][delegations]
#  [X] Sæt "Power" til 0 for alle andre end "hovedpersonen" i [staking][last_validator_powers]
#  [X] Sæt "Power" til [staking][last_total_power] for "hovedpersonen" i [staking][last_validator_powers]
#  [X] Sæt [staking][redelegations][validator_dst_address] til "hovedpersonen".
#  [X] Sæt alle [staking][validators][power] til 0
#  [X] Sæt "hovedpersonens" power til summen af de powers der er fjernet
#  [ ] Slet "validators" i roden af genesis
