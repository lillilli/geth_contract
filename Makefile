SERVICE_NAME 	:= api
DEPLOY_NAME 	:= deploy
CONTRACT_NAME := Counter
CONTRACT_PATH := ./contract
CONTRACT_FILE := contract.sol

.PHONY: setup
setup: ## Installing all service dependencies.
	echo "Setup..."
	GO111MODULE=off go get github.com/ethereum/go-ethereum
	GO111MODULE=on go mod vendor

	# go mod vendor remove all c files from vendor tree, so we copy it from GOPATH directory
	cp -r \
  "${GOPATH}/src/github.com/ethereum/go-ethereum/crypto/secp256k1/libsecp256k1" \
  "vendor/github.com/ethereum/go-ethereum/crypto/secp256k1/"

.PHONY: config
config: ## Creating the local config yml.
	echo "Creating local config yml ..."
	cp config.example.yml local.yml

.PHONY: gen_contract
gen_contract: ## Generate contract .go, .abi, .bin files.
	cd $(CONTRACT_PATH) && solc --abi $(CONTRACT_FILE) --overwrite -o build
	abigen --abi=$(CONTRACT_PATH)/build/$(CONTRACT_NAME).abi --pkg=store --out=$(CONTRACT_PATH)/$(CONTRACT_NAME).go
	cd $(CONTRACT_PATH) && solc --bin $(CONTRACT_FILE) --overwrite -o build
	cd $(CONTRACT_PATH) && abigen --bin=./build/$(CONTRACT_NAME).bin --abi=./build/$(CONTRACT_NAME).abi --pkg=store --out=./$(CONTRACT_NAME).go

.PHONY: build
build: ## Build the executable file of service.
	echo "Building..."
	cd cmd/$(SERVICE_NAME) && go build

.PHONY: run
run: build ## Run service with local config.
	echo "Running..."
	cd cmd/$(SERVICE_NAME) && ./$(SERVICE_NAME) -config=../../local.yml

.PHONY: lint
lint: ## Run lint for all packages.
	echo "Linting..."
	GO111MODULE=off go get -u github.com/golangci/golangci-lint/cmd/golangci-lint

	golangci-lint run --enable-all --disable gochecknoglobals \
	--disable lll -e "weak cryptographic primitive" -e InsecureSkipVerify \
	--disable dupl --print-issued-lines=false

.PHONY: run\:image
run\:image: ## Build service docker image and run it.
	echo "Running docker image..."
	docker build -t eth_example .
	docker stop eth_example_inst || true && docker rm -f eth_example_inst || true
	docker run -p 8080:8080 -v ${PWD}/local.yml:/root/local.yml --name eth_example_inst eth_example

.PHONY: deploy
deploy: ## Deploy docker image to docker hub.
	echo "Deploying ..."
	docker build -t eth_example .
	docker tag eth_example lillilli/geth-contract-example:latest
	docker push lillilli/geth-contract-example:latest

.PHONY: help
help: ## Display this help screen
	@grep -E '^[a-zA-Z_\-\:]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ": .*?## "}; {gsub(/[\\]*/,""); printf "\033[36m%-30s\033[0m%s\n", $$1, $$2}'
