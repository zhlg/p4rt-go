export CGO_ENABLED=0
export GO111MODULE=on

.PHONY: build


build-bmv2: # @HELP build the Go binaries for bmv2 and run all validations (default)
	go build -tags bmv2 -o build/_output/onos-control ./bin/main.go
	cp build/_output/onos-control /home/zhangl/go/src/github.com/onosproject/onos-control/build/_output/ 

build-tofino: # @HELP build the Go binaries for tofino and run all validations (default)
	go build -tags tofino -o build/_output/onos-control ./bin/main.go
	cp build/_output/onos-control /home/zhangl/go/src/github.com/onosproject/onos-control/build/_output/ 

clean: # @HELP remove all the build artifacts
	rm -rf ./build/_output ./vendor 
#	go clean -testcache github.com/onosproject/onos-config/...

copy: # @HELP copy the Go binaries to onos-control 
	cp build/_output/onos-control /home/zhangl/go/src/github.com/onosproject/onos-control/build/_output/ 
	
help:
	@grep -E '^.*: *# *@HELP' $(MAKEFILE_LIST) \
    | sort \
    | awk ' \
        BEGIN {FS = ": *# *@HELP"}; \
        {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}; \
    '
