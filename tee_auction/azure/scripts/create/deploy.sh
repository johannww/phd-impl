#!/bin/bash
az deployment group create --resource-group carbon --template-file ./azure/arm_template.json
