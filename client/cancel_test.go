package client

import (
	"encoding/binary"
	"log/slog"
	"testing"

	"github.com/caio-sobreiro/dicomnet/dimse"
	"github.com/caio-sobreiro/dicomnet/types"
)

func TestSendCCancel(t *testing.T) {
	conn := newMockConn()
	assoc := &Association{
		conn:           conn,
		callingAETitle: "TEST_SCU",
		calledAETitle:  "TEST_SCP",
		maxPDULength:   16384,
		presentationCtxs: map[byte]*PresentationContext{
			9: {
				ID:             9,
				AbstractSyntax: types.StudyRootQueryRetrieveInformationModelFind,
				Accepted:       true,
			},
		},
		logger: slog.Default(),
	}

	// Send C-CANCEL for message ID 5
	err := assoc.SendCCancel(5, types.StudyRootQueryRetrieveInformationModelFind)
	if err != nil {
		t.Fatalf("SendCCancel returned error: %v", err)
	}

	// Verify the C-CANCEL command was sent
	data := conn.writeBuf.Bytes()
	if len(data) == 0 {
		t.Fatal("No data written to connection")
	}

	// Parse the PDU to verify it's a P-DATA-TF with C-CANCEL
	// PDU header: Type (1) + Reserved (1) + Length (4)
	if len(data) < 6 {
		t.Fatal("PDU too short")
	}

	pduType := data[0]
	if pduType != 0x04 {
		t.Fatalf("Expected P-DATA-TF PDU (0x04), got 0x%02x", pduType)
	}

	// Skip to PDV (6 bytes PDU header + 4 bytes PDV length + 1 byte context ID + 1 byte control header)
	if len(data) < 12 {
		t.Fatal("PDU too short for command data")
	}

	// Find CommandField in the DIMSE command
	// The command is in DICOM implicit VR little endian format
	commandData := data[12:] // Skip PDU and PDV headers

	// Search for CommandField tag (0000,0100)
	// In little endian: group=0x0000 → 00 00, element=0x0100 → 00 01
	commandFieldFound := false
	var commandField uint16

	for i := 0; i+10 <= len(commandData); i++ {
		group := binary.LittleEndian.Uint16(commandData[i : i+2])
		element := binary.LittleEndian.Uint16(commandData[i+2 : i+4])

		if group == 0x0000 && element == 0x0100 { // CommandField tag
			length := binary.LittleEndian.Uint32(commandData[i+4 : i+8])
			if length == 2 {
				commandField = binary.LittleEndian.Uint16(commandData[i+8 : i+10])
				commandFieldFound = true
				break
			}
		}
	}

	if !commandFieldFound {
		t.Fatal("CommandField not found in DIMSE command")
	}

	if commandField != dimse.CCancelRQ {
		t.Fatalf("CommandField = 0x%04x, want C-CANCEL-RQ (0x%04x)", commandField, dimse.CCancelRQ)
	}

	// Verify MessageIDBeingRespondedTo is present and equals 5
	// Tag (0000,0120)
	msgIDFound := false
	var msgIDBeingRespondedTo uint16

	for i := 0; i+10 <= len(commandData); i++ {
		group := binary.LittleEndian.Uint16(commandData[i : i+2])
		element := binary.LittleEndian.Uint16(commandData[i+2 : i+4])

		if group == 0x0000 && element == 0x0120 { // MessageIDBeingRespondedTo tag
			length := binary.LittleEndian.Uint32(commandData[i+4 : i+8])
			if length == 2 {
				msgIDBeingRespondedTo = binary.LittleEndian.Uint16(commandData[i+8 : i+10])
				msgIDFound = true
				break
			}
		}
	}

	if !msgIDFound {
		t.Fatal("MessageIDBeingRespondedTo not found in DIMSE command")
	}

	if msgIDBeingRespondedTo != 5 {
		t.Fatalf("MessageIDBeingRespondedTo = %d, want 5", msgIDBeingRespondedTo)
	}
}

func TestSendCCancelErrors(t *testing.T) {
	conn := newMockConn()
	assoc := &Association{
		conn:           conn,
		callingAETitle: "TEST_SCU",
		calledAETitle:  "TEST_SCP",
		maxPDULength:   16384,
		presentationCtxs: map[byte]*PresentationContext{
			9: {
				ID:             9,
				AbstractSyntax: types.StudyRootQueryRetrieveInformationModelFind,
				Accepted:       true,
			},
		},
		logger: slog.Default(),
	}

	// Test zero message ID
	err := assoc.SendCCancel(0, types.StudyRootQueryRetrieveInformationModelFind)
	if err == nil {
		t.Fatal("Expected error for zero messageID, got nil")
	}

	// Test empty SOP class UID
	err = assoc.SendCCancel(5, "")
	if err == nil {
		t.Fatal("Expected error for empty sopClassUID, got nil")
	}

	// Test unsupported SOP class UID (no presentation context)
	err = assoc.SendCCancel(5, "1.2.3.4.5.6")
	if err == nil {
		t.Fatal("Expected error for unsupported SOP class, got nil")
	}
}
