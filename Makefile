VERSION = $(shell git describe --tags --always --dirty)
LDFLAGS=-ldflags "-X main.version=$(VERSION)"
OSARCH=$(shell go env GOHOSTOS)-$(shell go env GOHOSTARCH)

MDMUTIL=\
	mdmutil-darwin-amd64 \
	mdmutil-darwin-arm64 \
	mdmutil-linux-amd64 \
	mdmutil-linux-arm64 \
	mdmutil-windows-amd64.exe

my: mdmutil-$(OSARCH)

$(MDMUTIL): cmd/mdmutil
	GOOS=$(word 2,$(subst -, ,$@)) GOARCH=$(word 3,$(subst -, ,$(subst .exe,,$@))) go build $(LDFLAGS) -o $@ ./$<

%-$(VERSION).zip: %.exe
	rm -f $@
	zip $@ $<

%-$(VERSION).zip: %
	rm -f $@
	zip $@ $<

clean:
	rm -rf mdmutil-*

release: $(foreach bin,$(MDMUTIL),$(subst .exe,,$(bin))-$(VERSION).zip)

test:
	go test -v -cover -race ./...

.PHONY: my $(MDMUTIL) clean release test
