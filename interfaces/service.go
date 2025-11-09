// Package interfaces contains all service and handler interfaces
package interfaces

import (
	"context"

	"github.com/caio-sobreiro/dicomnet/types"
)

// ServiceHandler interface for handling DIMSE operations
type ServiceHandler interface {
	HandleDIMSE(ctx context.Context, msg *types.Message, data []byte) (*types.Message, []byte, error)
}

// StreamingServiceHandler interface for multi-response DIMSE operations
type StreamingServiceHandler interface {
	HandleDIMSEStreaming(ctx context.Context, msg *types.Message, data []byte, responder ResponseSender) error
}

// ResponseSender interface for sending intermediate responses
type ResponseSender interface {
	SendResponse(msg *types.Message, data []byte) error
}

// CGetResponder interface for C-GET operations that need to send C-STORE sub-operations
type CGetResponder interface {
	ResponseSender
	// SendCStore sends a C-STORE sub-operation on the same association
	SendCStore(sopClassUID, sopInstanceUID string, data []byte) error
}

// DIMSEHandler interface for PDU layer to communicate with DIMSE layer
type DIMSEHandler interface {
	HandleDIMSEMessage(presContextID byte, msgCtrlHeader byte, data []byte, pduLayer PDULayer) error
}

// PDULayer interface for DIMSE layer to communicate with PDU layer
type PDULayer interface {
	SendDIMSEResponse(presContextID byte, commandData []byte) error
	SendDIMSEResponseWithDataset(presContextID byte, commandData []byte, dataset []byte) error
}
