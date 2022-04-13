CGO_ENABLED=0
GOOS=linux

build: 
	CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) go build -o=./bin/covid-tracker .
update: build
	scp ./bin/covid-tracker root@dboat.cn:/root/tracker  && scp -r ./tpl root@dboat.cn:/root/tracker 
