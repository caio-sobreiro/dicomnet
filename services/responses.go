package services

import (
	"github.com/caio-sobreiro/dicomnet/dimse"
	"github.com/caio-sobreiro/dicomnet/types"
)

// ResponseBuilder provides convenient methods for creating standard DIMSE response messages.
//
// These builders ensure that response messages are properly formatted according to the
// DICOM standard and include all required fields.
type ResponseBuilder struct {
	request *types.Message
}

// NewResponseBuilder creates a new response builder for the given request message.
//
// The builder will automatically populate common fields like MessageIDBeingRespondedTo
// and AffectedSOPClassUID from the request.
func NewResponseBuilder(request *types.Message) *ResponseBuilder {
	return &ResponseBuilder{request: request}
}

// CEchoResponse creates a C-ECHO-RSP message.
//
// Parameters:
//   - status: The response status (typically dimse.StatusSuccess)
//
// Returns a C-ECHO-RSP message with no dataset.
func (b *ResponseBuilder) CEchoResponse(status uint16) *types.Message {
	return &types.Message{
		CommandField:              dimse.CEchoRSP,
		MessageIDBeingRespondedTo: b.request.MessageID,
		AffectedSOPClassUID:       "1.2.840.10008.1.1", // Verification SOP Class
		CommandDataSetType:        0x0101,              // No Data Set Present
		Status:                    status,
	}
}

// CFindResponse creates a C-FIND-RSP message.
//
// Parameters:
//   - status: The response status (dimse.StatusSuccess, dimse.StatusPending, etc.)
//   - hasDataset: Whether this response includes a dataset
//
// For pending responses with matches, set status=dimse.StatusPending and hasDataset=true.
// For the final response, set status=dimse.StatusSuccess and hasDataset=false.
func (b *ResponseBuilder) CFindResponse(status uint16, hasDataset bool) *types.Message {
	datasetType := uint16(0x0101) // No dataset
	if hasDataset {
		datasetType = 0x0000 // Dataset present
	}

	return &types.Message{
		CommandField:              dimse.CFindRSP,
		MessageIDBeingRespondedTo: b.request.MessageID,
		AffectedSOPClassUID:       b.request.AffectedSOPClassUID,
		CommandDataSetType:        datasetType,
		Status:                    status,
	}
}

// CMoveResponse creates a C-MOVE-RSP message with sub-operation counts.
//
// Parameters:
//   - status: The response status (dimse.StatusSuccess, dimse.StatusPending, dimse.StatusFailure, etc.)
//   - completed: Number of completed sub-operations (can be nil if not applicable)
//   - failed: Number of failed sub-operations (can be nil if not applicable)
//   - warning: Number of sub-operations with warnings (can be nil if not applicable)
//   - remaining: Number of remaining sub-operations (can be nil if not applicable)
//
// For pending responses during C-STORE operations, use dimse.StatusPending.
// For the final response, use dimse.StatusSuccess.
func (b *ResponseBuilder) CMoveResponse(status uint16, completed, failed, warning, remaining *uint16) *types.Message {
	return &types.Message{
		CommandField:                     dimse.CMoveRSP,
		MessageIDBeingRespondedTo:        b.request.MessageID,
		AffectedSOPClassUID:              b.request.AffectedSOPClassUID,
		CommandDataSetType:               0x0101, // No Data Set Present
		Status:                           status,
		NumberOfCompletedSuboperations:   completed,
		NumberOfFailedSuboperations:      failed,
		NumberOfWarningSuboperations:     warning,
		NumberOfRemainingSuboperations:   remaining,
	}
}

// CStoreResponse creates a C-STORE-RSP message.
//
// Parameters:
//   - status: The response status (typically dimse.StatusSuccess or an error code)
//   - sopInstanceUID: The SOP Instance UID from the request (optional, will use request's if empty)
//
// Returns a C-STORE-RSP message with no dataset.
func (b *ResponseBuilder) CStoreResponse(status uint16, sopInstanceUID string) *types.Message {
	if sopInstanceUID == "" {
		sopInstanceUID = b.request.AffectedSOPClassUID
	}

	return &types.Message{
		CommandField:              dimse.CStoreRSP,
		MessageIDBeingRespondedTo: b.request.MessageID,
		AffectedSOPClassUID:       sopInstanceUID,
		CommandDataSetType:        0x0101, // No Data Set Present
		Status:                    status,
	}
}

// Helper functions for creating responses without a builder instance

// NewCEchoResponse creates a C-ECHO-RSP message from a request.
func NewCEchoResponse(request *types.Message, status uint16) *types.Message {
	return NewResponseBuilder(request).CEchoResponse(status)
}

// NewCFindPendingResponse creates a pending C-FIND-RSP message (with dataset).
func NewCFindPendingResponse(request *types.Message) *types.Message {
	return NewResponseBuilder(request).CFindResponse(dimse.StatusPending, true)
}

// NewCFindSuccessResponse creates a final success C-FIND-RSP message (no dataset).
func NewCFindSuccessResponse(request *types.Message) *types.Message {
	return NewResponseBuilder(request).CFindResponse(dimse.StatusSuccess, false)
}

// NewCFindErrorResponse creates an error C-FIND-RSP message.
func NewCFindErrorResponse(request *types.Message, status uint16) *types.Message {
	return NewResponseBuilder(request).CFindResponse(status, false)
}

// NewCMoveSuccessResponse creates a final success C-MOVE-RSP message with sub-operation counts.
func NewCMoveSuccessResponse(request *types.Message, completed, failed, warning uint16) *types.Message {
	remaining := uint16(0)
	return NewResponseBuilder(request).CMoveResponse(
		dimse.StatusSuccess,
		&completed,
		&failed,
		&warning,
		&remaining,
	)
}

// NewCMovePendingResponse creates a pending C-MOVE-RSP message with sub-operation counts.
func NewCMovePendingResponse(request *types.Message, completed, failed, warning, remaining uint16) *types.Message {
	return NewResponseBuilder(request).CMoveResponse(
		dimse.StatusPending,
		&completed,
		&failed,
		&warning,
		&remaining,
	)
}

// NewCMoveErrorResponse creates an error C-MOVE-RSP message.
func NewCMoveErrorResponse(request *types.Message, status uint16) *types.Message {
	return NewResponseBuilder(request).CMoveResponse(status, nil, nil, nil, nil)
}

// NewCStoreResponse creates a C-STORE-RSP message.
func NewCStoreResponse(request *types.Message, status uint16) *types.Message {
	return NewResponseBuilder(request).CStoreResponse(status, "")
}
