.PHONY: build run clean

OS := $(shell uname -s)
APP_EXECUTABLE = gator
OUT_DIR = .out
OS ?= $(shell uname -s)
GOARCH ?= amd64

ifeq ($(OS),Darwin)
    GOOS := darwin
    EXECUTABLE := $(OUT_DIR)/$(APP_EXECUTABLE)
else ifeq ($(OS),Linux)
    GOOS := linux
    EXECUTABLE := $(OUT_DIR)/$(APP_EXECUTABLE)
else
    GOOS := windows
    EXECUTABLE := $(OUT_DIR)/$(APP_EXECUTABLE).exe
endif

$(shell if [ ! -d "$(OUT_DIR)" ]; then mkdir -p $(OUT_DIR); fi)

build:
	@echo "Building for $(OS) ($(GOOS)/$(GOARCH))"
	@mkdir -p $(OUT_DIR)
	GOARCH=$(GOARCH) GOOS=$(GOOS) go build  -ldflags="-s -w" -o $(EXECUTABLE) ./cmd/gator

build-docker:
	@echo "Building docker image"
	docker build  -ldflags="-s -w" -t gator .
	
clean:
	@echo "Cleaning up"
	go clean
	rm -rf $(OUT_DIR)/$(APP_EXECUTABLE)*

clean-docker:
	docker rmi gator