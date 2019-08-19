all:
	@echo "make lib - make S1AP library"
	@echo "make install - install S1AP library"
	@echo "make gen - regenerate C file from ASN file"
	@echo "make clean - clean up"

lib:
	gcc -I. -c *.c
	ar -rc libs1ap.a *.o

gen:
	asn1c -pdu=all -fcompound-names -findirect-choice -fno-include-deps -no-gen-example r14.4.0/36413-e40.asn

install:
	cp libs1ap.a /usr/local/lib

clean:
	rm -f *.[cho] Makefile.am Makefile.am.libasncodec libs1ap.a
