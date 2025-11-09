package client

import (
	"bytes"
	"encoding/binary"
	"io"
	"log/slog"
	"net"
	"testing"
	"time"

	"github.com/caio-sobreiro/dicomnet/dimse"
	"github.com/caio-sobreiro/dicomnet/pdu"
	"github.com/caio-sobreiro/dicomnet/types"
)

// mockConn implements net.Conn for testing
type mockConn struct {
	readBuf  *bytes.Buffer
	writeBuf *bytes.Buffer
	closed   bool
}

func newMockConn() *mockConn {
	return &mockConn{
		readBuf:  new(bytes.Buffer),
		writeBuf: new(bytes.Buffer),
	}
}

func (m *mockConn) Read(b []byte) (n int, err error) {
	if m.closed {
		return 0, io.EOF
	}
	return m.readBuf.Read(b)
}

func (m *mockConn) Write(b []byte) (n int, err error) {
	if m.closed {
		return 0, io.ErrClosedPipe
	}
	return m.writeBuf.Write(b)
}

func (m *mockConn) Close() error {
	m.closed = true
	return nil
}

func (m *mockConn) LocalAddr() net.Addr                { return nil }
func (m *mockConn) RemoteAddr() net.Addr               { return nil }
func (m *mockConn) SetDeadline(t time.Time) error      { return nil }
func (m *mockConn) SetReadDeadline(t time.Time) error  { return nil }
func (m *mockConn) SetWriteDeadline(t time.Time) error { return nil }

// Test encoding DIMSE commands in Implicit VR Little Endian
func TestEncodeCommand(t *testing.T) {
	msg := &types.Message{
		CommandField:           dimse.CStoreRQ,
		MessageID:              1,
		Priority:               0,
		CommandDataSetType:     0x0000, // Dataset present
		AffectedSOPClassUID:    "1.2.840.10008.5.1.4.1.1.2",
		AffectedSOPInstanceUID: "1.2.3.4.5",
	}

	encoded, err := dimse.EncodeCommand(msg)
	if err != nil {
		t.Fatalf("dimse.EncodeCommand failed: %v", err)
	}

	if len(encoded) == 0 {
		t.Fatal("Encoded command is empty")
	}

	// Verify encoding is in Implicit VR Little Endian format
	// First element should be Command Group Length (0000,0000)
	if len(encoded) < 8 {
		t.Fatal("Encoded command too short")
	}

	group := binary.LittleEndian.Uint16(encoded[0:2])
	element := binary.LittleEndian.Uint16(encoded[2:4])

	if group != 0x0000 || element != 0x0000 {
		t.Errorf("First element = (%04X,%04X), want (0000,0000)", group, element)
	}

	// Verify Command Field (0000,0100) is present
	found := false
	offset := 0
	for offset+8 <= len(encoded) {
		g := binary.LittleEndian.Uint16(encoded[offset : offset+2])
		e := binary.LittleEndian.Uint16(encoded[offset+2 : offset+4])
		length := binary.LittleEndian.Uint32(encoded[offset+4 : offset+8])

		if g == 0x0000 && e == 0x0100 {
			found = true
			if length != 2 {
				t.Errorf("Command Field length = %d, want 2", length)
			}
			if offset+8+int(length) <= len(encoded) {
				cmdValue := binary.LittleEndian.Uint16(encoded[offset+8 : offset+8+int(length)])
				if cmdValue != dimse.CStoreRQ {
					t.Errorf("Command Field value = 0x%04X, want 0x%04X (C-STORE-RQ)", cmdValue, dimse.CStoreRQ)
				}
			}
			break
		}

		offset += 8 + int(length)
	}

	if !found {
		t.Error("Command Field (0000,0100) not found in encoded command")
	}
}

