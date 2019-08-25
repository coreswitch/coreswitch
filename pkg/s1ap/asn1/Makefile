all:
	@echo "make lib - make S1AP library"
	@echo "make install - install S1AP library"
	@echo "make gen - regenerate C file from ASN file"
	@echo "make clean - clean up"
	@echo "make s1ap - build S1AP APP"

lib:
	gcc -I. -c *.c
	ar -rc libs1ap.a *.o

gen:
	asn1c -pdu=all -fcompound-names -findirect-choice -fno-include-deps -no-gen-example r14.4.0/36413-e40.asn

install:
	cp libs1ap.a /usr/local/lib

s1ap:
	gcc -I. -c -o s1ap_build.o s1ap_build.c

clean:
	rm -f *.[cho] Makefile.am Makefile.am.libasncodec libs1ap.a
