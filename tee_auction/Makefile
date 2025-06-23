# The order of commands follows the description in the README.md

REGISTRY_NAME=carbonchain
REGISTRY=carbonchain.azurecr.io

registry:
	az group create --name carbon --location eastus
	az acr create --resource-group carbon --name $(REGISTRY_NAME) --sku Basic --admin-enabled true
	az acr login --name $(REGISTRY_NAME)

docker:
	docker build -t $(REGISTRY)/carbon_auction:latest .
	docker push $(REGISTRY)/carbon_auction:latest

policy:
	REGISTRY_USER=`az acr credential show --name $(REGISTRY_NAME) --query "username"`; \
	REGISTRY_PASS=`az acr credential show --name $(REGISTRY_NAME) --query "passwords[0].value"`; \
	sed -i "s/REGISTRY_USER/$$REGISTRY_USER/" ./azure/arm_template.json; \
	sed -i "s/REGISTRY_PASS/$$REGISTRY_PASS/" ./azure/arm_template.json; \
	az confcom acipolicygen -a ./azure/arm_template.json

deploy:
	az deployment group create --resource-group carbon --template-file ./azure/arm_template.json

cleanup:
	az container delete --resource-group carbon --name carbon-auction-container --yes

verify-policy-image-layers:
	cd azure ; \
	if [ ! -d integrity-vhd ]; then \
		git clone https://github.com/microsoft/integrity-vhd/; \
	fi; \
	cd integrity-vhd; \
	touch .nosync; \
	echo ""; \
	echo "Local image FS layers:"; \
	go run cmd/dmverity-vhd/main.go -d roothash -i $(REGISTRY)/carbon_auction:latest; \
	cd ../..; \
	echo "Policy FS layers:"; \
	az confcom acipolicygen -a ./azure/arm_template.json --outraw-pretty-print 2>/dev/null | grep -A 3 "layers"
