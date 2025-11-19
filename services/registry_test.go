package services

import (
	"context"
	"errors"
	"testing"

	"github.com/caio-sobreiro/dicomnet/dicom"
	"github.com/caio-sobreiro/dicomnet/dimse"
	"github.com/caio-sobreiro/dicomnet/interfaces"
	"github.com/caio-sobreiro/dicomnet/types"
)

// mockHandler implements interfaces.ServiceHandler
type mockHandler struct {
	handleFunc func(ctx context.Context, msg *types.Message, data []byte, meta interfaces.MessageContext) (*types.Message, *dicom.Dataset, error)
}

func (m *mockHandler) HandleDIMSE(ctx context.Context, msg *types.Message, data []byte, meta interfaces.MessageContext) (*types.Message, *dicom.Dataset, error) {
	if m.handleFunc != nil {
		return m.handleFunc(ctx, msg, data, meta)
	}
	return &types.Message{
		CommandField:              msg.CommandField | 0x8000,
		MessageIDBeingRespondedTo: msg.MessageID,
		Status:                    dimse.StatusSuccess,
	}, nil, nil
}

// mockStreamingHandler implements both interfaces.ServiceHandler and interfaces.StreamingServiceHandler
type mockStreamingHandler struct {
	handleFunc          func(ctx context.Context, msg *types.Message, data []byte, meta interfaces.MessageContext) (*types.Message, *dicom.Dataset, error)
	handleStreamingFunc func(ctx context.Context, msg *types.Message, data []byte, meta interfaces.MessageContext, responder interfaces.ResponseSender) error
}

func (m *mockStreamingHandler) HandleDIMSE(ctx context.Context, msg *types.Message, data []byte, meta interfaces.MessageContext) (*types.Message, *dicom.Dataset, error) {
	if m.handleFunc != nil {
		return m.handleFunc(ctx, msg, data, meta)
	}
	return &types.Message{
		CommandField:              msg.CommandField | 0x8000,
		MessageIDBeingRespondedTo: msg.MessageID,
		Status:                    dimse.StatusSuccess,
	}, nil, nil
}

func (m *mockStreamingHandler) HandleDIMSEStreaming(ctx context.Context, msg *types.Message, data []byte, meta interfaces.MessageContext, responder interfaces.ResponseSender) error {
	if m.handleStreamingFunc != nil {
		return m.handleStreamingFunc(ctx, msg, data, meta, responder)
	}
	return responder.SendResponse(&types.Message{
		CommandField:              msg.CommandField | 0x8000,
		MessageIDBeingRespondedTo: msg.MessageID,
		Status:                    dimse.StatusSuccess,
	}, nil, meta.TransferSyntaxUID)
}

// mockResponder implements interfaces.ResponseSender
type mockResponder struct {
	responses         []*types.Message
	datasets          []*dicom.Dataset
	transferSyntaxUID []string
	sendFunc          func(msg *types.Message, dataset *dicom.Dataset, transferSyntaxUID string) error
}

func (m *mockResponder) SendResponse(msg *types.Message, dataset *dicom.Dataset, transferSyntaxUID string) error {
	if m.sendFunc != nil {
		return m.sendFunc(msg, dataset, transferSyntaxUID)
	}
	m.responses = append(m.responses, msg)
	m.datasets = append(m.datasets, dataset)
	m.transferSyntaxUID = append(m.transferSyntaxUID, transferSyntaxUID)
	return nil
}

func testMeta() interfaces.MessageContext {
	return interfaces.MessageContext{
		PresentationContextID: 1,
		TransferSyntaxUID:     dicom.TransferSyntaxExplicitVRLittleEndian,
	}
}

func sampleDataset() *dicom.Dataset {
	dataset := dicom.NewDataset()
	dataset.AddElement(dicom.Tag{Group: 0x0008, Element: 0x0018}, dicom.VR_UI, "1.2.3.4.5")
	return dataset
}

