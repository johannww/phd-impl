diagrams:
	$(MAKE) -C ./chaincodes/carbon docs
	$(MAKE) -C tee_auction docs

test:
	$(MAKE) -C chaincodes/carbon test-no-cache test-network

docker:
	$(MAKE) -C ./chaincodes/carbon cc-docker
	$(MAKE) -C ./chaincodes/interop cc-docker
	$(MAKE) -C tee_auction docker
	
ai-list-structs:
	./scripts/gopls_get_structs.sh ./chaincodes/carbon/
	./scripts/gopls_get_structs.sh ./chaincodes/interop/
	./scripts/gopls_get_structs.sh ./tee_auction/go
