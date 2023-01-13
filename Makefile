.PHONY: help clean test

LAMBDA_BUCKET ?= "pennsieve-cc-lambda-functions-use1"
SERVICE_NAME  ?= "pennsieve-go-api"
WORKING_DIR   ?= "$(shell pwd)"
PACKAGE_NAME  ?= "api-v2-authorizer-${VERSION}.zip"

.DEFAULT: help

help:
	@echo "Make Help for $(SERVICE_NAME)"
	@echo ""
	@echo "make clean   - removes node_modules directory"
	@echo "make test    - run tests"
	@echo "make package - create venv and package lambda functions"
	@echo "make publish - package and publish lambda function"

test:
	@echo ""
	@echo "********************"
	@echo "*   Testing API    *"
	@echo "********************"
	@echo ""
	@cd $(WORKING_DIR)/pkg; \
		go test ./... ;
	@echo ""
	@echo "***************************"
	@echo "*   Testing Authorizer    *"
	@echo "***************************"
	@echo ""
	@cd $(WORKING_DIR)/lambda/authorizer; \
		go test ./... ;

package:
	@echo ""
	@echo "**********************************"
	@echo "*   Building Authorizer lambda   *"
	@echo "**********************************"
	@echo ""
	cd $(WORKING_DIR)/lambda/authorizer; \
  		env GOOS=linux GOARCH=amd64 go build -o $(WORKING_DIR)/lambda/bin/authorizer/$(SERVICE_NAME)-$(VERSION); \
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