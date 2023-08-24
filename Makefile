.PHONY: build

SAMLOCAL := $(shell command -v samlocal 2> /dev/null)

SQLBOILER := $(shell command -v sqlboiler 2> /dev/null)

require-cli:
ifndef SAMLOCAL
	$(error "samlocal is not available please install with `pip install aws-sam-cli-local`")
endif

require-sqlboiler:
ifndef SQLBOILER
	$(error "sqlboiler is not available please install before proceeding`")
endif

generate_models: require-sqlboiler
	cd signer && sqlboiler psql -c sqlboiler_wallets.toml && cd -
	cd signer && sqlboiler psql -c sqlboiler_transactions.toml && cd -

build:
	sam build

integration_test:
	cd signer && go test ./... -v -tags integration && cd -

run-local: require-cli
	samlocal build && sam local invoke SignTransactionFunction --event events/event.json

vet:
	cd signer && go vet ./...  && cd -

unit_test:
	cd signer && go test ./... -v && cd -

tidy_dependencies:
	cd signer && go mod tidy && cd -
