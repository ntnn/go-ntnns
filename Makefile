GO ?= go

check:
	$(GO) test -cover -shuffle=on -vet=all ./...
