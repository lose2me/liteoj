# LiteOJ Makefile (Linux / WSL)

FRONTEND_DIR=frontend
BACKEND_DIR=backend
DIST_SRC=$(FRONTEND_DIR)/dist
DIST_DST=dist
BIN=liteoj

.PHONY: all dev backend frontend install build clean run

all: build

install:
	cd $(FRONTEND_DIR) && pnpm install
	cd $(BACKEND_DIR) && go mod tidy

dev-backend:
	cd $(BACKEND_DIR) && go run ./cmd/liteoj

dev-frontend:
	cd $(FRONTEND_DIR) && pnpm dev

frontend:
	cd $(FRONTEND_DIR) && pnpm build
	rm -rf $(DIST_DST)
	mkdir -p $(DIST_DST)
	cp -r $(DIST_SRC)/. $(DIST_DST)/

build: frontend
	cd $(BACKEND_DIR) && CGO_ENABLED=0 go build -ldflags="-s -w" -o ../$(BIN) ./cmd/liteoj
	@echo "Built ./$(BIN)"

run: build
	./$(BIN)

clean:
	rm -rf $(DIST_DST) $(FRONTEND_DIR)/dist $(BIN)
