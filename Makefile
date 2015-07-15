default: install

OUT=$(GOBIN)/dockerf
PWD=$(shell pwd)
VENDOR_SRC=$(PWD)/vendor

install:
	export GOPATH=$(GOPATH); go build -gcflags "-N -l" -o $(OUT)
clean:
	rm $(OUT)
