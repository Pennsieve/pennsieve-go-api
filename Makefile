.PHONY: help clean local-services test test-ci docker-clean package publish

LAMBDA_BUCKET ?= "pennsieve-cc-lambda-functions-use1"
SERVICE_NAME  ?= "pennsieve-go-api"
WORKING_DIR   ?= "$(shell pwd)"
PACKAGE_NAME  ?= "api-v2-authorizer-${IMAGE_TAG}.zip"

.DEFAULT: help

help:
	@echo "Make Help for $(SERVICE_NAME)"
	@echo ""
	@echo "make clean   	- removes dynamodb data directory"
	@echo "make test    	- run tests locally using docker containers"
	@echo "make test-ci 	- used by Jenkins to run tests without exposing ports"
	@echo "start-dynamodb 	- Start local DynamoDB container for testing"
	@echo "make package 	- create venv and package lambda functions"
	@echo "make publish 	- package and publish lambda function"

local-services:
	docker compose -f docker-compose.test.yml down --remove-orphans
	docker compose -f docker-compose.test.yml -f docker-compose.local.override.yml up -d dynamodb pennsievedb

test: local-services
	cd $(WORKING_DIR)/lambda/authorizer && go test -v ./...

test-ci:
	docker compose -f docker-compose.test.yml down --remove-orphans
	docker compose -f docker-compose.test.yml up --build --abort-on-container-exit --exit-code-from test

# Spin down active docker containers.
docker-clean:
	docker compose -f docker-compose.test.yml down

# Remove dynamodb database
clean: docker-clean
	rm -rf $(WORKING_DIR)/lambda/bin

tidy:
	cd $(WORKING_DIR)/lambda/authorizer && go mod tidy

package:
	@echo ""
	@echo "**********************************"
	@echo "*   Building Authorizer lambda   *"
	@echo "**********************************"
	@echo ""
	cd $(WORKING_DIR)/lambda/authorizer; \
  		env GOOS=linux GOARCH=amd64 go build -o $(WORKING_DIR)/lambda/bin/authorizer/authorizer_lambda; \
		cd $(WORKING_DIR)/lambda/bin/authorizer/ ; \
			zip -r $(WORKING_DIR)/lambda/bin/authorizer/$(PACKAGE_NAME) .

publish:
	@make package
	@echo ""
	@echo "************************************"
	@echo "*   Publishing Authorizer lambda   *"
	@echo "************************************"
	@echo ""
	aws s3 cp $(WORKING_DIR)/lambda/bin/authorizer/$(PACKAGE_NAME) s3://$(LAMBDA_BUCKET)/pennsieve-go-api/
	rm -rf $(WORKING_DIR)/lambda/bin/authorizer/$(PACKAGE_NAME)