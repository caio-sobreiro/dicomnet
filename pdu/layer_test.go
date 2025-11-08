package pdu

import (
	"net"
	"testing"
)

// MockConn is a mock implementation of net.Conn for testing
type MockConn struct {
	net.Conn
	RemoteAddrFunc func() net.Addr
	CloseFunc      func() error
}

func (m *MockConn) RemoteAddr() net.Addr {
	if m.RemoteAddrFunc != nil {
		return m.RemoteAddrFunc()
	}
	return &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 11112}
}

func (m *MockConn) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}

// MockDIMSEHandler is a mock implementation of DIMSEHandler for testing
type MockDIMSEHandler struct {
	HandleDIMSEMessageFunc func(presContextID byte, msgCtrlHeader byte, data []byte, pduLayer *Layer) error
}

func (m *MockDIMSEHandler) HandleDIMSEMessage(presContextID byte, msgCtrlHeader byte, data []byte, pduLayer *Layer) error {
	if m.HandleDIMSEMessageFunc != nil {
		return m.HandleDIMSEMessageFunc(presContextID, msgCtrlHeader, data, pduLayer)
	}
	return nil
}

// MockAddr implements net.Addr for testing
type MockAddr struct {
	NetworkString string
	AddrString    string
}

func (m *MockAddr) Network() string {
	if m.NetworkString != "" {
		return m.NetworkString
	}
	return "tcp"
}

func (m *MockAddr) String() string {
	if m.AddrString != "" {
		return m.AddrString
	}
	return "127.0.0.1:11112"
}

func TestNewLayer(t *testing.T) {
	mockConn := &MockConn{}
	mockHandler := &MockDIMSEHandler{}
	aeTitle := "TEST_AE"

	layer := NewLayer(mockConn, mockHandler, aeTitle)

	if layer == nil {
		t.Fatal("Expected non-nil layer")
	}

	if layer.conn != mockConn {
		t.Error("Layer conn not set correctly")
	}

	if layer.dimseHandler != mockHandler {
		t.Error("Layer dimseHandler not set correctly")
	}

	if layer.serverAETitle != aeTitle {
		t.Errorf("Layer serverAETitle = %s, want %s", layer.serverAETitle, aeTitle)
	}
}

func TestPDUTypeConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant byte
		expected byte
	}{
		{"Associate-RQ", TypeAssociateRQ, 0x01},
		{"Associate-AC", TypeAssociateAC, 0x02},
		{"Associate-RJ", TypeAssociateRJ, 0x03},
		{"P-DATA-TF", TypePDataTF, 0x04},
		{"Release-RQ", TypeReleaseRQ, 0x05},
		{"Release-RP", TypeReleaseRP, 0x06},
		{"Abort", TypeAbort, 0x07},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.expected {
				t.Errorf("%s = 0x%02x, want 0x%02x", tt.name, tt.constant, tt.expected)
			}
		})
	}
}

func TestPDU_Creation(t *testing.T) {
	tests := []struct {
		name string
		pdu  PDU
	}{
		{
			name: "Associate-RQ PDU",
			pdu: PDU{
				Type:   TypeAssociateRQ,
				Length: 100,
				Data:   make([]byte, 100),
			},
		},
		{
			name: "P-DATA-TF PDU",
			pdu: PDU{
				Type:   TypePDataTF,
				Length: 1024,
				Data:   make([]byte, 1024),
			},
		},
		{
			name: "Release-RQ PDU",
			pdu: PDU{
				Type:   TypeReleaseRQ,
				Length: 4,
				Data:   make([]byte, 4),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.pdu.Type == 0 {
				t.Error("PDU Type should be set")
			}
			if tt.pdu.Length == 0 {
				t.Error("PDU Length should be set")
			}
			if len(tt.pdu.Data) != int(tt.pdu.Length) {
				t.Errorf("PDU Data length = %d, want %d", len(tt.pdu.Data), tt.pdu.Length)
			}
		})
	}
}

