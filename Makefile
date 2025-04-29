diagrams:
	cd ./chaincodes/carbon/ && ./docs/scripts/generate-diagrams.sh

test:
	cd ./chaincodes/carbon/ && go test -count=1 ./tests
	
