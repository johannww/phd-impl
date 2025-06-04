#!/bin/bash
REGISTRY=carbonchain.azurecr.io

cd ./tee_auction
docker build -t $REGISTRY/carbon_auction:latest .
docker push $REGISTRY/carbon_auction:latest
