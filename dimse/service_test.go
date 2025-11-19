package dimse

import (
	"context"
	"errors"
	"testing"

	"github.com/caio-sobreiro/dicomnet/dicom"
	"github.com/caio-sobreiro/dicomnet/interfaces"
	"github.com/caio-sobreiro/dicomnet/types"
)

// MockPDULayer is a mock implementation of PDULayer for testing
type MockPDULayer struct {
	SendDIMSEResponseFunc            func(presContextID byte, commandData []byte) error
	SendDIMSEResponseWithDatasetFunc func(presContextID byte, commandData []byte, datasetData []byte) error
	GetTransferSyntaxFunc            func(presContextID byte) (string, error)
	TransferSyntaxUID                string
}

func (m *MockPDULayer) SendDIMSEResponse(presContextID byte, commandData []byte) error {
	if m.SendDIMSEResponseFunc != nil {
		return m.SendDIMSEResponseFunc(presContextID, commandData)
	}
	return nil
}

func (m *MockPDULayer) SendDIMSEResponseWithDataset(presContextID byte, commandData []byte, datasetData []byte) error {
	if m.SendDIMSEResponseWithDatasetFunc != nil {
		return m.SendDIMSEResponseWithDatasetFunc(presContextID, commandData, datasetData)
	}
	return nil
}

func (m *MockPDULayer) GetTransferSyntax(presContextID byte) (string, error) {
	if m.GetTransferSyntaxFunc != nil {
		return m.GetTransferSyntaxFunc(presContextID)
	}
	return m.TransferSyntaxUID, nil
}

// MockServiceHandler is a mock implementation of ServiceHandler for testing
type MockServiceHandler struct {
	HandleDIMSEFunc func(ctx context.Context, msg *types.Message, data []byte, meta interfaces.MessageContext) (*types.Message, *dicom.Dataset, error)
}

func (m *MockServiceHandler) HandleDIMSE(ctx context.Context, msg *types.Message, data []byte, meta interfaces.MessageContext) (*types.Message, *dicom.Dataset, error) {
	if m.HandleDIMSEFunc != nil {
		return m.HandleDIMSEFunc(ctx, msg, data, meta)
	}
	// Default response
	return &types.Message{
		CommandField:              CEchoRSP,
		Status:                    StatusSuccess,
		CommandDataSetType:        0x0101,
		MessageIDBeingRespondedTo: msg.MessageID,
	}, nil, nil
}

func TestNewService(t *testing.T) {
	handler := &MockServiceHandler{}
	service := NewService(handler, nil)

	if service == nil {
		t.Fatal("Expected non-nil service")
	}

	if service.handler == nil {
		t.Error("Service handler not set")
	}
}

func TestService_HandleDIMSEMessage_CEchoNoDataset(t *testing.T) {
	// Create handler that returns simple C-ECHO response
	handler := &MockServiceHandler{
		HandleDIMSEFunc: func(ctx context.Context, msg *types.Message, data []byte, meta interfaces.MessageContext) (*types.Message, *dicom.Dataset, error) {
			return &types.Message{
				CommandField:              CEchoRSP,
				Status:                    StatusSuccess,
				CommandDataSetType:        0x0101,
				MessageIDBeingRespondedTo: msg.MessageID,
			}, nil, nil
		},
	}

	service := NewService(handler, nil)
	pduLayer := &MockPDULayer{
		TransferSyntaxUID: dicom.TransferSyntaxExplicitVRLittleEndian,
		SendDIMSEResponseWithDatasetFunc: func(presContextID byte, commandData []byte, datasetData []byte) error {
			if presContextID != 1 {
				t.Errorf("Expected context ID 1, got %d", presContextID)
			}
			if len(commandData) == 0 {
				t.Error("Expected command data")
			}
			return nil
		},
	}

	// Create C-ECHO request command
	msg := &types.Message{
		CommandField:        CEchoRQ,
		MessageID:           1,
		AffectedSOPClassUID: "1.2.840.10008.1.1",
		CommandDataSetType:  0x0101, // No dataset
	}
	commandData := createDIMSECommand(msg)

	// Send command (last fragment, no dataset)
	err := service.HandleDIMSEMessage(1, 0x03, commandData, pduLayer)
	if err != nil {
		t.Fatalf("HandleDIMSEMessage failed: %v", err)
	}
}

