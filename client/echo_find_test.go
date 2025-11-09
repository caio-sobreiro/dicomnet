package client

import (
	"encoding/binary"
	"log/slog"
	"testing"

	"github.com/caio-sobreiro/dicomnet/dicom"
	"github.com/caio-sobreiro/dicomnet/dimse"
	"github.com/caio-sobreiro/dicomnet/types"
)

func TestSendCEcho(t *testing.T) {
	conn := newMockConn()
	assoc := &Association{
		conn:           conn,
		callingAETitle: "TEST_SCU",
		calledAETitle:  "TEST_SCP",
		maxPDULength:   16384,
		presentationCtxs: map[byte]*PresentationContext{
			7: {
				ID:             7,
				AbstractSyntax: types.VerificationSOPClass,
				Accepted:       true,
			},
		},
		logger: slog.Default(),
	}

	command := buildCommandDataset(&types.Message{
		CommandField:              dimse.CEchoRSP,
		MessageIDBeingRespondedTo: 1,
		CommandDataSetType:        0x0101,
		Status:                    dimse.StatusSuccess,
		AffectedSOPClassUID:       types.VerificationSOPClass,
	})

	conn.readBuf.Write(buildPDataPDU(7, true, true, command))

	resp, err := assoc.SendCEcho(1)
	if err != nil {
		t.Fatalf("SendCEcho returned error: %v", err)
	}

	if resp.Status != dimse.StatusSuccess {
		t.Fatalf("C-ECHO status = 0x%04X, want success", resp.Status)
	}

	if resp.MessageID != 1 {
		t.Fatalf("C-ECHO message ID = %d, want 1", resp.MessageID)
	}

	if conn.writeBuf.Len() == 0 {
		t.Fatal("expected C-ECHO request to be written to connection")
	}
}

func TestSendCFind(t *testing.T) {
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

	requestDataset := dicom.NewDataset()
	requestDataset.AddElement(dicom.Tag{Group: 0x0008, Element: 0x0052}, dicom.VR_CS, "STUDY")
	requestDataset.AddElement(dicom.Tag{Group: 0x0010, Element: 0x0010}, dicom.VR_PN, "DOE^JOHN")

	matchDataset := dicom.NewDataset()
	matchDataset.AddElement(dicom.Tag{Group: 0x0010, Element: 0x0010}, dicom.VR_PN, "DOE^JOHN")
	matchDatasetBytes := matchDataset.EncodeDataset()

	pendingCommand := buildCommandDataset(&types.Message{
		CommandField:              dimse.CFindRSP,
		MessageIDBeingRespondedTo: 2,
		CommandDataSetType:        0x0000,
		Status:                    dimse.StatusPending,
		AffectedSOPClassUID:       types.StudyRootQueryRetrieveInformationModelFind,
	})

	finalCommand := buildCommandDataset(&types.Message{
		CommandField:              dimse.CFindRSP,
		MessageIDBeingRespondedTo: 2,
		CommandDataSetType:        0x0101,
		Status:                    dimse.StatusSuccess,
		AffectedSOPClassUID:       types.StudyRootQueryRetrieveInformationModelFind,
	})

	conn.readBuf.Write(buildPDataPDU(9, true, true, pendingCommand))
	conn.readBuf.Write(buildPDataPDU(9, false, true, matchDatasetBytes))
	conn.readBuf.Write(buildPDataPDU(9, true, true, finalCommand))

	responses, err := assoc.SendCFind(&CFindRequest{
		MessageID: 2,
		Dataset:   requestDataset,
	})
	if err != nil {
		t.Fatalf("SendCFind returned error: %v", err)
	}

	if len(responses) != 2 {
		t.Fatalf("expected 2 responses, got %d", len(responses))
	}

	if responses[0].Status != dimse.StatusPending {
		t.Fatalf("first response status = 0x%04X, want pending", responses[0].Status)
	}

	if responses[0].Dataset == nil {
		t.Fatal("expected dataset in pending response")
	}

	if name := responses[0].Dataset.GetString(dicom.Tag{Group: 0x0010, Element: 0x0010}); name != "DOE^JOHN" {
		t.Fatalf("patient name = %s, want DOE^JOHN", name)
	}

	if responses[1].Status != dimse.StatusSuccess {
		t.Fatalf("final response status = 0x%04X, want success", responses[1].Status)
	}

	if responses[1].Dataset != nil {
		t.Fatal("final response should not contain dataset")
	}
}

