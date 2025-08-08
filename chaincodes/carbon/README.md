# TODOs

- [X] GetStateByRangeWithPagination only works for read only transactions.
- [X] use the contract api for faster development
- [X] create a config for the test-network including idemix orgs
- [X] GATEWAY: Make the Idemix Identity create the pseudonym key and make the signer fetch it from the Idemix Identity
- [] verify that the confidential containers is running the expected ccePolicy
  - file tee/azure.go
  - This video explains: https://www.youtube.com/watch?v=H9DP5CMqGac
  - Test the coded solution
- [] Handle the private multiplier
    - How can the TEE export it in a way that it is not exposed to every on on-chain?
    - I may export the private prices as a separate map that can be referenced by some ID.
        - Thus, the TEE invoker can publish the multiplier as transient data.


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
