setup:
	@go get -u -v ./...
	go get -u golang.org/x/tools/cmd/goimports
	go get -u github.com/golang/lint/golint
	go get -u github.com/Sirupsen/logrus
	go get -u github.com/pressly/chi

check:
	@echo "1. formating code"
	@goimports -w .
	@echo "2. lint go code"
	@golint ./...

test:
	go test -v ./engine/...
	go test -v ./pkg/...

ci: setup check test
