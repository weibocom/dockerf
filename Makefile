default: install

OUT=$(GOBIN)/dockerf

install:
	godep go build -gcflags "-N -l" -o $(OUT)
clean:
	rm $(OUT)
