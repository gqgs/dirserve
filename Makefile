
install: main

generate:
	go generate ./...

main: generate
	go install
