# TODOs

- [X] GetStateByRangeWithPagination only works for read only transactions.
- [X] use the contract api for faster development
- [] create a config for the test-network including idemix orgs


# Testing

Setup the carbon chaincode with the fabric test network:
```bash
./tests/scripts/setup-test-network.sh
# test the invoke:
./tests/scripts/test-invoke.sh
# cleanup (delete binaries, config and builders folders):
./tests/scripts/cleanup.sh

```
