CC=clang
DESTDIR=$${GOPATH-/usr/local}/bin

all: unutf_ſ
unutf_ſ: unutf_ſ.c
	$(CC) -O3 $< -o $@
install: all
	cp unutf_ſ $(DESTDIR)/unutf_ſ
	cp ser.sh $(DESTDIR)/ser.sh
	cp joins.sh $(DESTDIR)/joins.sh
clean:
	rm unutf_ſ
