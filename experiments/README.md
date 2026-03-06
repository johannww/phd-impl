# Experiments idea

- Measure the confidential container performance metrics (CPU, memory, network).
- Metrics over the expected lifecycle of the application:
    - Fix orderer, peers, fabric parameters, etc.
    - Deploy the application and perform expected transactions on our carbon credit use case.
    - I could design a single application for this tests, to avoid spawning multiple peer invokes.
        - use the fabric-gateway to perform transactions
        - measure transactions
        - beware of the waiting or not for the transaction to be commited in a block, as it can affect the performance metrics.
    - Assign all attributes found in `./chaincodes/carbon/identities/consts.go` to the admin user, so that we can perform all transactions without worrying about access control.
