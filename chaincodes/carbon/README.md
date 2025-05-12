# TODOs

- [X] GetStateByRangeWithPagination only works for read only transactions.
- use the contract api for faster development

# Testing

Setup the carbon chaincode with the fabric test network:
```bash
./tests/scripts/setup-test-network.sh
# test the invoke:
./tests/scripts/test-invoke.sh
# cleanup (delete binaries, config and builders folders):
./tests/scripts/cleanup.sh

```
