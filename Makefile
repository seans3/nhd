# Makefile for the NHD Service Project

.DEFAULT_GOAL := help

# Go variables
GOCMD=go
GOBUILD=$(GOCMD) build
GORUN=$(GOCMD) run
GOTEST=$(GOCMD) test
GOMOD=$(GOCMD) mod
BINARY_NAME=nhd-backend

# Python variables
PYTHON=python3
VENV_DIR=reporter/venv

# Proto variables
PROTOC=protoc
PROTOC_PY=$(VENV_DIR)/bin/python -m grpc_tools.protoc

.PHONY: all backend-build backend-run backend-test frontend-install frontend-start frontend-build proto reporter-install-deps clean help

all: backend-build frontend-build ## Build all application components

# ====================================================================================
# HELPERS
# ====================================================================================

help: ## Display this help screen
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n\nTargets:\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%%-24s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

# ====================================================================================
# PROTO BUFFERS
# ====================================================================================

proto: reporter-install-deps ## Generate Go and Python code from proto definitions
	@echo "--- Generating Go protobuf code..."
	@cd backend && $(PROTOC) --go_out=. --go_opt=paths=source_relative proto/nhd.proto
	@echo "--- Generating Python protobuf code..."
	@cd backend && $(PROTOC_PY) -I=proto --python_out=../reporter/proto/gen/python --grpc_python_out=../reporter/proto/gen/python proto/nhd.proto
	@echo "Protobuf code generated."

# ====================================================================================
# TESTING
# ====================================================================================

test: backend-test ## Run all unit tests for the Go backend

# ====================================================================================
# GO BACKEND
# ====================================================================================

backend-build: ## Build the Go backend binary
	@echo "--- Building Go backend..."
	@cd backend && $(GOBUILD) -o $(BINARY_NAME) .
	@echo "Go backend built."

backend-run: ## Run the Go backend server (requires GOOGLE_CLOUD_PROJECT env var)
ifndef GOOGLE_CLOUD_PROJECT
	$(error GOOGLE_CLOUD_PROJECT is not set. Please set it, e.g., export GOOGLE_CLOUD_PROJECT=<your-gcp-project-id>)
endif
	@echo "--- Running Go backend..."
	@cd backend && $(GORUN) main.go

backend-test: ## Run all unit tests for the Go backend
	@echo "--- Testing Go backend..."
	@cd backend && $(GOTEST) -v ./...

# ====================================================================================
# REACT FRONTEND
# ====================================================================================

frontend-install: ## Install frontend npm dependencies
	@echo "--- Installing frontend dependencies..."
	@cd frontend && npm install

frontend-start: frontend-install ## Start the frontend development server
	@echo "--- Starting frontend dev server..."
	@cd frontend && npm start

frontend-build: frontend-install ## Build the frontend for production
	@echo "--- Building frontend..."
	@cd frontend && npm run build
	@echo "Frontend built."

frontend-test: frontend-install ## Run unit tests for the frontend
	@echo "--- Testing frontend..."
	@cd frontend && npm test -- --watchAll=false




# ====================================================================================
# PYTHON REPORTER
# ====================================================================================

reporter-install-deps: ## Create a venv and install Python dependencies for the reporter
	@echo "--- Installing Python reporter dependencies..."
	@if [ ! -d "$(VENV_DIR)" ]; then \
		$(PYTHON) -m venv $(VENV_DIR); \
	fi
	@$(VENV_DIR)/bin/pip install -r reporter/requirements.txt

# ====================================================================================
# CLEANUP
# ====================================================================================

clean: ## Remove all generated files and build artifacts
	@echo "--- Cleaning up..."
	@rm -f backend/$(BINARY_NAME)
	@rm -rf backend/proto/gen
	@rm -rf reporter/proto/gen
	@rm -rf frontend/node_modules
	@rm -rf frontend/build
	@rm -rf $(VENV_DIR)
	@echo "Cleanup complete."
