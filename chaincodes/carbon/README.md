# TODOs

- [X] GetStateByRangeWithPagination only works for read only transactions.
- [X] use the contract api for faster development
- [X] create a config for the test-network including idemix orgs


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
