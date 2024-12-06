.PHONY: help clean test test-ci start-dynamodb docker-clean package publish

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

test:
	docker compose -f docker-compose.test.yml down --remove-orphans
	docker compose -f docker-compose.test.yml up --exit-code-from local_tests local_tests

test-ci:
	mkdir -p test-dynamodb-data
	chmod -R 777 test-dynamodb-data
	docker compose -f docker-compose.test.yml down --remove-orphans
	docker compose -f docker-compose.test.yml up --exit-code-from ci_tests ci_tests

# Start a clean DynamoDB container for local testing
start-dynamodb: docker-clean
	docker compose -f docker-compose.test.yml up dynamodb


# Spin down active docker containers.
docker-clean:
	docker compose -f docker-compose.test.yml down

# Remove dynamodb database
clean: docker-clean
	rm -rf test-dynamodb-data

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