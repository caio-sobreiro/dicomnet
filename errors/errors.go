// Package errors provides DICOM-specific error types for better error handling
package errors

import (
	"errors"
	"fmt"
)

// Common errors
var (
	ErrConnectionClosed    = errors.New("dicom: connection closed")
	ErrAssociationRejected = errors.New("dicom: association rejected")
	ErrInvalidPDU          = errors.New("dicom: invalid PDU")
	ErrUnsupportedTransfer = errors.New("dicom: unsupported transfer syntax")
	ErrNoPresentationCtx   = errors.New("dicom: no suitable presentation context")
	ErrInvalidMessage      = errors.New("dicom: invalid DIMSE message")
	ErrOperationCanceled   = errors.New("dicom: operation canceled")
)

// AssociationError represents an association-level error
type AssociationError struct {
	Reason AssociationRejectReason
	Source AssociationRejectSource
	Msg    string
}

func (e *AssociationError) Error() string {
	return fmt.Sprintf("association rejected: %s (source: %s, reason: %s)",
		e.Msg, e.Source, e.Reason)
}

// AssociationRejectReason represents why an association was rejected
type AssociationRejectReason byte

const (
	RejectReasonUnknown                        AssociationRejectReason = 0x00
	RejectReasonNoReasonGiven                  AssociationRejectReason = 0x01
	RejectReasonApplicationContextNotSupported AssociationRejectReason = 0x02
	RejectReasonCallingAETitleNotRecognized    AssociationRejectReason = 0x03
	RejectReasonCalledAETitleNotRecognized     AssociationRejectReason = 0x07
)

func (r AssociationRejectReason) String() string {
	switch r {
	case RejectReasonNoReasonGiven:
		return "no-reason-given"
	case RejectReasonApplicationContextNotSupported:
		return "application-context-not-supported"
	case RejectReasonCallingAETitleNotRecognized:
		return "calling-ae-title-not-recognized"
	case RejectReasonCalledAETitleNotRecognized:
		return "called-ae-title-not-recognized"
	default:
		return "unknown"
	}
}

// AssociationRejectSource represents who rejected the association
type AssociationRejectSource byte

const (
	RejectSourceUnknown         AssociationRejectSource = 0x00
	RejectSourceServiceUser     AssociationRejectSource = 0x01
	RejectSourceServiceProvider AssociationRejectSource = 0x02
)

func (s AssociationRejectSource) String() string {
	switch s {
	case RejectSourceServiceUser:
		return "service-user"
	case RejectSourceServiceProvider:
		return "service-provider"
	default:
		return "unknown"
	}
}

// NewAssociationError creates a new association error
func NewAssociationError(source AssociationRejectSource, reason AssociationRejectReason, msg string) *AssociationError {
	return &AssociationError{
		Source: source,
		Reason: reason,
		Msg:    msg,
	}
}

// DIMSEError represents a DIMSE operation error with status code
type DIMSEError struct {
	Status    uint16
	Operation string
	Msg       string
}

func (e *DIMSEError) Error() string {
	return fmt.Sprintf("DIMSE %s failed: %s (status: 0x%04X)", e.Operation, e.Msg, e.Status)
}

// NewDIMSEError creates a new DIMSE error
func NewDIMSEError(operation string, status uint16, msg string) *DIMSEError {
	return &DIMSEError{
		Operation: operation,
		Status:    status,
		Msg:       msg,
	}
}

// IsSuccess returns true if the DIMSE status indicates success
func (e *DIMSEError) IsSuccess() bool {
	return e.Status == 0x0000
}

// IsPending returns true if the DIMSE status indicates pending
func (e *DIMSEError) IsPending() bool {
	return e.Status == 0xFF00
}

// IsWarning returns true if the DIMSE status indicates a warning
func (e *DIMSEError) IsWarning() bool {
	return (e.Status & 0xFF00) == 0x0100
}

// IsFailure returns true if the DIMSE status indicates failure
func (e *DIMSEError) IsFailure() bool {
	return (e.Status&0xF000) == 0xC000 || (e.Status&0xF000) == 0xA000
}

// TimeoutError represents a timeout error
type TimeoutError struct {
	Operation string
	Duration  string
}

func (e *TimeoutError) Error() string {
	return fmt.Sprintf("timeout: %s exceeded %s", e.Operation, e.Duration)
}

func (e *TimeoutError) Timeout() bool {
	return true
}

// NewTimeoutError creates a new timeout error
func NewTimeoutError(operation, duration string) *TimeoutError {
	return &TimeoutError{
		Operation: operation,
		Duration:  duration,
	}
}

// NetworkError represents a network-level error
type NetworkError struct {
	Op  string
	Err error
}

func (e *NetworkError) Error() string {
	return fmt.Sprintf("network error during %s: %v", e.Op, e.Err)
}

func (e *NetworkError) Unwrap() error {
	return e.Err
}

// NewNetworkError creates a new network error
func NewNetworkError(op string, err error) *NetworkError {
	return &NetworkError{
		Op:  op,
		Err: err,
	}
}

// PDUError represents a PDU-level protocol error
type PDUError struct {
	PDUType byte
	Msg     string
}

func (e *PDUError) Error() string {
	return fmt.Sprintf("PDU error (type: 0x%02X): %s", e.PDUType, e.Msg)
}

// NewPDUError creates a new PDU error
func NewPDUError(pduType byte, msg string) *PDUError {
	return &PDUError{
		PDUType: pduType,
		Msg:     msg,
	}
}

// AbortError represents an A-ABORT PDU received
type AbortError struct {
	Source byte
	Reason byte
}

func (e *AbortError) Error() string {
	sourceStr := "unknown"
	if e.Source == 0x00 {
		sourceStr = "service-user"
	} else if e.Source == 0x02 {
		sourceStr = "service-provider"
	}

	return fmt.Sprintf("connection aborted by %s (reason: 0x%02X)", sourceStr, e.Reason)
}

// NewAbortError creates a new abort error
func NewAbortError(source, reason byte) *AbortError {
	return &AbortError{
		Source: source,
		Reason: reason,
	}
}
