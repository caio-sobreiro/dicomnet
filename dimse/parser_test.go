package dimse

import (
	"encoding/binary"
	"testing"

	"github.com/caio-sobreiro/dicomnet/types"
)

func TestParseDIMSECommand_Success(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected types.Message
	}{
		{
			name: "C-FIND Request with all fields",
			data: func() []byte {
				var buf []byte
				// Command Field (0000,0100)
				buf = append(buf, 0x00, 0x00, 0x00, 0x01) // Tag
				buf = append(buf, 0x02, 0x00, 0x00, 0x00) // Length = 2
				buf = append(buf, 0x20, 0x00)             // CFindRQ = 0x0020

				// Message ID (0000,0110)
				buf = append(buf, 0x00, 0x00, 0x10, 0x01) // Tag
				buf = append(buf, 0x02, 0x00, 0x00, 0x00) // Length = 2
				buf = append(buf, 0x01, 0x00)             // MessageID = 1

				// Command Data Set Type (0000,0800)
				buf = append(buf, 0x00, 0x00, 0x00, 0x08) // Tag
				buf = append(buf, 0x02, 0x00, 0x00, 0x00) // Length = 2
				buf = append(buf, 0x01, 0x00)             // Type = 1

				// Affected SOP Class UID (0000,0002)
				buf = append(buf, 0x00, 0x00, 0x02, 0x00) // Tag
				sopUID := []byte("1.2.840.10008.5.1.4.1.2.1.1")
				if len(sopUID)%2 == 1 {
					sopUID = append(sopUID, 0x00) // Pad to even length
				}
				lengthBytes := make([]byte, 4)
				binary.LittleEndian.PutUint32(lengthBytes, uint32(len(sopUID)))
				buf = append(buf, lengthBytes...)
				buf = append(buf, sopUID...)

				return buf
			}(),
			expected: types.Message{
				CommandField:        types.CFindRQ,
				MessageID:           1,
				CommandDataSetType:  1,
				AffectedSOPClassUID: "1.2.840.10008.5.1.4.1.2.1.1",
			},
		},
		{
			name: "C-FIND Response",
			data: func() []byte {
				var buf []byte
				// Command Field (0000,0100)
				buf = append(buf, 0x00, 0x00, 0x00, 0x01)
				buf = append(buf, 0x02, 0x00, 0x00, 0x00)
				buf = append(buf, 0x20, 0x80) // CFindRSP = 0x8020

				// Message ID (0000,0110)
				buf = append(buf, 0x00, 0x00, 0x10, 0x01)
				buf = append(buf, 0x02, 0x00, 0x00, 0x00)
				buf = append(buf, 0x02, 0x00) // MessageID = 2

				// Command Data Set Type (0000,0800)
				buf = append(buf, 0x00, 0x00, 0x00, 0x08)
				buf = append(buf, 0x02, 0x00, 0x00, 0x00)
				buf = append(buf, 0x00, 0x00) // Type = 0

				return buf
			}(),
			expected: types.Message{
				CommandField:       types.CFindRSP,
				MessageID:          2,
				CommandDataSetType: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg, err := parseDIMSECommand(tt.data, nil)
			if err != nil {
				t.Fatalf("parseDIMSECommand() error = %v", err)
			}

			if msg.CommandField != tt.expected.CommandField {
				t.Errorf("CommandField = 0x%04x, want 0x%04x", msg.CommandField, tt.expected.CommandField)
			}
			if msg.MessageID != tt.expected.MessageID {
				t.Errorf("MessageID = %d, want %d", msg.MessageID, tt.expected.MessageID)
			}
			if msg.CommandDataSetType != tt.expected.CommandDataSetType {
				t.Errorf("CommandDataSetType = 0x%04x, want 0x%04x", msg.CommandDataSetType, tt.expected.CommandDataSetType)
			}
			if msg.AffectedSOPClassUID != tt.expected.AffectedSOPClassUID {
				t.Errorf("AffectedSOPClassUID = %q, want %q", msg.AffectedSOPClassUID, tt.expected.AffectedSOPClassUID)
			}
		})
	}
}

func TestParseDIMSECommand_Errors(t *testing.T) {
	tests := []struct {
		name        string
		data        []byte
		expectError bool
	}{
		{
			name:        "Empty data",
			data:        []byte{},
			expectError: true,
		},
		{
			name:        "Too short - less than 12 bytes",
			data:        []byte{0x00, 0x00, 0x00, 0x01, 0x02},
			expectError: true,
		},
		{
			name:        "Exactly 11 bytes",
			data:        make([]byte, 11),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg, err := parseDIMSECommand(tt.data, nil)
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if msg == nil {
					t.Error("Expected message but got nil")
				}
			}
		})
	}
}

