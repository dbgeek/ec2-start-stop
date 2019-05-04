.PHONY: deps clean build

deps:
	go get -u ./...

clean: 
	rm -rf ./start-stop/dist
	
build:
	mkdir -p start-stop/dist; cd start-stop; GOOS=linux GOARCH=amd64 go build -o dist/start-stop ./; cd dist; zip start-stop.zip start-stop