func TestNewRegistry(t *testing.T) {
	registry := NewRegistry()
	if registry == nil {
		t.Fatal("Expected non-nil registry")
	}

	if registry.handlers == nil {
		t.Fatal("Expected initialized handlers map")
	}

	if len(registry.handlers) != 0 {
		t.Errorf("Expected empty handlers map, got %d handlers", len(registry.handlers))
	}
}

func TestRegistry_RegisterHandler(t *testing.T) {
	registry := NewRegistry()
	handler := &mockHandler{}

	registry.RegisterHandler(dimse.CEchoRQ, handler)

	if !registry.HasHandler(dimse.CEchoRQ) {
		t.Error("Handler should be registered for C-ECHO-RQ")
	}

	if registry.HasHandler(dimse.CFindRQ) {
		t.Error("Handler should not be registered for C-FIND-RQ")
	}
}

func TestRegistry_RegisterHandler_Replace(t *testing.T) {
	registry := NewRegistry()
	handler1 := &mockHandler{
		handleFunc: func(ctx context.Context, msg *types.Message, data []byte, meta interfaces.MessageContext) (*types.Message, *dicom.Dataset, error) {
			return &types.Message{Status: 1}, nil, nil
		},
	}
	handler2 := &mockHandler{
		handleFunc: func(ctx context.Context, msg *types.Message, data []byte, meta interfaces.MessageContext) (*types.Message, *dicom.Dataset, error) {
			return &types.Message{Status: 2}, nil, nil
		},
	}

	registry.RegisterHandler(dimse.CEchoRQ, handler1)
	registry.RegisterHandler(dimse.CEchoRQ, handler2)

	ctx := context.Background()
	msg := &types.Message{
		CommandField: dimse.CEchoRQ,
		MessageID:    1,
	}

	resp, _, _ := registry.HandleDIMSE(ctx, msg, nil, testMeta())
	if resp.Status != 2 {
		t.Errorf("Expected status 2 from second handler, got %d", resp.Status)
	}
}

func TestRegistry_UnregisterHandler(t *testing.T) {
	registry := NewRegistry()
	handler := &mockHandler{}

	registry.RegisterHandler(dimse.CEchoRQ, handler)
	if !registry.HasHandler(dimse.CEchoRQ) {
		t.Fatal("Handler should be registered")
	}

	registry.UnregisterHandler(dimse.CEchoRQ)
	if registry.HasHandler(dimse.CEchoRQ) {
		t.Error("Handler should be unregistered")
	}
}

func TestRegistry_HandleDIMSE(t *testing.T) {
	registry := NewRegistry()
	ctx := context.Background()

	handler := &mockHandler{
		handleFunc: func(ctx context.Context, msg *types.Message, data []byte, meta interfaces.MessageContext) (*types.Message, *dicom.Dataset, error) {
			return &types.Message{
				CommandField:              dimse.CEchoRSP,
				MessageIDBeingRespondedTo: msg.MessageID,
				Status:                    dimse.StatusSuccess,
			}, nil, nil
		},
	}

	registry.RegisterHandler(dimse.CEchoRQ, handler)

	msg := &types.Message{
		CommandField: dimse.CEchoRQ,
		MessageID:    42,
	}

	resp, dataset, err := registry.HandleDIMSE(ctx, msg, nil, testMeta())
	if err != nil {
		t.Fatalf("HandleDIMSE() error = %v", err)
	}

	if resp == nil {
		t.Fatal("Expected non-nil response")
	}

	if resp.CommandField != dimse.CEchoRSP {
		t.Errorf("CommandField = 0x%04x, want 0x%04x", resp.CommandField, dimse.CEchoRSP)
	}

	if resp.MessageIDBeingRespondedTo != 42 {
		t.Errorf("MessageIDBeingRespondedTo = %d, want 42", resp.MessageIDBeingRespondedTo)
	}

	if dataset != nil {
		t.Error("Expected nil data")
	}
}

func TestRegistry_HandleDIMSE_NoHandler(t *testing.T) {
	registry := NewRegistry()
	ctx := context.Background()

	msg := &types.Message{
		CommandField: dimse.CEchoRQ,
		MessageID:    1,
	}

	_, _, err := registry.HandleDIMSE(ctx, msg, nil, testMeta())
	if err == nil {
		t.Error("Expected error for unregistered command")
	}
}

