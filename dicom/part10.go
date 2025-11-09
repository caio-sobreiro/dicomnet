package dicom

import (
	"fmt"
	"log/slog"
	"strings"
)

// StripPart10Header removes the DICOM Part 10 preamble and File Meta Information
// to extract just the dataset.
//
// DICOM Part 10 files contain:
//   - 128 byte preamble
//   - 4 byte "DICM" prefix
//   - File Meta Information elements (group 0x0002)
//   - Dataset (the actual DICOM data)
//
// This function is useful when you need to send a DICOM dataset via DIMSE
// operations (like C-STORE), which expect only the dataset without the
// Part 10 wrapper.
//
// Parameters:
//   - data: The complete DICOM Part 10 file data
//
// Returns:
//   - Dataset bytes (without preamble and file meta information)
//   - Error if the data is not a valid DICOM Part 10 file
//
// Example:
//
//	fileData, _ := os.ReadFile("image.dcm")
//	datasetOnly, err := dicom.StripPart10Header(fileData)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	// Now datasetOnly can be sent via C-STORE
func StripPart10Header(data []byte) ([]byte, error) {
	if len(data) < 132 {
		return nil, fmt.Errorf("data too short to be DICOM Part 10 (need at least 132 bytes, got %d)", len(data))
	}

	// Check for DICM prefix at offset 128
	if string(data[128:132]) != "DICM" {
		return nil, fmt.Errorf("not a valid DICOM Part 10 file (missing DICM prefix at offset 128)")
	}

	// Skip preamble (128) + DICM (4) = start at offset 132
	offset := 132

	var transferSyntaxUID string

	// Skip all group 0x0002 elements (File Meta Information)
	for offset+8 <= len(data) {
		group := uint16(data[offset]) | (uint16(data[offset+1]) << 8)
		element := uint16(data[offset+2]) | (uint16(data[offset+3]) << 8)

		// If we've passed group 0x0002, we're at the dataset
		if group != 0x0002 {
			break
		}

		// Read VR (2 bytes)
		vr := string(data[offset+4 : offset+6])

		var length uint32
		var valueOffset int

		// Some VRs use different length encoding
		if vr == "OB" || vr == "OW" || vr == "OF" || vr == "SQ" || vr == "UN" || vr == "UT" {
			// Explicit VR with 32-bit length
			offset += 8 // Skip tag (4) + VR (2) + reserved (2)
			if offset+4 > len(data) {
				break
			}
			length = uint32(data[offset]) | (uint32(data[offset+1]) << 8) |
				(uint32(data[offset+2]) << 16) | (uint32(data[offset+3]) << 24)
			offset += 4
			valueOffset = offset
		} else {
			// Explicit VR with 16-bit length
			offset += 6 // Skip tag (4) + VR (2)
			if offset+2 > len(data) {
				break
			}
			length = uint32(data[offset]) | (uint32(data[offset+1]) << 8)
			offset += 2
			valueOffset = offset
		}

		// Check if this is Transfer Syntax UID (0002,0010)
		if group == 0x0002 && element == 0x0010 {
			if valueOffset+int(length) <= len(data) {
				transferSyntaxUID = string(data[valueOffset : valueOffset+int(length)])
				// Remove any padding
				transferSyntaxUID = strings.TrimRight(transferSyntaxUID, "\x00 ")
			}
		}

		// Skip value
		offset += int(length)
		if offset > len(data) {
			break
		}
	}

	if transferSyntaxUID != "" {
		slog.Debug("Found Transfer Syntax UID in File Meta Information",
			"transfer_syntax", transferSyntaxUID,
			"dataset_start_offset", offset)
	}

	if offset >= len(data) {
		return nil, fmt.Errorf("failed to find dataset after File Meta Information")
	}

	return data[offset:], nil
}

// HasPart10Header checks if the data starts with a DICOM Part 10 header.
//
// Returns true if the data contains the 128-byte preamble followed by "DICM".
func HasPart10Header(data []byte) bool {
	if len(data) < 132 {
		return false
	}
	return string(data[128:132]) == "DICM"
}
