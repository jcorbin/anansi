PACKAGES=github.com/jcorbin/anansi/...

.PHONY: test
test: lint
	go test -cover -coverprofile=.test_coverage $(PACKAGES)

.test_coverage: test

.PHONY: view-cover
view-cover: .test_coverage
	go tool cover -html=$<

.PHONY: lint
lint:
	./bin/go_list_sources.sh $(PACKAGES) | xargs gofmt -e -d
	golint $(PACKAGES)
	go vet $(PACKAGES)

.PHONY: fmt
fmt:
	./bin/go_list_sources.sh $(PACKAGES) | xargs gofmt -w
