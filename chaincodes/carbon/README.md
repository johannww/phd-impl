# TODOs

- [X] GetStateByRangeWithPagination only works for read only transactions.
- [X] use the contract api for faster development
- [X] create a config for the test-network including idemix orgs
- [X] GATEWAY: Make the Idemix Identity create the pseudonym key and make the signer fetch it from the Idemix Identity
- [X] verify that the confidential containers is running the expected ccePolicy
- [] Handle the private multiplier: we have a private part of MatchedBid
    - How can the parties involved in the get the price?
        - They trust the settlement service to transfer the tokens.
        - The settlement service will debit a value <= than the buy price and >= than the sell price.
- [] Handle fungibility of tokens for independent auctions
    - They are minted as non-fungible, and become fungible after the auction.


# Testing

Setup the carbon chaincode with the fabric test network:
```bash
# ./tests/scripts/setup-test-network.sh prereq # to download the fabric binaries
./tests/scripts/setup-test-network.sh
# add the idemix org
./tests/scripts/add-idemix-org.sh

# test the invoke:
./tests/scripts/test-invoke.sh
# cleanup (delete binaries, config and builders folders):
./tests/scripts/cleanup.sh

```
