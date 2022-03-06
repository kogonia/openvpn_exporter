.PHONY: build
build:
	go build -o openvpn-exporter -v .

.PHONY: run
run:
	go run -v .

.PHONY: test
test:
	go test -v -race -timeout 30s ./...

.PHONY: clean
clean:
	rm -f openvpn-exporter

.PHONY: install
install:
	go build -o /usr/bin/openvpn-exporter -v .

.PHONY: uninstall
uninstall:
	rm -f /usr/bin/openvpn-exporter

.DEFAULT_GOAL := build