// Test multiple transfer syntaxes in presentation context (prefers uncompressed)
func TestTransferSyntaxNegotiation(t *testing.T) {
	// Build A-ASSOCIATE-RQ with multiple transfer syntaxes
	buf := make([]byte, 0, 1024)
	buf = append(buf, 0x00, 0x01) // Protocol version
	buf = append(buf, 0x00, 0x00) // Reserved

	// Called AE Title (16 bytes, space padded)
	calledAE := "TEST_SCP        "
	buf = append(buf, []byte(calledAE[:16])...)

	// Calling AE Title (16 bytes, space padded)
	callingAE := "TEST_SCU        "
	buf = append(buf, []byte(callingAE[:16])...)

	// Reserved
	buf = append(buf, make([]byte, 32)...)

	// Application Context
	buf = append(buf, 0x10, 0x00) // Item type
	appContext := types.ApplicationContextUID
	buf = binary.BigEndian.AppendUint16(buf, uint16(len(appContext)))
	buf = append(buf, []byte(appContext)...)

	// Presentation Context with multiple transfer syntaxes
	pcBuf := make([]byte, 0)
	pcBuf = append(pcBuf, 0x01)             // Presentation Context ID
	pcBuf = append(pcBuf, 0x00, 0x00, 0x00) // Reserved

	// Abstract Syntax
	pcBuf = append(pcBuf, 0x30, 0x00) // Item type
	sopClass := "1.2.840.10008.5.1.4.1.1.2"
	pcBuf = binary.BigEndian.AppendUint16(pcBuf, uint16(len(sopClass)))
	pcBuf = append(pcBuf, []byte(sopClass)...)

	// Transfer Syntax 1: Explicit VR Little Endian
	pcBuf = append(pcBuf, 0x40, 0x00) // Item type
	ts1 := "1.2.840.10008.1.2.1"
	pcBuf = binary.BigEndian.AppendUint16(pcBuf, uint16(len(ts1)))
	pcBuf = append(pcBuf, []byte(ts1)...)

	// Transfer Syntax 2: Implicit VR Little Endian
	pcBuf = append(pcBuf, 0x40, 0x00) // Item type
	ts2 := "1.2.840.10008.1.2"
	pcBuf = binary.BigEndian.AppendUint16(pcBuf, uint16(len(ts2)))
	pcBuf = append(pcBuf, []byte(ts2)...)

	// Add presentation context to buffer
	buf = append(buf, 0x20, 0x00) // Item type
	buf = binary.BigEndian.AppendUint16(buf, uint16(len(pcBuf)))
	buf = append(buf, pcBuf...)

	// Verify both transfer syntaxes are present
	bufStr := string(buf)
	transferSyntaxes := []string{
		"1.2.840.10008.1.2.1", // Explicit VR
		"1.2.840.10008.1.2",   // Implicit VR
	}

	for _, ts := range transferSyntaxes {
		if !bytes.Contains([]byte(bufStr), []byte(ts)) {
			t.Errorf("Transfer syntax %s not found in A-ASSOCIATE-RQ", ts)
		}
	}
}

// Test PDU fragmentation for large datasets
func TestSendPDataTF_Fragmentation(t *testing.T) {
	conn := newMockConn()
	assoc := &Association{
		conn:             conn,
		callingAETitle:   "TEST_SCU",
		calledAETitle:    "TEST_SCP",
		maxPDULength:     200, // Small size to force fragmentation
		presentationCtxs: make(map[byte]*PresentationContext),
		logger:           slog.Default(),
	}

	// Create data larger than maxPDULength
	data := make([]byte, 500)
	for i := range data {
		data[i] = byte(i % 256)
	}

	err := dimse.SendPDataTF(assoc.conn, 1, assoc.maxPDULength, data, false, true)
	if err != nil {
		t.Fatalf("dimse.SendPDataTF failed: %v", err)
	}

	written := conn.writeBuf.Bytes()

	// Count PDUs
	pduCount := 0
	offset := 0

	for offset < len(written) {
		if offset+6 > len(written) {
			break
		}

		pduType := written[offset]
		if pduType != pdu.TypePDataTF {
			t.Errorf("PDU type = 0x%02X, want 0x%02X (P-DATA-TF)", pduType, pdu.TypePDataTF)
		}

		pduLength := binary.BigEndian.Uint32(written[offset+2 : offset+6])
		pduCount++
		offset += 6 + int(pduLength)
	}

	if pduCount < 2 {
		t.Errorf("Expected multiple PDUs due to fragmentation, got %d", pduCount)
	}
}

// Test sending a complete DIMSE message (command + dataset)
func TestSendDIMSEMessage(t *testing.T) {
	conn := newMockConn()
	assoc := &Association{
		conn:             conn,
		callingAETitle:   "TEST_SCU",
		calledAETitle:    "TEST_SCP",
		maxPDULength:     16384,
		presentationCtxs: make(map[byte]*PresentationContext),
		logger:           slog.Default(),
	}

	commandData := []byte{0x01, 0x02, 0x03, 0x04}
	datasetData := []byte{0x05, 0x06, 0x07, 0x08, 0x09, 0x0A}

	err := dimse.SendDIMSEMessage(assoc.conn, 1, assoc.maxPDULength, commandData, datasetData)
	if err != nil {
		t.Fatalf("dimse.SendDIMSEMessage failed: %v", err)
	}

	written := conn.writeBuf.Bytes()
	if len(written) == 0 {
		t.Fatal("No data written")
	}

	// Verify at least 2 PDUs were sent (command + dataset)
	pduCount := 0
	offset := 0

	for offset < len(written) {
		if offset+6 > len(written) {
			break
		}

		pduType := written[offset]
		if pduType != pdu.TypePDataTF {
			t.Errorf("PDU type = 0x%02X, want 0x%02X", pduType, pdu.TypePDataTF)
		}

		pduLength := binary.BigEndian.Uint32(written[offset+2 : offset+6])
		pduCount++
		offset += 6 + int(pduLength)
	}

	if pduCount < 2 {
		t.Errorf("Expected at least 2 PDUs (command + dataset), got %d", pduCount)
	}
}

