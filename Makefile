RESOURCE_GROUP ?= carbon
CLUSTER_NAME  ?= carbon-aks

diagrams:
	$(MAKE) -C ./chaincodes/carbon docs
	$(MAKE) -C tee_auction docs

test-network:
	$(MAKE) -C ./chaincodes/carbon test-network

BLUE := \033[0;34m
NC := \033[0m

test:
	@printf "$(BLUE)Running unit tests for carbon chaincode$(NC)\n"
	$(MAKE) -C chaincodes/carbon test-no-cache unit-test
	@printf "$(BLUE)Running unit tests for interop chaincode$(NC)\n"
	$(MAKE) -C chaincodes/interop test-no-cache unit-test
	@printf "$(BLUE)Running unit tests for tee_auction$(NC)\n"
	$(MAKE) -C tee_auction unit-test
	@printf "$(BLUE)Running unit tests for data_api$(NC)\n"
	$(MAKE) -C data_api test

docker:
	$(MAKE) -C ./chaincodes/carbon cc-docker
	$(MAKE) -C ./chaincodes/interop cc-docker
	$(MAKE) -C tee_auction docker
	$(MAKE) -C ./data_api docker
	$(MAKE) -C ./experiments/deploy/images/fabric-tools docker

docker-push:
	$(MAKE) -C ./chaincodes/carbon docker-push
	$(MAKE) -C ./chaincodes/interop docker-push
	$(MAKE) -C tee_auction docker-push
	$(MAKE) -C ./data_api docker-push
	$(MAKE) -C ./experiments/deploy/images/fabric-tools docker-push

.PHONY: experiments
experiments:
	cd ./experiments/deploy; \
	./scripts/deploy.sh

experiments-clean:
	cd ./experiments/deploy; \
	./scripts/shutdown.sh

experiments-aks:
	@printf "$(BLUE)Pushing images to ghcr.io/johannww/phd-impl...$(NC)\n"
	#$(MAKE) docker-push
	@printf "$(BLUE)Deploying to AKS...$(NC)\n"
	cd ./experiments/deploy/scripts; \
	./deploy-aks.sh --skip-push
	
aks-provision:
	./experiments/deploy/azure/provision_aks.sh

aks-clean:
	./experiments/deploy/scripts/shutdown-aks.sh

aks-stop:
	./experiments/deploy/azure/shutdown_aks.sh

aks-start:
	az aks start --resource-group $(RESOURCE_GROUP) --name $(CLUSTER_NAME)

aks-down:
	@echo "Deleting Resource Group: $(RESOURCE_GROUP)..."
	az group delete --name $(RESOURCE_GROUP) --yes --no-wait

ai-list-structs:
	./scripts/gopls_get_structs.sh ./chaincodes/carbon/
	./scripts/gopls_get_structs.sh ./chaincodes/interop/
	./scripts/gopls_get_structs.sh ./tee_auction/go
