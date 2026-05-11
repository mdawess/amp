BINARY := amp
INSTALL_DIR := /usr/local/bin

.PHONY: build install uninstall update clean

build:
	go build -o $(BINARY) ./cmd/amp

install:
	go build -o $(INSTALL_DIR)/$(BINARY) ./cmd/amp

update:
	git pull
	$(MAKE) install

uninstall:
	rm -f $(INSTALL_DIR)/$(BINARY)

clean:
	rm -f $(BINARY)
