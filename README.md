# A country-agnostic Blockchain ETS Model with Geographical and Time References inspired by the Brazilian Ecosystem using Hyperledger Fabric, Hyperledger Cacti and Microsoft Confidential Containers

# TODOs

<!-- TODO: project todos --> 

- [X] Understand and fix hyperledger bevel
- [ ] Finish chaincode
- [] Integrate cacti for interoperability
- [X] Enable TEE auction (microsoft confidential containers)
- [ ] Understand cacti and how to  uset it 
- [ ] Understand hyperledger caliper for experiments
- [ ] Understand grafana, prometheus for metrics

# phd-impl

# Description

This is the implementation of our PhD thesis. It consists of a Blokchain Emission Trading Systems (BETS) coupled with a cross-chain framework to 

## Technology Stack

Primary chains:
- Hyperledger Fabric/Besu
- Hyperledger Cacti 

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

<!-- TODO: how can the auction be maximized taking the many combinations? Should we implement it?-->


## Stakeholders

Here are the stakholders considered. We aimed to conform to the Brazilian bill 2148/2015:

<!-- TODO: continue here-->
- Project developers
- Project methodologies
...

## Interaction with external databases

## Interoperability

<!-- TODO: talk about the Interoperability measures --> 

