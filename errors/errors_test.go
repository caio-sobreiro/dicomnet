package errors

import (
	"errors"
	"testing"
)

func TestAssociationError(t *testing.T) {
	err := NewAssociationError(
		RejectSourceServiceUser,
		RejectReasonCalledAETitleNotRecognized,
		"AE title mismatch",
	)

	if err.Source != RejectSourceServiceUser {
		t.Errorf("Source = %v, want %v", err.Source, RejectSourceServiceUser)
	}

	if err.Reason != RejectReasonCalledAETitleNotRecognized {
		t.Errorf("Reason = %v, want %v", err.Reason, RejectReasonCalledAETitleNotRecognized)
	}

	errMsg := err.Error()
	if errMsg == "" {
		t.Error("Error message should not be empty")
	}
}

func TestDIMSEError(t *testing.T) {
	tests := []struct {
		name      string
		status    uint16
		isSuccess bool
		isPending bool
		isWarning bool
		isFailure bool
	}{
		{"Success", 0x0000, true, false, false, false},
		{"Pending", 0xFF00, false, true, false, false},
		{"Warning", 0x0107, false, false, true, false},
		{"Failure", 0xC000, false, false, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewDIMSEError("C-STORE", tt.status, "test error")

			if err.IsSuccess() != tt.isSuccess {
				t.Errorf("IsSuccess() = %v, want %v", err.IsSuccess(), tt.isSuccess)
			}
			if err.IsPending() != tt.isPending {
				t.Errorf("IsPending() = %v, want %v", err.IsPending(), tt.isPending)
			}
			if err.IsWarning() != tt.isWarning {
				t.Errorf("IsWarning() = %v, want %v", err.IsWarning(), tt.isWarning)
			}
			if err.IsFailure() != tt.isFailure {
				t.Errorf("IsFailure() = %v, want %v", err.IsFailure(), tt.isFailure)
			}
		})
	}
}

func TestTimeoutError(t *testing.T) {
	err := NewTimeoutError("connection", "30s")

	if err.Operation != "connection" {
		t.Errorf("Operation = %v, want connection", err.Operation)
	}

	if !err.Timeout() {
		t.Error("Timeout() should return true")
	}

	errMsg := err.Error()
	if errMsg == "" {
		t.Error("Error message should not be empty")
	}
}

func TestNetworkError(t *testing.T) {
	innerErr := errors.New("connection refused")
	err := NewNetworkError("dial", innerErr)

	if err.Op != "dial" {
		t.Errorf("Op = %v, want dial", err.Op)
	}

	if !errors.Is(err, innerErr) {
		t.Error("Should unwrap to inner error")
	}
}

func TestPDUError(t *testing.T) {
	err := NewPDUError(0x04, "invalid PDU length")

	if err.PDUType != 0x04 {
		t.Errorf("PDUType = 0x%02X, want 0x04", err.PDUType)
	}

	errMsg := err.Error()
	if errMsg == "" {
		t.Error("Error message should not be empty")
	}
}

func TestAbortError(t *testing.T) {
	err := NewAbortError(0x02, 0x01)

	if err.Source != 0x02 {
		t.Errorf("Source = 0x%02X, want 0x02", err.Source)
	}

	if err.Reason != 0x01 {
		t.Errorf("Reason = 0x%02X, want 0x01", err.Reason)
	}

	errMsg := err.Error()
	if errMsg == "" {
		t.Error("Error message should not be empty")
	}
}

func TestAssociationRejectReasonString(t *testing.T) {
	tests := []struct {
		reason   AssociationRejectReason
		expected string
	}{
		{RejectReasonNoReasonGiven, "no-reason-given"},
		{RejectReasonApplicationContextNotSupported, "application-context-not-supported"},
		{RejectReasonCallingAETitleNotRecognized, "calling-ae-title-not-recognized"},
		{RejectReasonCalledAETitleNotRecognized, "called-ae-title-not-recognized"},
		{AssociationRejectReason(0xFF), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.reason.String(); got != tt.expected {
				t.Errorf("String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestAssociationRejectSourceString(t *testing.T) {
	tests := []struct {
		source   AssociationRejectSource
		expected string
	}{
		{RejectSourceServiceUser, "service-user"},
		{RejectSourceServiceProvider, "service-provider"},
		{AssociationRejectSource(0xFF), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.source.String(); got != tt.expected {
				t.Errorf("String() = %v, want %v", got, tt.expected)
			}
		})
	}
}