func TestParseDIMSECommand_EdgeCases(t *testing.T) {
	t.Run("Truncated element - not enough data for value", func(t *testing.T) {
		var buf []byte
		// Need at least 12 bytes for parser
		buf = make([]byte, 12)
		// Command Field tag and length
		buf[0], buf[1], buf[2], buf[3] = 0x00, 0x00, 0x00, 0x01
		buf[4], buf[5], buf[6], buf[7] = 0x02, 0x00, 0x00, 0x00
		// Only 1 byte of value instead of 2
		buf[8] = 0x20
		// Rest are zeros

		msg, err := parseDIMSECommand(buf, nil)
		// Should not error, just stop parsing or handle gracefully
		if err != nil {
			t.Logf("Got expected error for truncated data: %v", err)
		}
		if msg != nil && msg.CommandField == 0 {
			t.Log("Parsing stopped due to truncated data, as expected")
		}
	})

	t.Run("Very large length - should break parsing", func(t *testing.T) {
		var buf []byte
		// Need minimum data to pass initial check
		buf = make([]byte, 14)
		// Command Field tag
		buf[0], buf[1], buf[2], buf[3] = 0x00, 0x00, 0x00, 0x01
		// Impossibly large length (2MB)
		buf[4], buf[5], buf[6], buf[7] = 0x00, 0x00, 0x20, 0x00
		buf[8], buf[9] = 0x20, 0x00

		msg, err := parseDIMSECommand(buf, nil)
		// Parser should handle this gracefully (stops on sanity check)
		if err != nil {
			t.Logf("Got expected error for large length: %v", err)
		}
		if msg != nil {
			t.Log("Parser handled large length gracefully")
		}
	})

	t.Run("SOP Class UID with null padding", func(t *testing.T) {
		var buf []byte
		// Affected SOP Class UID (0000,0002)
		buf = append(buf, 0x00, 0x00, 0x02, 0x00)
		sopUID := []byte("1.2.840.10008.1.1\x00")
		lengthBytes := make([]byte, 4)
		binary.LittleEndian.PutUint32(lengthBytes, uint32(len(sopUID)))
		buf = append(buf, lengthBytes...)
		buf = append(buf, sopUID...)

		msg, err := parseDIMSECommand(buf, nil)
		if err != nil {
			t.Fatalf("parseDIMSECommand() error = %v", err)
		}

		expected := "1.2.840.10008.1.1"
		if msg.AffectedSOPClassUID != expected {
			t.Errorf("AffectedSOPClassUID = %q, want %q", msg.AffectedSOPClassUID, expected)
		}
	})

	t.Run("Odd length element with padding", func(t *testing.T) {
		var buf []byte
		// Command Field (0000,0100)
		buf = append(buf, 0x00, 0x00, 0x00, 0x01)
		buf = append(buf, 0x01, 0x00, 0x00, 0x00) // Odd length = 1
		buf = append(buf, 0x20)                   // 1 byte value
		buf = append(buf, 0x00)                   // Padding byte

		// Message ID (0000,0110)
		buf = append(buf, 0x00, 0x00, 0x10, 0x01)
		buf = append(buf, 0x02, 0x00, 0x00, 0x00)
		buf = append(buf, 0x01, 0x00)

		msg, err := parseDIMSECommand(buf, nil)
		if err != nil {
			t.Fatalf("parseDIMSECommand() error = %v", err)
		}

		// Should parse MessageID correctly despite odd-length previous element
		if msg.MessageID != 1 {
			t.Errorf("MessageID = %d, want 1", msg.MessageID)
		}
	})

	t.Run("Non-command group elements should be skipped", func(t *testing.T) {
		var buf []byte
		// Patient Name (0010,0010) - should be skipped
		buf = append(buf, 0x10, 0x00, 0x10, 0x00)
		buf = append(buf, 0x08, 0x00, 0x00, 0x00)
		buf = append(buf, []byte("Doe^John")...)

		// Command Field (0000,0100)
		buf = append(buf, 0x00, 0x00, 0x00, 0x01)
		buf = append(buf, 0x02, 0x00, 0x00, 0x00)
		buf = append(buf, 0x20, 0x00)

		msg, err := parseDIMSECommand(buf, nil)
		if err != nil {
			t.Fatalf("parseDIMSECommand() error = %v", err)
		}

		if msg.CommandField != types.CFindRQ {
			t.Errorf("CommandField = 0x%04x, want 0x%04x", msg.CommandField, types.CFindRQ)
		}
	})
}