func TestRegistry_HandleDIMSE_HandlerError(t *testing.T) {
	registry := NewRegistry()
	ctx := context.Background()

	expectedErr := errors.New("handler error")
	handler := &mockHandler{
		handleFunc: func(ctx context.Context, msg *types.Message, data []byte, meta interfaces.MessageContext) (*types.Message, *dicom.Dataset, error) {
			return nil, nil, expectedErr
		},
	}

	registry.RegisterHandler(dimse.CEchoRQ, handler)

	msg := &types.Message{
		CommandField: dimse.CEchoRQ,
		MessageID:    1,
	}

	_, _, err := registry.HandleDIMSE(ctx, msg, nil, testMeta())
	if err != expectedErr {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}
}

func TestRegistry_HandleDIMSEStreaming_StreamingHandler(t *testing.T) {
	registry := NewRegistry()
	ctx := context.Background()

	handler := &mockStreamingHandler{
		handleStreamingFunc: func(ctx context.Context, msg *types.Message, data []byte, meta interfaces.MessageContext, responder interfaces.ResponseSender) error {
			// Send multiple responses (simulating C-FIND)
			for i := 0; i < 3; i++ {
				if err := responder.SendResponse(&types.Message{
					CommandField:              dimse.CFindRSP,
					MessageIDBeingRespondedTo: msg.MessageID,
					Status:                    dimse.StatusPending,
				}, nil, meta.TransferSyntaxUID); err != nil {
					return err
				}
			}
			// Final response
			return responder.SendResponse(&types.Message{
				CommandField:              dimse.CFindRSP,
				MessageIDBeingRespondedTo: msg.MessageID,
				Status:                    dimse.StatusSuccess,
			}, nil, meta.TransferSyntaxUID)
		},
	}

	registry.RegisterHandler(dimse.CFindRQ, handler)

	msg := &types.Message{
		CommandField: dimse.CFindRQ,
		MessageID:    1,
	}

	responder := &mockResponder{}
	err := registry.HandleDIMSEStreaming(ctx, msg, nil, testMeta(), responder)
	if err != nil {
		t.Fatalf("HandleDIMSEStreaming() error = %v", err)
	}

	if len(responder.responses) != 4 {
		t.Errorf("Expected 4 responses, got %d", len(responder.responses))
	}

	// Check pending responses
	for i := 0; i < 3; i++ {
		if responder.responses[i].Status != dimse.StatusPending {
			t.Errorf("Response %d: expected pending status, got 0x%04x", i, responder.responses[i].Status)
		}
	}

	// Check final response
	if responder.responses[3].Status != dimse.StatusSuccess {
		t.Errorf("Final response: expected success status, got 0x%04x", responder.responses[3].Status)
	}
}

func TestRegistry_HandleDIMSEStreaming_NonStreamingHandler(t *testing.T) {
	registry := NewRegistry()
	ctx := context.Background()

	// Register a non-streaming handler
	handler := &mockHandler{
		handleFunc: func(ctx context.Context, msg *types.Message, data []byte, meta interfaces.MessageContext) (*types.Message, *dicom.Dataset, error) {
			return &types.Message{
				CommandField:              dimse.CEchoRSP,
				MessageIDBeingRespondedTo: msg.MessageID,
				Status:                    dimse.StatusSuccess,
			}, sampleDataset(), nil
		},
	}

	registry.RegisterHandler(dimse.CEchoRQ, handler)

	msg := &types.Message{
		CommandField: dimse.CEchoRQ,
		MessageID:    1,
	}

	responder := &mockResponder{}
	err := registry.HandleDIMSEStreaming(ctx, msg, nil, testMeta(), responder)
	if err != nil {
		t.Fatalf("HandleDIMSEStreaming() error = %v", err)
	}

	if len(responder.responses) != 1 {
		t.Errorf("Expected 1 response, got %d", len(responder.responses))
	}

	if len(responder.datasets) != 1 {
		t.Fatalf("Expected one dataset, got %d", len(responder.datasets))
	}

	if responder.datasets[0] == nil {
		t.Fatal("Expected non-nil dataset in response")
	}

	if element, ok := responder.datasets[0].GetElement(dicom.Tag{Group: 0x0008, Element: 0x0018}); !ok {
		t.Error("Expected SOP Instance UID element in dataset")
	} else if value := element.Value.(string); value != "1.2.3.4.5" {
		t.Errorf("Unexpected SOP Instance UID value: %s", value)
	}
}

