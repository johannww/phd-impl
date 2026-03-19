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

.PHONY: experiments
experiments:
	cd ./experiments/deploy; \
	./scripts/deploy.sh; \
	./scripts/shutdown.sh
	
ai-list-structs:
	./scripts/gopls_get_structs.sh ./chaincodes/carbon/
	./scripts/gopls_get_structs.sh ./chaincodes/interop/
	./scripts/gopls_get_structs.sh ./tee_auction/go
