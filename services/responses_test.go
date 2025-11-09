package services

import (
	"testing"

	"github.com/caio-sobreiro/dicomnet/dimse"
	"github.com/caio-sobreiro/dicomnet/types"
)

func TestResponseBuilder_CEchoResponse(t *testing.T) {
	request := &types.Message{
		CommandField: dimse.CEchoRQ,
		MessageID:    42,
	}

	builder := NewResponseBuilder(request)
	response := builder.CEchoResponse(dimse.StatusSuccess)

	if response.CommandField != dimse.CEchoRSP {
		t.Errorf("CommandField = 0x%04x, want 0x%04x", response.CommandField, dimse.CEchoRSP)
	}

	if response.MessageIDBeingRespondedTo != 42 {
		t.Errorf("MessageIDBeingRespondedTo = %d, want 42", response.MessageIDBeingRespondedTo)
	}

	if response.Status != dimse.StatusSuccess {
		t.Errorf("Status = 0x%04x, want success", response.Status)
	}

	if response.AffectedSOPClassUID != types.VerificationSOPClass {
		t.Errorf("AffectedSOPClassUID = %s, want Verification SOP Class", response.AffectedSOPClassUID)
	}

	if response.CommandDataSetType != 0x0101 {
		t.Errorf("CommandDataSetType = 0x%04x, want 0x0101", response.CommandDataSetType)
	}
}

func TestResponseBuilder_CFindResponse_Pending(t *testing.T) {
	request := &types.Message{
		CommandField:        dimse.CFindRQ,
		MessageID:           10,
		AffectedSOPClassUID: "1.2.840.10008.5.1.4.1.2.2.1",
	}

	builder := NewResponseBuilder(request)
	response := builder.CFindResponse(dimse.StatusPending, true)

	if response.CommandField != dimse.CFindRSP {
		t.Errorf("CommandField = 0x%04x, want 0x%04x", response.CommandField, dimse.CFindRSP)
	}

	if response.Status != dimse.StatusPending {
		t.Errorf("Status = 0x%04x, want pending", response.Status)
	}

	if response.CommandDataSetType != 0x0000 {
		t.Errorf("CommandDataSetType = 0x%04x, want 0x0000 (dataset present)", response.CommandDataSetType)
	}

	if response.AffectedSOPClassUID != request.AffectedSOPClassUID {
		t.Errorf("AffectedSOPClassUID not preserved from request")
	}
}

func TestResponseBuilder_CFindResponse_Success(t *testing.T) {
	request := &types.Message{
		CommandField:        dimse.CFindRQ,
		MessageID:           10,
		AffectedSOPClassUID: "1.2.840.10008.5.1.4.1.2.2.1",
	}

	builder := NewResponseBuilder(request)
	response := builder.CFindResponse(dimse.StatusSuccess, false)

	if response.Status != dimse.StatusSuccess {
		t.Errorf("Status = 0x%04x, want success", response.Status)
	}

	if response.CommandDataSetType != 0x0101 {
		t.Errorf("CommandDataSetType = 0x%04x, want 0x0101 (no dataset)", response.CommandDataSetType)
	}
}

func TestResponseBuilder_CMoveResponse(t *testing.T) {
	request := &types.Message{
		CommandField:        dimse.CMoveRQ,
		MessageID:           15,
		AffectedSOPClassUID: "1.2.840.10008.5.1.4.1.2.2.1",
	}

	completed := uint16(10)
	failed := uint16(2)
	warning := uint16(1)
	remaining := uint16(5)

	builder := NewResponseBuilder(request)
	response := builder.CMoveResponse(dimse.StatusPending, &completed, &failed, &warning, &remaining)

	if response.CommandField != dimse.CMoveRSP {
		t.Errorf("CommandField = 0x%04x, want 0x%04x", response.CommandField, dimse.CMoveRSP)
	}

	if response.Status != dimse.StatusPending {
		t.Errorf("Status = 0x%04x, want pending", response.Status)
	}

	if response.NumberOfCompletedSuboperations == nil || *response.NumberOfCompletedSuboperations != 10 {
		t.Errorf("NumberOfCompletedSuboperations = %v, want 10", response.NumberOfCompletedSuboperations)
	}

	if response.NumberOfFailedSuboperations == nil || *response.NumberOfFailedSuboperations != 2 {
		t.Errorf("NumberOfFailedSuboperations = %v, want 2", response.NumberOfFailedSuboperations)
	}

	if response.NumberOfWarningSuboperations == nil || *response.NumberOfWarningSuboperations != 1 {
		t.Errorf("NumberOfWarningSuboperations = %v, want 1", response.NumberOfWarningSuboperations)
	}

	if response.NumberOfRemainingSuboperations == nil || *response.NumberOfRemainingSuboperations != 5 {
		t.Errorf("NumberOfRemainingSuboperations = %v, want 5", response.NumberOfRemainingSuboperations)
	}
}

