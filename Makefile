.PHONY: build test dev lint dashboard qa clean

VERSION ?= 1.0.0
BINARY=bin/agentops
DASHBOARD=dashboard
LDFLAGS=-ldflags "-X github.com/gog1withme/AgentOps/cli/version.Version=$(VERSION)"

build:
	go build $(LDFLAGS) -o $(BINARY) ./cli

test:
	go test -p 1 ./cli/... ./schema/... ./scripts/...

lint:
	go vet ./...

dev: build
	./$(BINARY) dev

dashboard-install:
	cd $(DASHBOARD) && npm install

dashboard-build:
	cd $(DASHBOARD) && npm run build

dashboard-dev:
	cd $(DASHBOARD) && npm run dev

qa:
	go run scripts/generate-qa.go

clean:
	rm -rf bin $(DASHBOARD)/.next $(DASHBOARD)/out

all: build dashboard-build