func TestService_HandleDIMSEMessage_WithDataset(t *testing.T) {
	// Create handler
	handler := &MockServiceHandler{
		HandleDIMSEFunc: func(ctx context.Context, msg *types.Message, data []byte, meta interfaces.MessageContext) (*types.Message, *dicom.Dataset, error) {
			// Verify dataset was received
			if len(data) == 0 {
				t.Error("Expected dataset data")
			}
			parsed, err := dicom.ParseDatasetWithTransferSyntax(data, meta.TransferSyntaxUID)
			if err != nil {
				t.Fatalf("Failed to parse dataset: %v", err)
			}
			return &types.Message{
				CommandField:              CFindRSP,
				Status:                    StatusSuccess,
				CommandDataSetType:        0x0000,
				MessageIDBeingRespondedTo: msg.MessageID,
			}, parsed, nil
		},
	}

	service := NewService(handler, nil)
	pduLayer := &MockPDULayer{
		TransferSyntaxUID: dicom.TransferSyntaxExplicitVRLittleEndian,
		SendDIMSEResponseWithDatasetFunc: func(presContextID byte, commandData []byte, datasetData []byte) error {
			if len(datasetData) == 0 {
				t.Error("Expected dataset in response")
			}
			return nil
		},
	}

	// Create C-FIND request command
	msg := &types.Message{
		CommandField:        CFindRQ,
		MessageID:           2,
		AffectedSOPClassUID: "1.2.840.10008.5.1.4.1.2.1.1",
		CommandDataSetType:  0x0000, // Has dataset
	}
	commandData := createDIMSECommand(msg)

	// Send command (last fragment)
	err := service.HandleDIMSEMessage(1, 0x03, commandData, pduLayer)
	if err != nil {
		t.Fatalf("HandleDIMSEMessage failed: %v", err)
	}

	// Send dataset (last fragment)
	datasetData := []byte{0x10, 0x00, 0x10, 0x00, 0x08, 0x00, 0x00, 0x00, 'T', 'E', 'S', 'T'}
	err = service.HandleDIMSEMessage(1, 0x02, datasetData, pduLayer)
	if err != nil {
		t.Fatalf("HandleDIMSEMessage failed: %v", err)
	}
}

func TestService_HandleDIMSEMessage_MultiFragment(t *testing.T) {
	handler := &MockServiceHandler{
		HandleDIMSEFunc: func(ctx context.Context, msg *types.Message, data []byte, meta interfaces.MessageContext) (*types.Message, *dicom.Dataset, error) {
			// Verify all fragments were received
			if len(data) < 20 {
				t.Errorf("Expected at least 20 bytes of data, got %d", len(data))
			}
			return &types.Message{
				CommandField:              CFindRSP,
				Status:                    StatusSuccess,
				CommandDataSetType:        0x0101,
				MessageIDBeingRespondedTo: msg.MessageID,
			}, nil, nil
		},
	}

	service := NewService(handler, nil)
	pduLayer := &MockPDULayer{TransferSyntaxUID: dicom.TransferSyntaxExplicitVRLittleEndian}

	// Create command
	msg := &types.Message{
		CommandField:        CFindRQ,
		MessageID:           3,
		AffectedSOPClassUID: "1.2.840.10008.5.1.4.1.2.1.1",
		CommandDataSetType:  0x0000,
	}
	commandData := createDIMSECommand(msg)

	// Send command (last fragment)
	err := service.HandleDIMSEMessage(1, 0x03, commandData, pduLayer)
	if err != nil {
		t.Fatalf("HandleDIMSEMessage failed: %v", err)
	}

	// Send dataset fragments
	fragment1 := []byte{0x10, 0x00, 0x10, 0x00, 0x08, 0x00, 0x00, 0x00, 'F', 'R', 'A', 'G'}
	err = service.HandleDIMSEMessage(1, 0x00, fragment1, pduLayer) // More fragments
	if err != nil {
		t.Fatalf("HandleDIMSEMessage failed: %v", err)
	}

	fragment2 := []byte{'M', 'E', 'N', 'T', '1', '2', '3', '4'}
	err = service.HandleDIMSEMessage(1, 0x02, fragment2, pduLayer) // Last fragment
	if err != nil {
		t.Fatalf("HandleDIMSEMessage failed: %v", err)
	}
}

