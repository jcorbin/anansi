PACKAGES=github.com/jcorbin/anansi/...

.PHONY: test
test: lint
	go test $(PACKAGES)

.PHONY: lint
lint:
	./bin/go_list_sources.sh $(PACKAGES) | xargs gofmt -e -d
	golint $(PACKAGES)
	go vet $(PACKAGES)

.PHONY: fmt
fmt:
	./bin/go_list_sources.sh $(PACKAGES) | xargs gofmt -w
