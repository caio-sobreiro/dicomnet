package services

import (
	"context"
	"testing"

	"github.com/caio-sobreiro/dicomnet/dimse"
	"github.com/caio-sobreiro/dicomnet/types"
)

func TestNewEchoService(t *testing.T) {
	service := NewEchoService()
	if service == nil {
		t.Fatal("Expected non-nil service")
	}
}

func TestEchoService_HandleDIMSE(t *testing.T) {
	service := NewEchoService()
	ctx := context.Background()

	tests := []struct {
		name           string
		msg            *types.Message
		expectedStatus uint16
	}{
		{
			name: "Basic C-ECHO request",
			msg: &types.Message{
				CommandField:        dimse.CEchoRQ,
				MessageID:           1,
				AffectedSOPClassUID: types.VerificationSOPClass,
				CommandDataSetType:  0x0101,
			},
			expectedStatus: dimse.StatusSuccess,
		},
		{
			name: "C-ECHO with different message ID",
			msg: &types.Message{
				CommandField:        dimse.CEchoRQ,
				MessageID:           42,
				AffectedSOPClassUID: types.VerificationSOPClass,
				CommandDataSetType:  0x0101,
			},
			expectedStatus: dimse.StatusSuccess,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			respMsg, respData, err := service.HandleDIMSE(ctx, tt.msg, nil)

			if err != nil {
				t.Fatalf("HandleDIMSE() error = %v", err)
			}

			if respMsg == nil {
				t.Fatal("Expected non-nil response message")
			}

			if respMsg.CommandField != dimse.CEchoRSP {
				t.Errorf("CommandField = 0x%04x, want 0x%04x",
					respMsg.CommandField, dimse.CEchoRSP)
			}

			if respMsg.Status != tt.expectedStatus {
				t.Errorf("Status = 0x%04x, want 0x%04x",
					respMsg.Status, tt.expectedStatus)
			}

			if respMsg.MessageIDBeingRespondedTo != tt.msg.MessageID {
				t.Errorf("MessageIDBeingRespondedTo = %d, want %d",
					respMsg.MessageIDBeingRespondedTo, tt.msg.MessageID)
			}

			if respMsg.AffectedSOPClassUID != types.VerificationSOPClass {
				t.Errorf("AffectedSOPClassUID = %s, want %s",
					respMsg.AffectedSOPClassUID, types.VerificationSOPClass)
			}

			if respMsg.CommandDataSetType != 0x0101 {
				t.Errorf("CommandDataSetType = 0x%04x, want 0x0101",
					respMsg.CommandDataSetType)
			}

			if respData != nil {
				t.Error("Expected nil response data for C-ECHO")
			}
		})
	}
}

func TestEchoService_HandleDIMSE_WithContext(t *testing.T) {
	service := NewEchoService()

	// Test with cancellable context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	msg := &types.Message{
		CommandField:        dimse.CEchoRQ,
		MessageID:           1,
		AffectedSOPClassUID: types.VerificationSOPClass,
		CommandDataSetType:  0x0101,
	}

	respMsg, _, err := service.HandleDIMSE(ctx, msg, nil)
	if err != nil {
		t.Fatalf("HandleDIMSE() error = %v", err)
	}

	if respMsg.Status != dimse.StatusSuccess {
		t.Errorf("Status = 0x%04x, want success", respMsg.Status)
	}
}

func TestEchoService_HealthCheck(t *testing.T) {
	service := NewEchoService()
	ctx := context.Background()

	err := service.HealthCheck(ctx)
	if err != nil {
		t.Errorf("HealthCheck() error = %v, want nil", err)
	}
}

func TestEchoService_HealthCheck_WithCancelledContext(t *testing.T) {
	service := NewEchoService()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Echo service should still return healthy even with cancelled context
	// since it has no external dependencies
	err := service.HealthCheck(ctx)
	if err != nil {
		t.Errorf("HealthCheck() error = %v, want nil", err)
	}
}
