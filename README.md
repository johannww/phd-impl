# A country-agnostic Blockchain ETS Model with Geographical and Time References inspired by the Brazilian Ecosystem using Hyperledger Fabric, Hyperledger Cacti and Microsoft Confidential Containers

This repository is an ongoing implementation of a Blockchain-based Emission Trading System (ETS) model that incorporates **geographical** and **temporal** references. To ensure security and privacy during auctions, we utilize **Microsoft Confidential Containers** for Trusted Execution Environments (TEE).

# Techonologies used

- [Hyperledger Fabric](https://github.com/hyperledger/fabric) v2.5.12
    - For the main blockchain network
- [Hyperledger Cacti](https://github.com/hyperledger-cacti/cacti)
    - For interoperability between different blockchains
    - Supports Ethereum-based chains, Hyperledger Fabric, and Corda
- [Microsoft Confidential Containers](https://github.com/microsoft/confidential-sidecar-containers/tree/4814b442cf71de2b1317f00846f16727e40a3088) (for TEE)
    - Auction HTTPS service runs inside a confidential container attested by the hardware
    - For secure auctioning of carbon credits
    - Ensures data privacy and integrity during auctions
    - Uses AMD SEV-SNP technology

# Model overview

<img width="1168" height="673" alt="image" src="https://github.com/user-attachments/assets/fe673c8d-8380-4187-96a6-5384d9b58ee1" />

# Auction types

## Independent auction

Auction module with independent policies only. In this auction type, the credits are fungible, and the multiplier weight is applied before and after the auction.

<img width="556" height="479" alt="image" src="https://github.com/user-attachments/assets/e4016650-280f-4550-b4d2-d670f2dc83bc" />

## Coupled auction

Auction module with coupled policies. In this auction type, the same carbon sink-credit \(w\) results in different burn-credit values (\(x\), \(y\), \(z\)) when sold to different burners. Because this auction might require the geolocation of the burners to apply the policies, it has access to the relevant private data that is not available to all participants in the network.

<img width="556" height="479" alt="image" src="https://github.com/user-attachments/assets/2792bc56-54f2-4f66-a169-81524e910c39" />

# Description

This is the implementation of our PhD thesis. It consists of a Blokchain Emission Trading Systems (BETS) coupled with a cross-chain framework to 

## Use case

Imagine a scenario where small farmers can be rewarded 

# Architecture

## Core functionalities

### Spatial and time data

Spatial data and time data are important components for weighing carbon offset. Roughly, it does not make sense for a company in Brasil to buy carbon credits from a REDD iniciative in Japan. The distance is too long and the carbon will unlikely be settled by that project.

Also, for a sustainable environment, the carbon sinking rate should be higher or equal to the emission rate at the present moment. To ensure that, our model also considers the time window between the emission and the sinking. The credits for sinking are minted periodically to provide such a control.

### Offset multiplier based on policies

Space-time metrics between emitter and sinker are transformed into a multiplier. This multiplier affects the sinking power of the carbon credits. In our market, carbon credits are non-fungible as they value are not fixed. Generally, credits recently minted and near the emission source will be more valuable to a emitter company. On the other hand, credits minted way back in the past from a distant location will not have as much value.

<!-- TODO: find articles, perhaps talk to environment engineer to subsidise the multiplier calculation --> 

### Policy modularity

<!-- TODO: Ensure policy plugability --> 
Our model provides plugability for other policy types

### Auction


## Stakeholders

Here are the stakholders considered. We aimed to conform to the Brazilian bill 2148/2015:

<!-- TODO: continue here-->
- Project developers
- Project methodologies
...

## Interaction with external databases

## Interoperability

# TODOs

<!-- TODO: project todos --> 

- [X] Understand and fix hyperledger bevel
- [ ] Finish chaincode
- [] Integrate cacti for interoperability
- [X] Enable TEE auction (microsoft confidential containers)
- [ ] Understand hyperledger caliper for experiments
- [ ] Understand grafana, prometheus for metrics
