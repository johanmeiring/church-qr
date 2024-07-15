build:
	go build -tags netgo -ldflags '-s -w' -o church-qr

.PHONY: run
run:
	./church-qr