// Test A-ABORT handling
func TestReceiveCStoreResponse_Abort(t *testing.T) {
	conn := newMockConn()
	assoc := &Association{
		conn:             conn,
		callingAETitle:   "TEST_SCU",
		calledAETitle:    "TEST_SCP",
		maxPDULength:     16384,
		presentationCtxs: make(map[byte]*PresentationContext),
		logger:           slog.Default(),
	}

	// Build A-ABORT PDU
	var abortPDU bytes.Buffer
	abortPDU.WriteByte(0x07) // A-ABORT PDU type
	abortPDU.WriteByte(0x00) // Reserved

	// Length = 4 (fixed for A-ABORT)
	abortPDU.WriteByte(0x00)
	abortPDU.WriteByte(0x00)
	abortPDU.WriteByte(0x00)
	abortPDU.WriteByte(0x04)

	// Abort parameters
	abortPDU.WriteByte(0x00) // Reserved
	abortPDU.WriteByte(0x00) // Reserved
	abortPDU.WriteByte(0x02) // Source: service-provider
	abortPDU.WriteByte(0x01) // Reason: unrecognized PDU

	conn.readBuf.Write(abortPDU.Bytes())

	// Receive response - should return error
	_, _, err := dimse.ReceiveDIMSEMessage(assoc.conn)

	if err == nil {
		t.Fatal("Expected error for A-ABORT, got nil")
	}

	if !bytes.Contains([]byte(err.Error()), []byte("A-ABORT")) {
		t.Errorf("Error message should mention A-ABORT, got: %v", err)
	}
}

// Test Implicit VR element appending
func TestAppendImplicitElement(t *testing.T) {
	tests := []struct {
		name    string
		group   uint16
		element uint16
		value   []byte
	}{
		{
			name:    "Command Field",
			group:   0x0000,
			element: 0x0100,
			value:   []byte{0x01, 0x00},
		},
		{
			name:    "Status",
			group:   0x0000,
			element: 0x0900,
			value:   []byte{0x00, 0x00},
		},
		{
			name:    "Patient Name",
			group:   0x0010,
			element: 0x0010,
			value:   []byte("DOE^JOHN"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := make([]byte, 0)
			result := dimse.AppendImplicitElement(buf, tt.group, tt.element, tt.value)

			if len(result) < 8 {
				t.Fatalf("Result too short: %d bytes", len(result))
			}

			// Verify tag (group and element) in little endian
			group := binary.LittleEndian.Uint16(result[0:2])
			element := binary.LittleEndian.Uint16(result[2:4])

			if group != tt.group {
				t.Errorf("Group = 0x%04X, want 0x%04X", group, tt.group)
			}

			if element != tt.element {
				t.Errorf("Element = 0x%04X, want 0x%04X", element, tt.element)
			}

			// Verify length in little endian
			length := binary.LittleEndian.Uint32(result[4:8])
			if int(length) != len(tt.value) {
				t.Errorf("Length = %d, want %d", length, len(tt.value))
			}

			// Verify value
			if len(result) >= 8+len(tt.value) {
				value := result[8 : 8+len(tt.value)]
				if !bytes.Equal(value, tt.value) {
					t.Errorf("Value = %v, want %v", value, tt.value)
				}
			}
		})
	}
}

// Test AE title padding
// Test basic CStoreRequest structure
func TestCStoreRequest(t *testing.T) {
	req := &CStoreRequest{
		SOPClassUID:    "1.2.840.10008.5.1.4.1.1.2",
		SOPInstanceUID: "1.2.3.4.5",
		Data:           []byte("test data"),
		MessageID:      1,
	}

	if req.SOPClassUID != "1.2.840.10008.5.1.4.1.1.2" {
		t.Errorf("SOPClassUID = %s, want 1.2.840.10008.5.1.4.1.1.2", req.SOPClassUID)
	}

	if len(req.Data) != 9 {
		t.Errorf("Data length = %d, want 9", len(req.Data))
	}
}

// Test Close closes the connection
func TestClose(t *testing.T) {
	conn := newMockConn()
	assoc := &Association{
		conn:             conn,
		callingAETitle:   "TEST_SCU",
		calledAETitle:    "TEST_SCP",
		maxPDULength:     16384,
		presentationCtxs: make(map[byte]*PresentationContext),
		logger:           slog.Default(),
	}

	err := assoc.Close()
	if err != nil {
		t.Errorf("Close() error = %v, want nil", err)
	}

	if !conn.closed {
		t.Error("Connection not closed")
	}
}
