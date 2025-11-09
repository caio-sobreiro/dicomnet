// Package services provides reusable DICOM service implementations.
//
// This package contains standard DICOM service implementations that can be
// used by any DICOM server application. These implementations follow the
// DICOM standard and have no external backend dependencies.
package services

import (
	"context"
	"log/slog"

	"github.com/caio-sobreiro/dicomnet/dimse"
	"github.com/caio-sobreiro/dicomnet/types"
)

// EchoService handles C-ECHO verification requests.
//
// C-ECHO is used to verify connectivity and application-level communication
// between two DICOM Application Entities (AEs). It's the DICOM equivalent
// of a "ping" operation.
//
// The C-ECHO service is stateless and requires no external dependencies.
// It simply echoes back a success response to verify that the DICOM
// application entity is operational.
type EchoService struct {
	// No configuration or dependencies needed for echo service
}

// NewEchoService creates a new C-ECHO service instance.
//
// The echo service is stateless and has no configuration options.
func NewEchoService() *EchoService {
	return &EchoService{}
}

// HandleDIMSE processes a C-ECHO request and returns a success response.
//
// According to DICOM standard PS3.4, C-ECHO has no dataset and simply
// returns a status indicating whether the AE is operational.
//
// This method implements the interfaces.ServiceHandler interface.
//
// Parameters:
//   - ctx: Context for cancellation and deadlines
//   - msg: The incoming C-ECHO-RQ DIMSE message
//   - data: Dataset (always empty for C-ECHO)
//
// Returns:
//   - Response message (C-ECHO-RSP) with success status
//   - Response dataset (always nil for C-ECHO)
//   - Error (always nil for successful echo)
func (s *EchoService) HandleDIMSE(ctx context.Context, msg *types.Message, data []byte) (*types.Message, []byte, error) {
	slog.DebugContext(ctx, "Processing C-ECHO request",
		"message_id", msg.MessageID,
		"affected_sop_class", msg.AffectedSOPClassUID)

	// Create C-ECHO-RSP according to DICOM PS3.7
	response := &types.Message{
		CommandField:              dimse.CEchoRSP,
		MessageIDBeingRespondedTo: msg.MessageID,
		AffectedSOPClassUID:       "1.2.840.10008.1.1", // Verification SOP Class UID
		CommandDataSetType:        0x0101,              // No Data Set Present
		Status:                    dimse.StatusSuccess,
	}

	slog.InfoContext(ctx, "C-ECHO request successful",
		"message_id", msg.MessageID)

	return response, nil, nil
}

// HealthCheck verifies that the echo service is operational.
//
// Since echo service is stateless with no external dependencies,
// this always returns healthy.
func (s *EchoService) HealthCheck(ctx context.Context) error {
	return nil
}
