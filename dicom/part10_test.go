package dicom

import (
	"bytes"
	"encoding/binary"
	"testing"
)

// createValidPart10File creates a minimal valid DICOM Part 10 file for testing
func createValidPart10File() []byte {
	var data []byte

	// 128-byte preamble (all zeros)
	preamble := make([]byte, 128)
	data = append(data, preamble...)

	// DICM prefix
	data = append(data, []byte("DICM")...)

	// Transfer Syntax UID (0002,0010) - using short VR format
	data = append(data, 0x02, 0x00, 0x10, 0x00) // Tag
	data = append(data, 'U', 'I')                // VR
	tsUID := "1.2.840.10008.1.2.1\x00"          // Explicit VR Little Endian (padded)
	tsLength := make([]byte, 2)
	binary.LittleEndian.PutUint16(tsLength, uint16(len(tsUID)))
	data = append(data, tsLength...)
	data = append(data, []byte(tsUID)...)

	// Dataset starts here (group > 0x0002)
	// Patient Name (0010,0010)
	data = append(data, 0x10, 0x00, 0x10, 0x00) // Tag
	data = append(data, 'P', 'N')                // VR
	patientName := "TEST^PATIENT"
	nameLength := make([]byte, 2)
	binary.LittleEndian.PutUint16(nameLength, uint16(len(patientName)))
	data = append(data, nameLength...)
	data = append(data, []byte(patientName)...)

	return data
}

func TestStripPart10Header_ValidFile(t *testing.T) {
	data := createValidPart10File()

	dataset, err := StripPart10Header(data)
	if err != nil {
		t.Fatalf("StripPart10Header() error = %v", err)
	}

	// Dataset should start with Patient Name tag (0010,0010)
	if len(dataset) < 4 {
		t.Fatal("Dataset too short")
	}

	expectedTag := []byte{0x10, 0x00, 0x10, 0x00}
	if !bytes.Equal(dataset[0:4], expectedTag) {
		t.Errorf("Expected dataset to start with tag 0010,0010, got %02x%02x,%02x%02x",
			dataset[1], dataset[0], dataset[3], dataset[2])
	}
}

func TestStripPart10Header_TooShort(t *testing.T) {
	data := []byte{0x01, 0x02, 0x03}

	_, err := StripPart10Header(data)
	if err == nil {
		t.Error("Expected error for data too short")
	}

	if !bytes.Contains([]byte(err.Error()), []byte("too short")) {
		t.Errorf("Expected 'too short' error, got: %v", err)
	}
}

func TestStripPart10Header_MissingDICM(t *testing.T) {
	// Create data with 132 bytes but no DICM prefix
	data := make([]byte, 200)

	_, err := StripPart10Header(data)
	if err == nil {
		t.Error("Expected error for missing DICM prefix")
	}

	if !bytes.Contains([]byte(err.Error()), []byte("missing DICM")) {
		t.Errorf("Expected 'missing DICM' error, got: %v", err)
	}
}

func TestStripPart10Header_InvalidDICM(t *testing.T) {
	data := make([]byte, 200)
	// Put wrong prefix at offset 128
	copy(data[128:132], []byte("XXXX"))

	_, err := StripPart10Header(data)
	if err == nil {
		t.Error("Expected error for invalid DICM prefix")
	}
}

func TestStripPart10Header_EmptyMetaInfo(t *testing.T) {
	var data []byte

	// 128-byte preamble
	preamble := make([]byte, 128)
	data = append(data, preamble...)

	// DICM prefix
	data = append(data, []byte("DICM")...)

	// Immediately start dataset (group 0x0010)
	data = append(data, 0x10, 0x00, 0x10, 0x00) // Patient Name tag
	data = append(data, 'P', 'N')                // VR
	data = append(data, 0x04, 0x00)              // Length
	data = append(data, []byte("TEST")...)

	dataset, err := StripPart10Header(data)
	if err != nil {
		t.Fatalf("StripPart10Header() error = %v", err)
	}

	// Should start right after DICM prefix
	if len(dataset) < 4 {
		t.Fatal("Dataset too short")
	}

	expectedTag := []byte{0x10, 0x00, 0x10, 0x00}
	if !bytes.Equal(dataset[0:4], expectedTag) {
		t.Errorf("Expected dataset to start with tag 0010,0010")
	}
}