func TestResponseBuilder_CMoveResponse_NilCounters(t *testing.T) {
	request := &types.Message{
		CommandField: dimse.CMoveRQ,
		MessageID:    15,
	}

	builder := NewResponseBuilder(request)
	response := builder.CMoveResponse(dimse.StatusFailure, nil, nil, nil, nil)

	if response.NumberOfCompletedSuboperations != nil {
		t.Error("Expected nil NumberOfCompletedSuboperations")
	}

	if response.NumberOfFailedSuboperations != nil {
		t.Error("Expected nil NumberOfFailedSuboperations")
	}

	if response.NumberOfWarningSuboperations != nil {
		t.Error("Expected nil NumberOfWarningSuboperations")
	}

	if response.NumberOfRemainingSuboperations != nil {
		t.Error("Expected nil NumberOfRemainingSuboperations")
	}
}

func TestResponseBuilder_CStoreResponse(t *testing.T) {
	request := &types.Message{
		CommandField:        dimse.CStoreRQ,
		MessageID:           20,
		AffectedSOPClassUID: "1.2.840.10008.5.1.4.1.1.2",
	}

	builder := NewResponseBuilder(request)
	response := builder.CStoreResponse(dimse.StatusSuccess, "")

	if response.CommandField != dimse.CStoreRSP {
		t.Errorf("CommandField = 0x%04x, want 0x%04x", response.CommandField, dimse.CStoreRSP)
	}

	if response.Status != dimse.StatusSuccess {
		t.Errorf("Status = 0x%04x, want success", response.Status)
	}

	if response.AffectedSOPClassUID != request.AffectedSOPClassUID {
		t.Errorf("AffectedSOPClassUID not preserved from request")
	}

	if response.CommandDataSetType != 0x0101 {
		t.Errorf("CommandDataSetType = 0x%04x, want 0x0101", response.CommandDataSetType)
	}
}

func TestResponseBuilder_CStoreResponse_CustomUID(t *testing.T) {
	request := &types.Message{
		CommandField: dimse.CStoreRQ,
		MessageID:    20,
	}

	customUID := "1.2.3.4.5.6"
	builder := NewResponseBuilder(request)
	response := builder.CStoreResponse(dimse.StatusSuccess, customUID)

	if response.AffectedSOPClassUID != customUID {
		t.Errorf("AffectedSOPClassUID = %s, want %s", response.AffectedSOPClassUID, customUID)
	}
}

// Test helper functions

func TestNewCEchoResponse(t *testing.T) {
	request := &types.Message{
		CommandField: dimse.CEchoRQ,
		MessageID:    1,
	}

	response := NewCEchoResponse(request, dimse.StatusSuccess)

	if response.CommandField != dimse.CEchoRSP {
		t.Errorf("CommandField = 0x%04x, want 0x%04x", response.CommandField, dimse.CEchoRSP)
	}

	if response.Status != dimse.StatusSuccess {
		t.Errorf("Status = 0x%04x, want success", response.Status)
	}
}

func TestNewCFindPendingResponse(t *testing.T) {
	request := &types.Message{
		CommandField:        dimse.CFindRQ,
		MessageID:           1,
		AffectedSOPClassUID: "1.2.840.10008.5.1.4.1.2.2.1",
	}

	response := NewCFindPendingResponse(request)

	if response.Status != dimse.StatusPending {
		t.Errorf("Status = 0x%04x, want pending", response.Status)
	}

	if response.CommandDataSetType != 0x0000 {
		t.Errorf("CommandDataSetType = 0x%04x, want 0x0000 (dataset present)", response.CommandDataSetType)
	}
}