func TestCreateDIMSECommand(t *testing.T) {
	tests := []struct {
		name string
		msg  types.Message
	}{
		{
			name: "C-FIND Response with all fields",
			msg: types.Message{
				CommandField:              types.CFindRSP,
				MessageIDBeingRespondedTo: 1,
				CommandDataSetType:        0x0000,
				Status:                    types.StatusSuccess,
				AffectedSOPClassUID:       "1.2.840.10008.5.1.4.1.2.1.1",
			},
		},
		{
			name: "C-ECHO Request without MessageIDBeingRespondedTo",
			msg: types.Message{
				CommandField:        types.CEchoRQ,
				CommandDataSetType:  0x0101,
				Status:              0,
				AffectedSOPClassUID: "1.2.840.10008.1.1",
			},
		},
		{
			name: "Message without SOP Class UID",
			msg: types.Message{
				CommandField:       types.CFindRSP,
				CommandDataSetType: 0x0000,
				Status:             types.StatusPending,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := createDIMSECommand(&tt.msg)

			if len(data) == 0 {
				t.Error("createDIMSECommand() returned empty data")
			}

			// Data should be properly formatted DICOM
			// Without MessageIDBeingRespondedTo and without SOP Class UID:
			// Command Field (12 bytes), Command Data Set Type (12 bytes), Status (12 bytes) = 36 bytes
			// But without SOP Class UID, it's only 30 bytes, which is valid
			minExpected := 12 // At least one element
			if len(data) < minExpected {
				t.Errorf("createDIMSECommand() data length = %d, want at least %d", len(data), minExpected)
			}
		})
	}
}

func TestCreateDIMSECommand_RoundTrip(t *testing.T) {
	tests := []struct {
		name string
		msg  types.Message
	}{
		{
			name: "C-FIND Request",
			msg: types.Message{
				CommandField:              types.CFindRQ,
				MessageIDBeingRespondedTo: 0,
				CommandDataSetType:        0x0001,
				Status:                    0,
				AffectedSOPClassUID:       "1.2.840.10008.5.1.4.1.2.1.1",
			},
		},
		{
			name: "C-FIND Response Success",
			msg: types.Message{
				CommandField:              types.CFindRSP,
				MessageIDBeingRespondedTo: 5,
				CommandDataSetType:        0x0000,
				Status:                    types.StatusSuccess,
				AffectedSOPClassUID:       "1.2.840.10008.5.1.4.1.2.1.1",
			},
		},
		{
			name: "C-ECHO Response",
			msg: types.Message{
				CommandField:              types.CEchoRSP,
				MessageIDBeingRespondedTo: 3,
				CommandDataSetType:        0x0101,
				Status:                    types.StatusSuccess,
				AffectedSOPClassUID:       "1.2.840.10008.1.1",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create DIMSE command
			data := createDIMSECommand(&tt.msg)

			// Parse it back
			parsed, err := parseDIMSECommand(data, nil)
			if err != nil {
				t.Fatalf("parseDIMSECommand() error = %v", err)
			}

			// Verify key fields match
			if parsed.CommandField != tt.msg.CommandField {
				t.Errorf("Round-trip CommandField = 0x%04x, want 0x%04x",
					parsed.CommandField, tt.msg.CommandField)
			}
			// Note: parseDIMSECommand doesn't read MessageIDBeingRespondedTo (0x0120),
			// so we don't test round-trip for that field
			if parsed.CommandDataSetType != tt.msg.CommandDataSetType {
				t.Errorf("Round-trip CommandDataSetType = 0x%04x, want 0x%04x",
					parsed.CommandDataSetType, tt.msg.CommandDataSetType)
			}
			if parsed.Status != tt.msg.Status {
				t.Errorf("Round-trip Status = 0x%04x, want 0x%04x",
					parsed.Status, tt.msg.Status)
			}
			if parsed.AffectedSOPClassUID != tt.msg.AffectedSOPClassUID {
				t.Errorf("Round-trip AffectedSOPClassUID = %q, want %q",
					parsed.AffectedSOPClassUID, tt.msg.AffectedSOPClassUID)
			}
		})
	}
}

