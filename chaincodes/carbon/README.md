# TODOs

- [X] GetStateByRangeWithPagination only works for read only transactions.
- [X] use the contract api for faster development
- [X] create a config for the test-network including idemix orgs
- [] GATEWAY: Make the Idemix Identity create the pseudonym key and make the signer fetch it from the Idemix Identity
    - This is necessary because the public identity is serialized before the signature. Therefore, we must
        create the pseudonym before the signature function is invoked.
    - [] There is an error with the constructed ZKP for the credential:
        ```bash
        docker logs peer0.org1.example.com -f:
        # signature invalid: zero-knowledge proof is invalid

        ```
    - Should I CalculateProof() for each idemix nym?


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
