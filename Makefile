CGO_ENABLED=0
GOOS=linux
USER=root
HOST=

build: 
	CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) go build -o=./bin/covid-tracker .
update: build
	scp ./bin/covid-tracker ${USER}@${HOST}:/root/tracker  && scp -r ./tpl ${USER}@${HOST}:/root/tracker 
