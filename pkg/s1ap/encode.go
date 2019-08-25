package s1ap

// #cgo CFLAGS: -I./asn1
// #cgo LDFLAGS: -L/usr/local/lib -ls1ap
// #include "S1AP-PDU.h"
// #include "SuccessfulOutcome.h"
// #include "s1ap_build.h"
import "C"
import (
	"fmt"
	"log"
	"unsafe"
)

func S1SetupResponse() ([]byte, error) {
	pdu := (*C.S1AP_PDU_t)(C.calloc(C.sizeof_struct_S1AP_PDU, 1))
	C.S1SetupResponseBuild(pdu, 0)

	// encode
	return Encode(pdu)
}

const (
	MAX_SDU_LEN = 8192
)

func Encode(pdu *C.S1AP_PDU_t) ([]byte, error) {
	var constraints *C.asn_per_constraints_t = nil
	buf := make([]byte, MAX_SDU_LEN)

	ret := C.aper_encode_to_buffer(
		&C.asn_DEF_S1AP_PDU,
		constraints,
		unsafe.Pointer(pdu),
		unsafe.Pointer(&buf[0]),
		MAX_SDU_LEN)

	if ret.encoded < 0 {
		return nil, fmt.Errorf("Encode() error %v", ret)
	}
	len := ret.encoded >> 3
	log.Printf("Encode() success", ret.encoded, len)
	buf = buf[:len]

	return buf, nil
}
