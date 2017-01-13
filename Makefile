setup:
	go get -v ./...
	go get -u golang.org/x/tools/cmd/goimports
	go get -u github.com/golang/lint/golint

check:
	@echo "1. formating code"
	@goimports -w .
	@echo "2. lint go code"
	@golint ./...