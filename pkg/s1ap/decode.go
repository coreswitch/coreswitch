package s1ap

// #cgo CFLAGS: -I./asn1
// #cgo LDFLAGS: -L/usr/local/lib -ls1ap
// #include "S1AP-PDU.h"
import "C"
import (
	"fmt"

	"unsafe"

	"github.com/ishidawataru/sctp"
)

func Decode(buf []byte) {
	// file, err := os.Open("/home/kunihiro/packet")
	// if err != nil {
	// 	log.Fatal(err.Error())
	// }

	// buf, err := ioutil.ReadAll(file)
	// if err != nil {
	// 	log.Fatal(err.Error())
	// }
	// file.Close()

	message := C.malloc(C.sizeof_struct_S1AP_PDU)
	var opt_codec *C.asn_codec_ctx_t = nil

	C.aper_decode(
		opt_codec,
		&C.asn_DEF_S1AP_PDU,
		(*unsafe.Pointer)(&message),
		(unsafe.Pointer)(&buf[0]),
		(C.size_t)(len(buf)),
		0,
		0)

	C.xer_fprint(C.stdout, &C.asn_DEF_S1AP_PDU, unsafe.Pointer(message))
}

func ReadHeader(conn *sctp.SCTPConn) error {
	buf := make([]byte, 4)
	n, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Read err", err.Error())
		return err
	}
	fmt.Println("Readfull", n)
	return nil
}
