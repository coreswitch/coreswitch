package s1ap

// #cgo CFLAGS: -I./asn1
// #cgo LDFLAGS: -L/usr/local/lib -ls1ap
// #include "S1AP-PDU.h"
import "C"
import (
	"unsafe"
)

func Decode(buf []byte) {
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
