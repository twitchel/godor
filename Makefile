.PHONY: help

# COLORS
GREEN  := $(shell tput -Txterm setaf 2)
YELLOW := $(shell tput -Txterm setaf 3)
WHITE  := $(shell tput -Txterm setaf 7)
RESET  := $(shell tput -Txterm sgr0)
PI_ADDRESS := "192.168.1.1"

TARGET_MAX_CHAR_NUM=20

## Show help
help:
	@echo ''
	@echo 'Usage:'
	@echo '  ${YELLOW}make${RESET} ${GREEN}<target>${RESET}'
	@echo ''
	@echo 'Targets:'
	@awk '/^[a-zA-Z\-\_0-9]+:/ { \
		helpMessage = match(lastLine, /^## (.*)/); \
		if (helpMessage) { \
			helpCommand = substr($$1, 0, index($$1, ":")); \
			helpMessage = substr(lastLine, RSTART + 3, RLENGTH); \
			printf "  ${YELLOW}%-$(TARGET_MAX_CHAR_NUM)s${RESET} ${GREEN}%s${RESET}\n", helpCommand, helpMessage; \
		} \
	} \
	{ lastLine = $$0 }' $(MAKEFILE_LIST)

## Build a binary for the current local environment
build-local:
	go build -o godor .

## Build a binary for the ARM environment
build-arm:
	env GOOS=linux GOARCH=arm GOARM=5 go build -o dist/godor .

## Deploy a new version of the binary to the chosen ARM environment "make deploy-arm PI_ADDRESS=192.168.1.1"
deploy-arm: build-arm
	scp dist/godor pi@${PI_ADDRESS}:/home/pi
	rm dist/godor


