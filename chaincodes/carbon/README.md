# TODOs

- GetStateByRangeWithPagination only works for read only transactions.
    - GetStateByRange works for read/write, but it is capped by "totalQueryLimit" in core.yaml.
    - Thus, I will ask for the keys in range.
    - If the number of keys is equal to the limit, I will ask perform a search using the last key.
    - How do I retrieve totalQueryLimit?

# Testing

Setup the carbon chaincode with the fabric test network:
```bash
./tests/scripts/setup-test-network.sh
# test the invoke:
./tests/scripts/test-invoke.sh
# cleanup (delete binaries, config and builders folders):
./tests/scripts/cleanup.sh

```
