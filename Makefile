setup:
	go get -u github.com/cosmtrek/libgo/...
	go get -u golang.org/x/tools/cmd/goimports
	go get -u github.com/golang/lint/golint
	go get -u github.com/Sirupsen/logrus
	go get -u github.com/huichen/sego
	go get -u github.com/adamzy/cedar-go
	go get -u github.com/pkg/errors
	go get -u github.com/kurrik/json
	go get -u github.com/pressly/chi
	go get -u github.com/stretchr/testify/assert

check:
	@echo "1. formating code"
	@goimports -w .
	@echo "2. lint go code"
	@golint ./...

test:
	go test -v ./engine/...
	go test -v ./pkg/...

ci: setup check test
