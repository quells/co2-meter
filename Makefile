# TODO: add source files as dependencies
.PHONY: co2-meter
co2-meter:
	GOOS=linux GOARCH=arm GOARM=6 go build ./cmd/co2-meter
