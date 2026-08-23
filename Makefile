run:
	go run cmd/shortener/*.go -g on

build:
	go build -o cmd/shortener/shortener cmd/shortener/*.go

update:
	git fetch template && git checkout template/v2 .github


test:
	go test -v ./...

test_iter: build
	@chmod +x ./autotests/shortenertest-darwin-arm64
	$(eval BRANCH_NAME := $(shell git branch --show-current 2>/dev/null))
	$(eval ITER_NUM := $(shell echo "$(BRANCH_NAME)" | grep -oE '[0-9]+'))
	./autotests/shortenertest-darwin-arm64 -test.v -test.run=^TestIteration$(ITER_NUM)$$ -binary-path=cmd/shortener/shortener

