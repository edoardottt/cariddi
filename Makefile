REPO=github.com/edoardottt/cariddi

fmt:
	@gofmt -s ./*
	@echo "Done."

remod:
	@rm -rf go.*
	@go mod init ${REPO}
	@go get
	@echo "Done."

update:
	@go get -u
	@go mod tidy -v
	@make unlinux
	@git pull
	@make linux
	@echo "Done."

linux:
	@go build -o ./cariddi
	@sudo mv ./cariddi /usr/bin/
	@echo "Done."

unlinux:
	@sudo rm -rf /usr/bin/cariddi
	@echo "Done."

test:
	@go test -v -race ./...
	@echo "Done."
