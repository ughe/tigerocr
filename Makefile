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

install: all
	mv $(BIN)/* ${GOPATH}/bin/
clean:
	rm -rf $(objs)
clobber: clean
	ls $(BIN)/* | grep -v "\.py" | xargs rm
tar: clobber
	export COPYFILE_DISABLE=true && tar -czvf tigerocr.tar cmd/* editdist/* ocr/* go.* Makefile README.md
