SERVER_PORT := $(shell ./autotests/random unused-port 2>/dev/null)
TEMP_FILE := $(shell ./autotests/random tempfile 2>/dev/null)


all: vet test test_iter

run:
	go run cmd/shortener/*.go -g on

build:
	go build -o cmd/shortener/shortener cmd/shortener/*.go

update:
	git fetch template && git checkout template/v2 .github


test:
	go test -v ./...

vet:
	go vet -vettool=./.tools/statictest ./...

test_iter:
	$(eval BRANCH_NAME := $(shell git branch --show-current 2>/dev/null))
	$(eval ITER_NUM := $(shell echo "$(BRANCH_NAME)" | grep -oE '[0-9]+'))
	make test_iter$(ITER_NUM)

test_iter1: build
	./autotests/shortenertest -test.v -test.run=^TestIteration1$$ -binary-path=cmd/shortener/shortener

test_iter2: test_iter1
	./autotests/shortenertest -test.v -test.run=^TestIteration2$$  -source-path=.

test_iter3: test_iter2
	./autotests/shortenertest -test.v -test.run=^TestIteration3$$  -source-path=.

test_iter4: test_iter3
	./autotests/shortenertest -test.v -test.run=^TestIteration4$$ \
		-binary-path=cmd/shortener/shortener \
	 	-server-port=$(SERVER_PORT)

test_iter5: test_iter4
	./autotests/shortenertest -test.v -test.run=^TestIteration5$$ \
		-binary-path=cmd/shortener/shortener \
	 	-server-port=$(SERVER_PORT)

test_iter6: test_iter5
	./autotests/shortenertest -test.v -test.run=^TestIteration6$$ \
			-source-path=.

test_iter7: test_iter6
	./autotests/shortenertest -test.v -test.run=^TestIteration7$$ \
			-binary-path=cmd/shortener/shortener \
			-source-path=.

test_iter8: test_iter7
	./autotests/shortenertest -test.v -test.run=^TestIteration8$$ \
			-binary-path=cmd/shortener/shortener

test_iter9: test_iter8
	./autotests/shortenertest -test.v -test.run=^TestIteration9$$ \
			-binary-path=cmd/shortener/shortener \
			-source-path=. \
			-file-storage-path=$(TEMP_FILE)
