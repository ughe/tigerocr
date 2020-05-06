CXX = g++
CXXFLAGS = -O3 -pedantic
srcs := $(patsubst %.cc,%,$(wildcard */*.cc))
objs := $(patsubst %.cc,%.o,$(wildcard */*.cc))

BIN = bin

%.o: %.cc %.h
	$(CXX) $(CXXFLAGS) -c $< -o $@
%: %.o
	$(CXX) $(CXXFLAGS) $^ -o $(BIN)/$(notdir $@)

all: golang $(objs) $(srcs)

golang:
	mkdir -p $(BIN)/
	go build -o $(BIN)/ ./...

install:
	cp $(BIN)/* ${GOPATH}/bin/
clean:
	rm -rf $(objs)
clobber: clean
	rm -rf $(BIN)/*
tar: clobber
	tar -czvf tigerocr.tar cmd/* editdist/* ocr/* go.* Makefile README.md
