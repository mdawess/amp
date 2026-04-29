BINARY := amp
INSTALL_DIR := $(shell go env GOPATH)/bin

.PHONY: build install uninstall update clean

build:
	go build -o $(BINARY) .

install:
	go install .

update:
	git pull
	$(MAKE) install

uninstall:
	rm -f $(INSTALL_DIR)/$(BINARY)

clean:
	rm -f $(BINARY)
