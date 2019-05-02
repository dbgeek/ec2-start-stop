.PHONY: deps clean build

deps:
	go get -u ./...

clean: 
	rm -rf ./start-stop/start-stop ./start-stop/start-stop.zip
	
build:
	cd start-stop; GOOS=linux GOARCH=amd64 go build -o start-stop ./; zip start-stop.zip start-stop