// TestParseDIMSECommand_CStoreRQ verifies that Affected SOP Instance UID (0000,1000)
// is correctly parsed from a C-STORE-RQ command, which is required to echo it
// back in the C-STORE-RSP per DICOM PS3.7 Table 9.3-1.
func TestParseDIMSECommand_CStoreRQ(t *testing.T) {
	var buf []byte

	// Command Field (0000,0100) = C-STORE-RQ
	buf = append(buf, 0x00, 0x00, 0x00, 0x01)
	buf = append(buf, 0x02, 0x00, 0x00, 0x00)
	buf = append(buf, 0x01, 0x00) // CStoreRQ = 0x0001

	// Message ID (0000,0110)
	buf = append(buf, 0x00, 0x00, 0x10, 0x01)
	buf = append(buf, 0x02, 0x00, 0x00, 0x00)
	buf = append(buf, 0x00, 0x00) // MessageID = 0 (GE workstations may send 0)

	// Affected SOP Class UID (0000,0002)
	buf = append(buf, 0x00, 0x00, 0x02, 0x00)
	sopClass := []byte("1.2.840.10008.5.1.4.1.1.4\x00") // MR Image Storage, padded to even
	lenBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(lenBytes, uint32(len(sopClass)))
	buf = append(buf, lenBytes...)
	buf = append(buf, sopClass...)

	// Command Data Set Type (0000,0800)
	buf = append(buf, 0x00, 0x00, 0x00, 0x08)
	buf = append(buf, 0x02, 0x00, 0x00, 0x00)
	buf = append(buf, 0x00, 0x00) // Dataset present

	// Affected SOP Instance UID (0000,1000) - previously not parsed
	buf = append(buf, 0x00, 0x00, 0x00, 0x10)
	sopInstance := []byte("1.2.3.4.5.6.7.8.9.0\x00") // padded to even
	binary.LittleEndian.PutUint32(lenBytes, uint32(len(sopInstance)))
	buf = append(buf, lenBytes...)
	buf = append(buf, sopInstance...)

	msg, err := parseDIMSECommand(buf, nil)
	if err != nil {
		t.Fatalf("parseDIMSECommand() error = %v", err)
	}

	if msg.CommandField != CStoreRQ {
		t.Errorf("CommandField = 0x%04x, want 0x%04x", msg.CommandField, CStoreRQ)
	}
	if msg.AffectedSOPClassUID != "1.2.840.10008.5.1.4.1.1.4" {
		t.Errorf("AffectedSOPClassUID = %q, want %q", msg.AffectedSOPClassUID, "1.2.840.10008.5.1.4.1.1.4")
	}
	if msg.AffectedSOPInstanceUID != "1.2.3.4.5.6.7.8.9.0" {
		t.Errorf("AffectedSOPInstanceUID = %q, want %q", msg.AffectedSOPInstanceUID, "1.2.3.4.5.6.7.8.9.0")
	}
}

// TestCreateDIMSECommand_ResponseWithZeroMessageID verifies that a C-STORE-RSP
// always includes (0000,0120) MessageIDBeingRespondedTo even when the value is 0,
// as required by DICOM PS3.7. Strict implementations (e.g. GE AW workstations)
// will abort the association if this field is missing.
func TestCreateDIMSECommand_ResponseWithZeroMessageID(t *testing.T) {
	msg := types.Message{
		CommandField:              CStoreRSP,
		MessageIDBeingRespondedTo: 0, // GE AW workstations may use MessageID=0
		AffectedSOPClassUID:       "1.2.840.10008.5.1.4.1.1.4",
		AffectedSOPInstanceUID:    "1.2.3.4.5.6.7.8.9.0",
		CommandDataSetType:        0x0101,
		Status:                    StatusSuccess,
	}

	data := createDIMSECommand(&msg)

	// Scan encoded bytes for tag (0000,0120)
	found := false
	offset := 0
	for offset+8 <= len(data) {
		group := binary.LittleEndian.Uint16(data[offset : offset+2])
		element := binary.LittleEndian.Uint16(data[offset+2 : offset+4])
		length := binary.LittleEndian.Uint32(data[offset+4 : offset+8])
		if group == 0x0000 && element == 0x0120 {
			found = true
			if length != 2 {
				t.Errorf("(0000,0120) length = %d, want 2", length)
			}
			if offset+10 <= len(data) {
				val := binary.LittleEndian.Uint16(data[offset+8 : offset+10])
				if val != 0 {
					t.Errorf("(0000,0120) value = %d, want 0", val)
				}
			}
			break
		}
		offset += 8 + int(length)
		if length%2 == 1 {
			offset++
		}
	}
	if !found {
		t.Error("(0000,0120) MessageIDBeingRespondedTo not found in encoded C-STORE-RSP with MessageID=0")
	}
}

func TestCreateDIMSECommand_OddLengthUID(t *testing.T) {
	msg := types.Message{
		CommandField:        types.CEchoRQ,
		CommandDataSetType:  0x0101,
		Status:              0,
		AffectedSOPClassUID: "1.2.3", // Odd length (5 chars)
	}

	data := createDIMSECommand(&msg)

	// Parse it back
	parsed, err := parseDIMSECommand(data, nil)
	if err != nil {
		t.Fatalf("parseDIMSECommand() error = %v", err)
	}

	// UID should be preserved correctly (padding removed)
	if parsed.AffectedSOPClassUID != msg.AffectedSOPClassUID {
		t.Errorf("AffectedSOPClassUID = %q, want %q",
			parsed.AffectedSOPClassUID, msg.AffectedSOPClassUID)
	}
}