func TestStripPart10Header_MultipleMetaElements(t *testing.T) {
	var data []byte

	// 128-byte preamble
	preamble := make([]byte, 128)
	data = append(data, preamble...)

	// DICM prefix
	data = append(data, []byte("DICM")...)

	// Media Storage SOP Class UID (0002,0002)
	data = append(data, 0x02, 0x00, 0x02, 0x00) // Tag
	data = append(data, 'U', 'I')                // VR
	sopClass := "1.2.3.4\x00"                    // Padded
	sopLength := make([]byte, 2)
	binary.LittleEndian.PutUint16(sopLength, uint16(len(sopClass)))
	data = append(data, sopLength...)
	data = append(data, []byte(sopClass)...)

	// Transfer Syntax UID (0002,0010)
	data = append(data, 0x02, 0x00, 0x10, 0x00) // Tag
	data = append(data, 'U', 'I')                // VR
	tsUID := "1.2.840.10008.1.2\x00"            // Implicit VR Little Endian
	tsLength := make([]byte, 2)
	binary.LittleEndian.PutUint16(tsLength, uint16(len(tsUID)))
	data = append(data, tsLength...)
	data = append(data, []byte(tsUID)...)

	// Dataset starts here
	data = append(data, 0x10, 0x00, 0x10, 0x00) // Patient Name tag
	data = append(data, 'P', 'N')                // VR
	data = append(data, 0x04, 0x00)              // Length
	data = append(data, []byte("TEST")...)

	dataset, err := StripPart10Header(data)
	if err != nil {
		t.Fatalf("StripPart10Header() error = %v", err)
	}

	// Should skip both meta elements
	if len(dataset) < 4 {
		t.Fatal("Dataset too short")
	}

	expectedTag := []byte{0x10, 0x00, 0x10, 0x00}
	if !bytes.Equal(dataset[0:4], expectedTag) {
		t.Errorf("Expected dataset to start with tag 0010,0010")
	}
}

func TestStripPart10Header_LongVRElement(t *testing.T) {
	var data []byte

	// 128-byte preamble
	preamble := make([]byte, 128)
	data = append(data, preamble...)

	// DICM prefix
	data = append(data, []byte("DICM")...)

	// Use OB VR which has 32-bit length
	data = append(data, 0x02, 0x00, 0x01, 0x00) // Tag (0002,0001)
	data = append(data, 'O', 'B')                // VR
	data = append(data, 0x00, 0x00)              // Reserved
	valueData := make([]byte, 100)               // 100 bytes of data
	length := make([]byte, 4)
	binary.LittleEndian.PutUint32(length, uint32(len(valueData)))
	data = append(data, length...)
	data = append(data, valueData...)

	// Dataset starts here
	data = append(data, 0x10, 0x00, 0x10, 0x00) // Patient Name tag
	data = append(data, 'P', 'N')                // VR
	data = append(data, 0x04, 0x00)              // Length
	data = append(data, []byte("TEST")...)

	dataset, err := StripPart10Header(data)
	if err != nil {
		t.Fatalf("StripPart10Header() error = %v", err)
	}

	expectedTag := []byte{0x10, 0x00, 0x10, 0x00}
	if !bytes.Equal(dataset[0:4], expectedTag) {
		t.Errorf("Expected dataset to start with tag 0010,0010")
	}
}

func TestHasPart10Header_Valid(t *testing.T) {
	data := createValidPart10File()

	if !HasPart10Header(data) {
		t.Error("Expected HasPart10Header to return true for valid Part 10 file")
	}
}

func TestHasPart10Header_TooShort(t *testing.T) {
	data := []byte{0x01, 0x02, 0x03}

	if HasPart10Header(data) {
		t.Error("Expected HasPart10Header to return false for short data")
	}
}

func TestHasPart10Header_NoDICM(t *testing.T) {
	data := make([]byte, 200)
	copy(data[128:132], []byte("XXXX"))

	if HasPart10Header(data) {
		t.Error("Expected HasPart10Header to return false without DICM prefix")
	}
}

func TestHasPart10Header_RawDataset(t *testing.T) {
	// Create raw dataset (no Part 10 header)
	var data []byte
	data = append(data, 0x10, 0x00, 0x10, 0x00) // Patient Name tag
	data = append(data, 'P', 'N')                // VR
	data = append(data, 0x04, 0x00)              // Length
	data = append(data, []byte("TEST")...)

	if HasPart10Header(data) {
		t.Error("Expected HasPart10Header to return false for raw dataset")
	}
}
