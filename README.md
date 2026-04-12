# A country-agnostic Blockchain ETS Model with Geographical and Time References inspired by the Brazilian Ecosystem using Hyperledger Fabric, Hyperledger Cacti and Microsoft Confidential Containers

This repository implements a country-agnostic Blockchain-based Emission Trading System (BETS) model that incorporates **geographical** and **temporal** references, inspired by the Brazilian ecosystem. The system ensures security and privacy during auctions using **Microsoft Confidential Containers** for Trusted Execution Environments (TEE).

## Technologies Used

- **[Hyperledger Fabric](https://github.com/hyperledger/fabric) v3.1.4**
  - Core blockchain network implementation
  - Private data collections for sensitive auction data
  - Chaincode-as-a-Service (CCAAS) deployment model
  
- **[Hyperledger Cacti](https://github.com/hyperledger-cacti/cacti)**
  - Cross-chain interoperability framework
  - Hash Time-Locked Contracts (HTLC) for atomic swaps
  - Supports Ethereum-based chains, Hyperledger Fabric, and Corda
  
- **[Microsoft Confidential Containers](https://github.com/microsoft/confidential-sidecar-containers)**
  - TEE-based auction service with hardware attestation
  - Secure auctioning of carbon credits with data privacy guarantees
  - Uses AMD SEV-SNP technology for confidential computing
  - Azure Kubernetes Service (AKS) integration for cloud deployment

- **Go 1.26.1**
  - Primary programming language for all components
  - Go workspaces for multi-module project structure

## Project Structure

```
.
├── chaincodes/
│   ├── carbon/          # Main carbon credit chaincode
│   ├── interop/         # Interoperability chaincode (HTLC)
│   └── common/          # Shared utilities and state management
├── tee_auction/         # TEE auction service
├── data_api/            # External data API service (SICAR mock)
├── experiments/
│   ├── deploy/          # Kubernetes deployment (Helm charts, CCAAS)
│   └── exp-app/         # Performance benchmarking tool
└── scripts/             # Build and deployment scripts
```

## Core Features

### 1. Spatial and Temporal Data Integration

Carbon offset effectiveness depends heavily on geographical and temporal factors:

- **Geographical Proximity**: Credits from nearby carbon sinks are more valuable to emitters
- **Temporal Alignment**: Carbon sinking rate should match or exceed emission rate at the present moment
- **Periodic Minting**: Credits are minted at regular intervals to maintain temporal control

### 2. Dynamic Offset Multiplier Based on Policies

Space-time metrics between emitter and sinker are transformed into multipliers that affect carbon credit value:

- **Non-Fungible Credits**: Credit value varies based on distance, time, and other factors
- **Policy Modularity**: Pluggable policy system for custom multiplier calculations
- **Coupled vs Independent Auctions**: Support for both auction types (see below)

### 3. Auction Mechanisms

#### Independent Auction
- Credits are fungible within the auction
- Multiplier weights applied uniformly before and after auction
- Traditional double auction mechanism

#### Coupled Auction
- Same carbon sink credit results in different burn-credit values for different buyers
- Considers buyer geolocation and other private data
- Implements privacy-preserving policies using private data collections
- TEE-based computation ensures fairness and privacy

### 4. UTXO-Based Payment Model

To eliminate MVCC (Multi-Version Concurrency Control) conflicts during high-concurrency auction settlement, the system implements a **UTXO (Unspent Transaction Output) model** for virtual token payments:

- **Parallel Settlement**: Each seller/buyer receives unique UTXO keys instead of updating centralized wallets
- **Aggregation Function**: Users can aggregate multiple UTXOs into their wallet via separate transactions
- **Backward Compatibility**: Traditional `VirtualTokenWallet` retained for bid publishing
- **One UTXO per Owner per Auction**: Aggregated payments/refunds to minimize state bloat

### 5. Private Data Collections

Sensitive auction data is protected using Hyperledger Fabric's private data collections:

- **Bid Privacy**: Buy/sell bid prices and quantities kept private
- **Auction Results**: Clearing prices and matched bids stored in private collections
- **Organization-Specific Access**: Only authorized organizations can access private data

### 6. TEE-Based Auction Settlement

Auction clearing price computation happens inside a Trusted Execution Environment:

- **Hardware Attestation**: AMD SEV-SNP hardware reports verify execution integrity
- **Cryptographic Signatures**: Auction results signed by TEE private key
- **Verifiable Results**: On-chain verification of TEE attestation and signatures

### 7. Cross-Chain Interoperability

Integration with Hyperledger Cacti enables:

- **HTLC-Based Swaps**: Atomic token swaps across chains
- **Credit Locking**: Lock credits on one chain while unlocking on another
- **Multi-Chain Settlement**: Support for Ethereum and other blockchain networks

## Stakeholders

The system supports the following stakeholders (conforming to Brazilian bill 2148/2015):

- Project Developers
- Project Methodologies
- Governments
- Farmers
- REDD Project Developers
- Emitting Companies
- Project Certifiers/Auditors
- Technical Committee
- Data Providers
- Normative Bodies
- Third-Party Brokers/Representatives
- Settlement Entities
- Cross-Chain Relayers
- General Public

## Development

### Prerequisites

- Go 1.26.1 or later
- Docker and Docker Compose
- Minikube (for local Kubernetes testing)
- Helm 3
- Hyperledger Fabric binaries v3.1.4
- Azure CLI (for AKS deployment)

### Build Commands

#### Quick Start

```bash
# Build all chaincodes
make test                          # Run all unit tests

# Carbon chaincode specific
cd chaincodes/carbon
make cc                            # Build chaincode
make unit-test                     # Run unit tests
make test                          # Run integration tests
make fmt                           # Format code

# TEE auction service
cd tee_auction
make unit-test                     # Run tests
make docker                        # Build Docker image

# Data API
cd data_api
make build                         # Build server
make test                          # Run tests
```

### Running Tests

```bash
# Run a single test
cd chaincodes/carbon
go test --tags=testing ./tests -run TestFunctionName

# Run without cache
make test-no-cache

# Skip Azure CEE tests
make test-no-amd-network-request
```

## Deployment

### Local Deployment (Minikube)

The project includes automated deployment scripts for Kubernetes experiments:

```bash
cd experiments/deploy/scripts
./deploy.sh                        # Full deployment
./shutdown.sh                      # Teardown
```

Deployment includes:
1. Minikube cluster setup
2. Fabric binaries installation
3. Custom Docker images build and load
4. Helm chart installation (orderers, peers, CAs)
5. Channel creation and joining
6. Chaincode deployment (CCAAS model)
7. Data integration services

### Cloud Deployment (Azure AKS)

Deploy to Azure Kubernetes Service with confidential computing:

```bash
# Provision AKS cluster with SEV-SNP nodes
make aks-provision LOCATION=centralindia RESOURCE_GROUP=carbon

# Pause cluster (stops billing)
make aks-stop

# Resume cluster
make aks-start

# Teardown
make aks-down RESOURCE_GROUP=carbon
```

## Performance Experiments

The `experiments/exp-app/` module provides high-concurrency benchmarking:

### Current Capabilities
- **Automated Setup**: Registers trusted providers, companies, wallets, and properties
- **Parallel Workloads**: Concurrent transaction execution using `SubmitAsync`
- **Continuous Minting**: Property-wide credit minting with wallet transfers
- **Metrics Collection**: Real-time latency, throughput (TPS), and success rate tracking

### Roadmap
- Continuous bidding simulation
- Periodic TEE auction settlement benchmarks
- Interoperability performance testing
- Dynamic topology testing

## Model Overview

<img width="1168" height="673" alt="image" src="https://github.com/user-attachments/assets/fe673c8d-8380-4187-96a6-5384d9b58ee1" />

### Independent Auction Flow

<img width="556" height="479" alt="image" src="https://github.com/user-attachments/assets/e4016650-280f-4550-b4d2-d670f2dc83bc" />

### Coupled Auction Flow

<img width="556" height="479" alt="image" src="https://github.com/user-attachments/assets/2792bc56-54f2-4f66-a169-81524e910c39" />

## Code Style

The project follows Go best practices:

- **Imports**: Grouped by standard library, external, and internal packages
- **Formatting**: `gofmt` with tabs for indentation
- **Naming**: PascalCase for exports, camelCase for private
- **Error Handling**: Always check errors, use `fmt.Errorf` with `%v` or `%w`
- **Testing**: Table-driven tests with `testify/require`
- **Documentation**: Godoc comments for all exported types and functions

## Project Status

### Completed
- [x] Hyperledger Fabric v3.1.4 network setup
- [x] Carbon credit chaincode with auction mechanisms
- [x] TEE auction service with Microsoft Confidential Containers
- [x] UTXO-based payment model (eliminates MVCC conflicts)
- [x] Private data collections for bid privacy
- [x] Kubernetes deployment with Helm charts
- [x] CCAAS (Chaincode-as-a-Service) deployment
- [x] Azure AKS confidential computing integration
- [x] Data API integration (SICAR mock service)
- [x] Comprehensive test suite (unit + integration)

### In Progress
- [ ] Hyperledger Cacti integration for cross-chain interoperability
- [ ] Performance benchmarking with Hyperledger Caliper
- [ ] Grafana + Prometheus metrics dashboard
- [ ] Production-grade deployment scripts

### Planned
- [ ] Additional policy modules (wind direction, vegetation indices)

## Contributing

This is a PhD research project. For collaboration inquiries, please open an issue.

## License

See LICENSE file for details.

## References

- Hyperledger Fabric Documentation: https://hyperledger-fabric.readthedocs.io/
- Go Code Review Comments: https://go.dev/wiki/CodeReviewComments
- Microsoft Confidential Containers: https://github.com/microsoft/confidential-sidecar-containers
- Brazilian Bill 2148/2015: Carbon Credit Framework

## Contact

For questions or collaboration opportunities, please open an issue in this repository.