func TestNewCFindSuccessResponse(t *testing.T) {
	request := &types.Message{
		CommandField: dimse.CFindRQ,
		MessageID:    1,
	}

	response := NewCFindSuccessResponse(request)

	if response.Status != dimse.StatusSuccess {
		t.Errorf("Status = 0x%04x, want success", response.Status)
	}

	if response.CommandDataSetType != 0x0101 {
		t.Errorf("CommandDataSetType = 0x%04x, want 0x0101 (no dataset)", response.CommandDataSetType)
	}
}

func TestNewCFindErrorResponse(t *testing.T) {
	request := &types.Message{
		CommandField: dimse.CFindRQ,
		MessageID:    1,
	}

	response := NewCFindErrorResponse(request, dimse.StatusFailure)

	if response.Status != dimse.StatusFailure {
		t.Errorf("Status = 0x%04x, want failure", response.Status)
	}

	if response.CommandDataSetType != 0x0101 {
		t.Errorf("CommandDataSetType = 0x%04x, want 0x0101 (no dataset)", response.CommandDataSetType)
	}
}

func TestNewCMoveSuccessResponse(t *testing.T) {
	request := &types.Message{
		CommandField: dimse.CMoveRQ,
		MessageID:    1,
	}

	response := NewCMoveSuccessResponse(request, 10, 2, 1)

	if response.Status != dimse.StatusSuccess {
		t.Errorf("Status = 0x%04x, want success", response.Status)
	}

	if response.NumberOfCompletedSuboperations == nil || *response.NumberOfCompletedSuboperations != 10 {
		t.Error("NumberOfCompletedSuboperations incorrect")
	}

	if response.NumberOfFailedSuboperations == nil || *response.NumberOfFailedSuboperations != 2 {
		t.Error("NumberOfFailedSuboperations incorrect")
	}

	if response.NumberOfWarningSuboperations == nil || *response.NumberOfWarningSuboperations != 1 {
		t.Error("NumberOfWarningSuboperations incorrect")
	}

	if response.NumberOfRemainingSuboperations == nil || *response.NumberOfRemainingSuboperations != 0 {
		t.Error("NumberOfRemainingSuboperations should be 0")
	}
}

func TestNewCMovePendingResponse(t *testing.T) {
	request := &types.Message{
		CommandField: dimse.CMoveRQ,
		MessageID:    1,
	}

	response := NewCMovePendingResponse(request, 5, 1, 0, 10)

	if response.Status != dimse.StatusPending {
		t.Errorf("Status = 0x%04x, want pending", response.Status)
	}

	if response.NumberOfRemainingSuboperations == nil || *response.NumberOfRemainingSuboperations != 10 {
		t.Error("NumberOfRemainingSuboperations incorrect")
	}
}

func TestNewCMoveErrorResponse(t *testing.T) {
	request := &types.Message{
		CommandField: dimse.CMoveRQ,
		MessageID:    1,
	}

	response := NewCMoveErrorResponse(request, dimse.StatusFailure)

	if response.Status != dimse.StatusFailure {
		t.Errorf("Status = 0x%04x, want failure", response.Status)
	}

	if response.NumberOfCompletedSuboperations != nil {
		t.Error("Error response should have nil counters")
	}
}

func TestNewCStoreResponse(t *testing.T) {
	request := &types.Message{
		CommandField:        dimse.CStoreRQ,
		MessageID:           1,
		AffectedSOPClassUID: "1.2.840.10008.5.1.4.1.1.2",
	}

	response := NewCStoreResponse(request, dimse.StatusSuccess)

	if response.CommandField != dimse.CStoreRSP {
		t.Errorf("CommandField = 0x%04x, want 0x%04x", response.CommandField, dimse.CStoreRSP)
	}

	if response.Status != dimse.StatusSuccess {
		t.Errorf("Status = 0x%04x, want success", response.Status)
	}
}
