.PHONY: dev build test lint clean docker

# Go
build:
	@cd server && go build -o ../hf-local .

server:
	@if [ -d "../server" ]; then \
		cd server && go build -o ../hf-local; \
	else \
		go build -o hf-local; \
	fi

server-test:
	@cd server && go test -v ./...

server-lint:
	@cd server && go vet ./...

# Python
python-install:
	@cd python && pip install -e .

python-test:
	@cd python && pytest

python-lint:
	@cd python && ruff check src/ tests/
	@cd python && mypy src/

# Development
dev: server
	@./hf-local.exe

test: server-test python-test

lint: server-lint python-lint

clean:
	@rm -f hf-local.exe hf-local
	@cd python && rm -rf build/ dist/ *.egg-info .pytest_cache/ .mypy_cache/ .ruff_cache/

# Docker
docker-build:
	@docker build -t hf-local-hub:latest .

docker-run:
	@docker-compose up -d

docker-down:
	@docker-compose down
