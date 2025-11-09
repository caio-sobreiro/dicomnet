package client

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/caio-sobreiro/dicomnet/dicom"
	"github.com/caio-sobreiro/dicomnet/dimse"
	"github.com/caio-sobreiro/dicomnet/types"
)

func TestSendCGet(t *testing.T) {
	conn := &mockConn{
		readBuf:  bytes.NewBuffer(nil),
		writeBuf: bytes.NewBuffer(nil),
	}

	assoc := &Association{
		conn:           conn,
		callingAETitle: "TEST_SCU",
		calledAETitle:  "TEST_SCP",
		maxPDULength:   16384,
		presentationCtxs: map[byte]*PresentationContext{
			11: {
				ID:             11,
				AbstractSyntax: types.StudyRootQueryRetrieveInformationModelGet,
				Accepted:       true,
			},
		},
		logger: slog.Default(),
	}

	requestDataset := dicom.NewDataset()
	requestDataset.AddElement(dicom.Tag{Group: 0x0008, Element: 0x0052}, dicom.VR_CS, "STUDY")
	requestDataset.AddElement(dicom.Tag{Group: 0x0020, Element: 0x000D}, dicom.VR_UI, "1.2.3.4.5")

	// Setup response - pending with sub-operation counts
	remaining := uint16(5)
	completed := uint16(0)
	failed := uint16(0)
	warning := uint16(0)

	pendingCommand := buildCommandDataset(&types.Message{
		CommandField:                   dimse.CGetRSP,
		MessageIDBeingRespondedTo:      1,
		CommandDataSetType:             0x0101,
		Status:                         dimse.StatusPending,
		AffectedSOPClassUID:            types.StudyRootQueryRetrieveInformationModelGet,
		NumberOfRemainingSuboperations: &remaining,
		NumberOfCompletedSuboperations: &completed,
		NumberOfFailedSuboperations:    &failed,
		NumberOfWarningSuboperations:   &warning,
	})

	// Final response
	remaining = 0
	completed = 5
	finalCommand := buildCommandDataset(&types.Message{
		CommandField:                   dimse.CGetRSP,
		MessageIDBeingRespondedTo:      1,
		CommandDataSetType:             0x0101,
		Status:                         dimse.StatusSuccess,
		AffectedSOPClassUID:            types.StudyRootQueryRetrieveInformationModelGet,
		NumberOfRemainingSuboperations: &remaining,
		NumberOfCompletedSuboperations: &completed,
		NumberOfFailedSuboperations:    &failed,
		NumberOfWarningSuboperations:   &warning,
	})

	// Write responses to mock connection
	conn.readBuf.Write(buildPDataPDU(11, true, true, pendingCommand))
	conn.readBuf.Write(buildPDataPDU(11, true, true, finalCommand))

	// Execute C-GET
	req := &CGetRequest{
		MessageID: 1,
		Dataset:   requestDataset,
	}

	responses, err := assoc.SendCGet(req)
	if err != nil {
		t.Fatalf("SendCGet failed: %v", err)
	}

	if len(responses) != 2 {
		t.Fatalf("Expected 2 responses, got %d", len(responses))
	}

	// Check pending response
	if responses[0].Status != dimse.StatusPending {
		t.Errorf("First response status = 0x%04X, want 0x%04X", responses[0].Status, dimse.StatusPending)
	}
	if responses[0].NumberOfRemainingSuboperations == nil || *responses[0].NumberOfRemainingSuboperations != 5 {
		t.Error("Expected remaining sub-operations = 5")
	}

	// Check final response
	if responses[1].Status != dimse.StatusSuccess {
		t.Errorf("Final response status = 0x%04X, want 0x%04X", responses[1].Status, dimse.StatusSuccess)
	}
	if responses[1].NumberOfCompletedSuboperations == nil || *responses[1].NumberOfCompletedSuboperations != 5 {
		t.Error("Expected completed sub-operations = 5")
	}
	if responses[1].NumberOfRemainingSuboperations == nil || *responses[1].NumberOfRemainingSuboperations != 0 {
		t.Error("Expected remaining sub-operations = 0")
	}
}

func TestSendCGet_NilRequest(t *testing.T) {
	conn := &mockConn{
		readBuf:  bytes.NewBuffer(nil),
		writeBuf: bytes.NewBuffer(nil),
	}

	assoc := &Association{
		conn:   conn,
		logger: slog.Default(),
	}

	_, err := assoc.SendCGet(nil)
	if err == nil {
		t.Fatal("Expected error for nil request, got nil")
	}
}

func TestSendCGet_NilDataset(t *testing.T) {
	conn := &mockConn{
		readBuf:  bytes.NewBuffer(nil),
		writeBuf: bytes.NewBuffer(nil),
	}

	assoc := &Association{
		conn:   conn,
		logger: slog.Default(),
	}

	req := &CGetRequest{
		MessageID: 1,
		Dataset:   nil,
	}

	_, err := assoc.SendCGet(req)
	if err == nil {
		t.Fatal("Expected error for nil dataset, got nil")
	}
}
