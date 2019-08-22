package s1ap

// #cgo CFLAGS: -I./asn1
// #cgo LDFLAGS: -L/usr/local/lib -ls1ap
// #include "S1AP-PDU.h"
// #include "SuccessfulOutcome.h"
import "C"
import (
	"fmt"
	"log"
	"unsafe"
)

func S1SetupResponse() {
	pdu := (*C.S1AP_PDU_t)(C.calloc(C.sizeof_struct_S1AP_PDU, 1))
	pdu.present = C.S1AP_PDU_PR_successfulOutcome

	msg := (*C.SuccessfulOutcome_t)(C.calloc(C.sizeof_struct_SuccessfulOutcome, 1))
	msg.value.present = C.SuccessfulOutcome__value_PR_S1SetupResponse
	ptr1 := (**C.SuccessfulOutcome_t)(unsafe.Pointer(&pdu.choice))
	*ptr1 = msg

	val := (*C.S1SetupResponse_t)(C.calloc(C.sizeof_struct_S1SetupResponse, 1))
	ptr2 := (**C.S1SetupResponse_t)(unsafe.Pointer(&msg.value.choice))
	*ptr2 = val

	// encode
	buf := Encode(pdu)
	fmt.Println(buf)

	defer func() {
		C.free(unsafe.Pointer(val))
		C.free(unsafe.Pointer(msg))
		C.free(unsafe.Pointer(pdu))
	}()
}

const (
	MAX_SDU_LEN = 8192
)

func Encode(pdu *C.S1AP_PDU_t) []byte {
	var constraints *C.asn_per_constraints_t = nil
	buf := make([]byte, MAX_SDU_LEN)

	ret := C.aper_encode_to_buffer(
		&C.asn_DEF_S1AP_PDU,
		constraints,
		unsafe.Pointer(pdu),
		unsafe.Pointer(&buf[0]),
		MAX_SDU_LEN)

	if ret.encoded < 0 {
		log.Printf("XXX encode error")
	} else {
		log.Printf("XXX encode success", ret.encoded)
		buf = buf[:ret.encoded]
	}

	return buf
}
