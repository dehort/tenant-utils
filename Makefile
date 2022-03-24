BINARY = org-id-column-populator


build: $(BINARY)
	go build ./...

lint:
	golangci-lint run ./...

test:
	go test -v ./...

$(BINARY): cmd/$(BINARY)/*.go pkg/tenantconv/*.go
	go build -o $(BINARY) cmd/org-id-column-populator/*.go

build_image:
	podman build . -t tenant-utils

clean:
	go clean
	rm -f $(BINARY)
