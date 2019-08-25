#include "S1AP-PDU.h"
#include "SuccessfulOutcome.h"
#include "ProtocolIE-Field.h"
#include "ServedGUMMEIsItem.h"

#define PLMN_ID_LEN 3

void
s1ap_buffer_to_OCTET_STRING(void *buf, int size, TBCD_STRING_t *tbcd_string)
{
  tbcd_string->size = size;
  tbcd_string->buf = calloc(tbcd_string->size, 1);

  memcpy(tbcd_string->buf, buf, size);
}

void
S1SetupResponseBuild(S1AP_PDU_t *pdu, int num_served_gummei) {
  // S1AP_PDU_t pdu;
  SuccessfulOutcome_t *outcome = calloc(sizeof(SuccessfulOutcome_t), 1);
  S1SetupResponse_t *response = NULL;
  S1SetupResponseIEs_t *ie = NULL;
  ServedGUMMEIs_t *gmmei = NULL;
  ServedGUMMEIsItem_t *gmmei_item = NULL;
  RelativeMMECapacity_t *relative = NULL;

  memset(pdu, 0, sizeof(S1AP_PDU_t));
  pdu->present = S1AP_PDU_PR_successfulOutcome;
  pdu->choice.successfulOutcome = outcome;

  outcome->procedureCode = ProcedureCode_id_S1Setup;
  outcome->criticality = Criticality_reject;
  outcome->value.present = SuccessfulOutcome__value_PR_S1SetupResponse;

  response = &outcome->value.choice.S1SetupResponse;

  // ProtocolIEs for served GUMMEI.
  ie = calloc(sizeof(S1SetupResponseIEs_t), 1);
  ASN_SEQUENCE_ADD(&response->protocolIEs, ie);

  // Served GUMMEI.
  ie->id = ProtocolIE_ID_id_ServedGUMMEIs;
  ie->criticality = Criticality_reject;
  ie->value.present = S1SetupResponseIEs__value_PR_ServedGUMMEIs;

  // GMMEI and GMMEI items.
  gmmei = &ie->value.choice.ServedGUMMEIs;
  gmmei_item = calloc(sizeof(ServedGUMMEIsItem_t), 1);

  // PLMN.
  PLMNidentity_t *plmn = calloc(sizeof(PLMNidentity_t), 1);
  unsigned char plmn_data[3] = { 0x02, 0xf8, 0x39 };
  s1ap_buffer_to_OCTET_STRING(plmn_data, PLMN_ID_LEN, plmn);
  ASN_SEQUENCE_ADD(&gmmei_item->servedPLMNs.list, plmn);

  // Group ID.
  MME_Group_ID_t *group = calloc(sizeof(MME_Group_ID_t), 1);
  unsigned char group_data[2] = { 0x00, 0x04 };
  s1ap_buffer_to_OCTET_STRING(group_data, 2, group);
  ASN_SEQUENCE_ADD(&gmmei_item->servedGroupIDs.list, group);

  // MME Code.
  MME_Code_t *mme_code = calloc(sizeof(MME_Code_t), 1);
  unsigned char mme_code_data[2] = { 0x01 };
  s1ap_buffer_to_OCTET_STRING(mme_code_data, 1, mme_code);
  ASN_SEQUENCE_ADD(&gmmei_item->servedMMECs.list, mme_code);

  ASN_SEQUENCE_ADD(&gmmei->list, gmmei_item);

  /* ie = calloc(sizeof(S1SetupResponseIEs_t), 1); */
  /* ASN_SEQUENCE_ADD(&response->protocolIEs, ie); */

  /* ie->id = ProtocolIE_ID_id_RelativeMMECapacity; */
  /* ie->criticality = Criticality_ignore; */
  /* ie->value.present = S1SetupResponseIEs__value_PR_RelativeMMECapacity; */

  /* relative = &ie->value.choice.RelativeMMECapacity; */

  /* for (int i = 0; i < num_served_gummei; i++) { */
  /*   ; */
  /* } */
}

void
S1SetupResponseFree() {
}
