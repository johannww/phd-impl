diagrams:
	$(MAKE) -C ./chaincodes/carbon docs
	$(MAKE) -C tee_auction docs

test:
	cd ./chaincodes/carbon/ && go test -count=1 ./tests
	