func TestAssociationContext_Creation(t *testing.T) {
	ctx := &AssociationContext{
		CalledAETitle:    "CALLED_AE",
		CallingAETitle:   "CALLING_AE",
		MaxPDULength:     16384,
		PresentationCtxs: make(map[byte]*PresentationContext),
	}

	if ctx.CalledAETitle != "CALLED_AE" {
		t.Errorf("CalledAETitle = %s, want CALLED_AE", ctx.CalledAETitle)
	}
	if ctx.CallingAETitle != "CALLING_AE" {
		t.Errorf("CallingAETitle = %s, want CALLING_AE", ctx.CallingAETitle)
	}
	if ctx.MaxPDULength != 16384 {
		t.Errorf("MaxPDULength = %d, want 16384", ctx.MaxPDULength)
	}
	if ctx.PresentationCtxs == nil {
		t.Error("PresentationCtxs should be initialized")
	}
}

func TestPresentationContext_Creation(t *testing.T) {
	tests := []struct {
		name string
		ctx  PresentationContext
	}{
		{
			name: "C-FIND context",
			ctx: PresentationContext{
				ID:             1,
				Result:         0, // Accepted
				AbstractSyntax: "1.2.840.10008.5.1.4.1.2.1.1",
				TransferSyntax: "1.2.840.10008.1.2",
			},
		},
		{
			name: "C-ECHO context",
			ctx: PresentationContext{
				ID:             3,
				Result:         0,
				AbstractSyntax: "1.2.840.10008.1.1",
				TransferSyntax: "1.2.840.10008.1.2",
			},
		},
		{
			name: "Rejected context",
			ctx: PresentationContext{
				ID:             5,
				Result:         3, // Transfer syntaxes not supported
				AbstractSyntax: "1.2.840.10008.5.1.4.1.1.1",
				TransferSyntax: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.ctx.ID == 0 {
				t.Error("Presentation context ID should be non-zero")
			}
			if tt.ctx.AbstractSyntax == "" && tt.ctx.Result == 0 {
				t.Error("Accepted context should have AbstractSyntax")
			}
		})
	}
}

func TestAssociationContext_AddPresentationContext(t *testing.T) {
	ctx := &AssociationContext{
		CalledAETitle:    "SERVER",
		CallingAETitle:   "CLIENT",
		MaxPDULength:     16384,
		PresentationCtxs: make(map[byte]*PresentationContext),
	}

	// Add a presentation context
	presCtx := &PresentationContext{
		ID:             1,
		Result:         0,
		AbstractSyntax: "1.2.840.10008.1.1",
		TransferSyntax: "1.2.840.10008.1.2",
	}
	ctx.PresentationCtxs[presCtx.ID] = presCtx

	// Verify it was added
	retrieved, exists := ctx.PresentationCtxs[1]
	if !exists {
		t.Error("Presentation context not found")
	}
	if retrieved.AbstractSyntax != "1.2.840.10008.1.1" {
		t.Errorf("AbstractSyntax = %s, want 1.2.840.10008.1.1", retrieved.AbstractSyntax)
	}

	// Add another context
	presCtx2 := &PresentationContext{
		ID:             3,
		Result:         0,
		AbstractSyntax: "1.2.840.10008.5.1.4.1.2.1.1",
		TransferSyntax: "1.2.840.10008.1.2",
	}
	ctx.PresentationCtxs[presCtx2.ID] = presCtx2

	if len(ctx.PresentationCtxs) != 2 {
		t.Errorf("Expected 2 presentation contexts, got %d", len(ctx.PresentationCtxs))
	}
}

func TestMockAddr(t *testing.T) {
	addr := &MockAddr{
		NetworkString: "tcp",
		AddrString:    "192.168.1.1:104",
	}

	if addr.Network() != "tcp" {
		t.Errorf("Network() = %s, want tcp", addr.Network())
	}
	if addr.String() != "192.168.1.1:104" {
		t.Errorf("String() = %s, want 192.168.1.1:104", addr.String())
	}
}

func TestMockConn_RemoteAddr(t *testing.T) {
	customAddr := &MockAddr{
		NetworkString: "tcp",
		AddrString:    "10.0.0.1:11112",
	}

	mockConn := &MockConn{
		RemoteAddrFunc: func() net.Addr {
			return customAddr
		},
	}

	addr := mockConn.RemoteAddr()
	if addr.String() != "10.0.0.1:11112" {
		t.Errorf("RemoteAddr().String() = %s, want 10.0.0.1:11112", addr.String())
	}
}