func TestRegistry_HandleDIMSEStreaming_NoHandler(t *testing.T) {
	registry := NewRegistry()
	ctx := context.Background()

	msg := &types.Message{
		CommandField: dimse.CEchoRQ,
		MessageID:    1,
	}

	responder := &mockResponder{}
	err := registry.HandleDIMSEStreaming(ctx, msg, nil, testMeta(), responder)
	if err == nil {
		t.Error("Expected error for unregistered command")
	}
}

func TestRegistry_RegisteredCommands(t *testing.T) {
	registry := NewRegistry()
	handler := &mockHandler{}

	registry.RegisterHandler(dimse.CEchoRQ, handler)
	registry.RegisterHandler(dimse.CFindRQ, handler)
	registry.RegisterHandler(dimse.CStoreRQ, handler)

	commands := registry.RegisteredCommands()
	if len(commands) != 3 {
		t.Errorf("Expected 3 registered commands, got %d", len(commands))
	}

	// Check that all commands are present
	found := make(map[uint16]bool)
	for _, cmd := range commands {
		found[cmd] = true
	}

	expectedCommands := []uint16{dimse.CEchoRQ, dimse.CFindRQ, dimse.CStoreRQ}
	for _, expected := range expectedCommands {
		if !found[expected] {
			t.Errorf("Expected command 0x%04x not found in registered commands", expected)
		}
	}
}

func TestCreateErrorResponse(t *testing.T) {
	req := &types.Message{
		CommandField:        dimse.CEchoRQ,
		MessageID:           42,
		AffectedSOPClassUID: types.VerificationSOPClass,
	}

	resp := CreateErrorResponse(req, dimse.StatusFailure)

	if resp.CommandField != dimse.CEchoRSP {
		t.Errorf("CommandField = 0x%04x, want 0x%04x", resp.CommandField, dimse.CEchoRSP)
	}

	if resp.MessageIDBeingRespondedTo != 42 {
		t.Errorf("MessageIDBeingRespondedTo = %d, want 42", resp.MessageIDBeingRespondedTo)
	}

	if resp.Status != dimse.StatusFailure {
		t.Errorf("Status = 0x%04x, want 0x%04x", resp.Status, dimse.StatusFailure)
	}

	if resp.CommandDataSetType != 0x0101 {
		t.Errorf("CommandDataSetType = 0x%04x, want 0x0101", resp.CommandDataSetType)
	}

	if resp.AffectedSOPClassUID != req.AffectedSOPClassUID {
		t.Errorf("AffectedSOPClassUID = %s, want %s", resp.AffectedSOPClassUID, req.AffectedSOPClassUID)
	}
}

func TestRegistry_Integration(t *testing.T) {
	// Integration test simulating a real server setup
	registry := NewRegistry()
	ctx := context.Background()

	// Register echo service
	echoService := NewEchoService()
	registry.RegisterHandler(dimse.CEchoRQ, echoService)

	// Test C-ECHO
	echoMsg := &types.Message{
		CommandField:        dimse.CEchoRQ,
		MessageID:           1,
		AffectedSOPClassUID: types.VerificationSOPClass,
		CommandDataSetType:  0x0101,
	}

	resp, dataset, err := registry.HandleDIMSE(ctx, echoMsg, nil, testMeta())
	if err != nil {
		t.Fatalf("C-ECHO failed: %v", err)
	}

	if resp.Status != dimse.StatusSuccess {
		t.Errorf("C-ECHO status = 0x%04x, want success", resp.Status)
	}

	if dataset != nil {
		t.Error("C-ECHO should not return data")
	}
}