func buildCommandDataset(msg *types.Message) []byte {
	var body []byte

	if msg.AffectedSOPClassUID != "" {
		value := []byte(msg.AffectedSOPClassUID)
		if len(value)%2 == 1 {
			value = append(value, 0x00)
		}
		body = dimse.AppendImplicitElement(body, 0x0000, 0x0002, value)
	}

	if msg.RequestedSOPClassUID != "" {
		value := []byte(msg.RequestedSOPClassUID)
		if len(value)%2 == 1 {
			value = append(value, 0x00)
		}
		body = dimse.AppendImplicitElement(body, 0x0000, 0x0003, value)
	}

	if msg.CommandField != 0 {
		buf := make([]byte, 2)
		binary.LittleEndian.PutUint16(buf, msg.CommandField)
		body = dimse.AppendImplicitElement(body, 0x0000, 0x0100, buf)
	}

	if msg.MessageID != 0 {
		buf := make([]byte, 2)
		binary.LittleEndian.PutUint16(buf, msg.MessageID)
		body = dimse.AppendImplicitElement(body, 0x0000, 0x0110, buf)
	}

	if msg.MessageIDBeingRespondedTo != 0 {
		buf := make([]byte, 2)
		binary.LittleEndian.PutUint16(buf, msg.MessageIDBeingRespondedTo)
		body = dimse.AppendImplicitElement(body, 0x0000, 0x0120, buf)
	}

	buf := make([]byte, 2)
	binary.LittleEndian.PutUint16(buf, msg.CommandDataSetType)
	body = dimse.AppendImplicitElement(body, 0x0000, 0x0800, buf)

	if (msg.CommandField&0x8000) != 0 || msg.Status != 0 {
		statusBuf := make([]byte, 2)
		binary.LittleEndian.PutUint16(statusBuf, msg.Status)
		body = dimse.AppendImplicitElement(body, 0x0000, 0x0900, statusBuf)
	}

	// C-GET/C-MOVE sub-operation counters
	if msg.NumberOfRemainingSuboperations != nil {
		buf := make([]byte, 2)
		binary.LittleEndian.PutUint16(buf, *msg.NumberOfRemainingSuboperations)
		body = dimse.AppendImplicitElement(body, 0x0000, 0x1020, buf)
	}

	if msg.NumberOfCompletedSuboperations != nil {
		buf := make([]byte, 2)
		binary.LittleEndian.PutUint16(buf, *msg.NumberOfCompletedSuboperations)
		body = dimse.AppendImplicitElement(body, 0x0000, 0x1021, buf)
	}

	if msg.NumberOfFailedSuboperations != nil {
		buf := make([]byte, 2)
		binary.LittleEndian.PutUint16(buf, *msg.NumberOfFailedSuboperations)
		body = dimse.AppendImplicitElement(body, 0x0000, 0x1022, buf)
	}

	if msg.NumberOfWarningSuboperations != nil {
		buf := make([]byte, 2)
		binary.LittleEndian.PutUint16(buf, *msg.NumberOfWarningSuboperations)
		body = dimse.AppendImplicitElement(body, 0x0000, 0x1023, buf)
	}

	groupLength := make([]byte, 4)
	binary.LittleEndian.PutUint32(groupLength, uint32(len(body)))

	command := dimse.AppendImplicitElement(nil, 0x0000, 0x0000, groupLength)
	command = append(command, body...)

	return command
}

func buildPDataPDU(contextID byte, isCommand bool, isLast bool, data []byte) []byte {
	pdvLength := uint32(len(data) + 2)

	payload := make([]byte, 0, len(data)+6)

	pdvHeader := make([]byte, 4)
	binary.BigEndian.PutUint32(pdvHeader, pdvLength)
	payload = append(payload, pdvHeader...)
	payload = append(payload, contextID)

	control := byte(0)
	if isCommand {
		control |= 0x01
	}
	if isLast {
		control |= 0x02
	}
	payload = append(payload, control)
	payload = append(payload, data...)

	header := make([]byte, 6)
	header[0] = dimsePDataTF
	header[1] = 0x00
	binary.BigEndian.PutUint32(header[2:6], uint32(len(payload)))

	return append(header, payload...)
}

const dimsePDataTF = byte(0x04)
