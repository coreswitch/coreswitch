/*
 * Generated by asn1c-0.9.29 (http://lionet.info/asn1c)
 * From ASN.1 module "S1AP-IEs"
 * 	found in "r14.4.0/36413-e40.asn"
 * 	`asn1c -pdu=all -fcompound-names -findirect-choice -fno-include-deps -no-gen-example`
 */

#ifndef	_RelayNode_Indicator_H_
#define	_RelayNode_Indicator_H_


#include <asn_application.h>

/* Including external dependencies */
#include <NativeEnumerated.h>

#ifdef __cplusplus
extern "C" {
#endif

/* Dependencies */
typedef enum RelayNode_Indicator {
	RelayNode_Indicator_true	= 0
	/*
	 * Enumeration is extensible
	 */
} e_RelayNode_Indicator;

/* RelayNode-Indicator */
typedef long	 RelayNode_Indicator_t;

/* Implementation */
extern asn_TYPE_descriptor_t asn_DEF_RelayNode_Indicator;
asn_struct_free_f RelayNode_Indicator_free;
asn_struct_print_f RelayNode_Indicator_print;
asn_constr_check_f RelayNode_Indicator_constraint;
ber_type_decoder_f RelayNode_Indicator_decode_ber;
der_type_encoder_f RelayNode_Indicator_encode_der;
xer_type_decoder_f RelayNode_Indicator_decode_xer;
xer_type_encoder_f RelayNode_Indicator_encode_xer;
oer_type_decoder_f RelayNode_Indicator_decode_oer;
oer_type_encoder_f RelayNode_Indicator_encode_oer;
per_type_decoder_f RelayNode_Indicator_decode_uper;
per_type_encoder_f RelayNode_Indicator_encode_uper;
per_type_decoder_f RelayNode_Indicator_decode_aper;
per_type_encoder_f RelayNode_Indicator_encode_aper;

#ifdef __cplusplus
}
#endif

#endif	/* _RelayNode_Indicator_H_ */
#include <asn_internal.h>