func TestService_HandleDIMSEMessage_ParseError(t *testing.T) {
	handler := &MockServiceHandler{}
	service := NewService(handler, nil)
	pduLayer := &MockPDULayer{}

	// Send invalid command data (too short)
	invalidData := []byte{0x00, 0x01, 0x02}
	err := service.HandleDIMSEMessage(1, 0x03, invalidData, pduLayer)
	if err == nil {
		t.Error("Expected error for invalid command data")
	}
}

func TestService_HandleDIMSEMessage_HandlerError(t *testing.T) {
	// Create handler that returns an error
	handler := &MockServiceHandler{
		HandleDIMSEFunc: func(ctx context.Context, msg *types.Message, data []byte, meta interfaces.MessageContext) (*types.Message, *dicom.Dataset, error) {
			return nil, nil, errors.New("handler processing failed")
		},
	}

	service := NewService(handler, nil)
	pduLayer := &MockPDULayer{TransferSyntaxUID: dicom.TransferSyntaxExplicitVRLittleEndian}

	// Create valid command
	msg := &types.Message{
		CommandField:        CEchoRQ,
		MessageID:           4,
		AffectedSOPClassUID: "1.2.840.10008.1.1",
		CommandDataSetType:  0x0101,
	}
	commandData := createDIMSECommand(msg)

	// Send command
	err := service.HandleDIMSEMessage(1, 0x03, commandData, pduLayer)
	if err == nil {
		t.Error("Expected error from handler")
	}
	if err.Error() != "service handler failed: handler processing failed" {
		t.Errorf("Unexpected error message: %v", err)
	}
}

func TestService_HandleDIMSEMessage_PDULayerError(t *testing.T) {
	handler := &MockServiceHandler{
		HandleDIMSEFunc: func(ctx context.Context, msg *types.Message, data []byte, meta interfaces.MessageContext) (*types.Message, *dicom.Dataset, error) {
			return &types.Message{
				CommandField:              CEchoRSP,
				Status:                    StatusSuccess,
				CommandDataSetType:        0x0101,
				MessageIDBeingRespondedTo: msg.MessageID,
			}, nil, nil
		},
	}

	service := NewService(handler, nil)
	pduLayer := &MockPDULayer{
		TransferSyntaxUID: dicom.TransferSyntaxExplicitVRLittleEndian,
		SendDIMSEResponseWithDatasetFunc: func(presContextID byte, commandData []byte, datasetData []byte) error {
			return errors.New("PDU send failed")
		},
	}

	// Create valid command
	msg := &types.Message{
		CommandField:        CEchoRQ,
		MessageID:           5,
		AffectedSOPClassUID: "1.2.840.10008.1.1",
		CommandDataSetType:  0x0101,
	}
	commandData := createDIMSECommand(msg)

	// Send command
	err := service.HandleDIMSEMessage(1, 0x03, commandData, pduLayer)
	if err == nil {
		t.Error("Expected PDU layer error")
	}
	if err.Error() != "PDU send failed" {
		t.Errorf("Unexpected error message: %v", err)
	}
}

func TestService_CommandConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant uint16
		expected uint16
	}{
		{"C-STORE-RQ", CStoreRQ, 0x0001},
		{"C-STORE-RSP", CStoreRSP, 0x8001},
		{"C-FIND-RQ", CFindRQ, 0x0020},
		{"C-FIND-RSP", CFindRSP, 0x8020},
		{"C-MOVE-RQ", CMoveRQ, 0x0021},
		{"C-MOVE-RSP", CMoveRSP, 0x8021},
		{"C-ECHO-RQ", CEchoRQ, 0x0030},
		{"C-ECHO-RSP", CEchoRSP, 0x8030},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.expected {
				t.Errorf("%s = 0x%04x, want 0x%04x", tt.name, tt.constant, tt.expected)
			}
		})
	}
}

func TestService_StatusConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant uint16
		expected uint16
	}{
		{"Success", StatusSuccess, 0x0000},
		{"Pending", StatusPending, 0xFF00},
		{"Failure", StatusFailure, 0xC000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.expected {
				t.Errorf("Status%s = 0x%04x, want 0x%04x", tt.name, tt.constant, tt.expected)
			}
		})
	}
}
