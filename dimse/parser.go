package dimse

import (
	"encoding/binary"
	"fmt"
	"log/slog"
	"strings"

	"github.com/caio-sobreiro/dicomnet/types"
)

// parseDIMSECommand parses a DIMSE command from raw bytes
func parseDIMSECommand(data []byte) (*types.Message, error) {
	msg := &types.Message{}

	// This is a simplified parser - in practice you'd need a full DICOM parser
	// For now, we'll extract key fields assuming implicit VR little endian

	if len(data) < 12 {
		return nil, fmt.Errorf("DIMSE data too short: %d bytes", len(data))
	}

	slog.Debug("Parsing DIMSE command data", "size_bytes", len(data))

	// Parse DICOM elements with proper variable-length handling
	offset := 0
	for offset < len(data)-8 {
		if offset+8 > len(data) {
			slog.Debug("Not enough data for header", "offset", offset)
			break
		}

		// Read tag (group, element)
		group := binary.LittleEndian.Uint16(data[offset : offset+2])
		element := binary.LittleEndian.Uint16(data[offset+2 : offset+4])
		length := binary.LittleEndian.Uint32(data[offset+4 : offset+8])

		// Sanity check length
		if length > 1000000 { // 1MB limit
			slog.Warn("Element length too large, probably parsing error", "length", length)
			break
		}

		// Ensure we have enough data for the value
		if offset+8+int(length) > len(data) {
			slog.Debug("Not enough data for element value",
				"have_bytes", len(data),
				"need_bytes", offset+8+int(length))
			break
		}

		// Only process command group elements (group 0000)
		if group == 0x0000 {
			valueStart := offset + 8
			valueEnd := valueStart + int(length)

			switch element {
			case 0x0100: // Command Field
				if length == 2 {
					msg.CommandField = binary.LittleEndian.Uint16(data[valueStart:valueEnd])
				} else {
					slog.Warn("Command Field has wrong length", "length", length)
				}
			case 0x0110: // Message ID
				if length == 2 {
					msg.MessageID = binary.LittleEndian.Uint16(data[valueStart:valueEnd])
				} else {
					slog.Warn("Message ID has wrong length", "length", length)
				}
			case 0x0800: // Command Data Set Type
				if length == 2 {
					msg.CommandDataSetType = binary.LittleEndian.Uint16(data[valueStart:valueEnd])
				} else {
					slog.Warn("Command Data Set Type has wrong length", "length", length)
				}
			case 0x0002: // Affected SOP Class UID
				if length > 0 {
					sopClassUID := string(data[valueStart:valueEnd])
					// Remove null padding
					if idx := strings.IndexByte(sopClassUID, 0); idx != -1 {
						sopClassUID = sopClassUID[:idx]
					}
					msg.AffectedSOPClassUID = strings.TrimSpace(sopClassUID)
				}
			case 0x0600: // Move Destination (for C-MOVE-RQ)
				if length > 0 {
					moveDestination := string(data[valueStart:valueEnd])
					// Remove null padding
					if idx := strings.IndexByte(moveDestination, 0); idx != -1 {
						moveDestination = moveDestination[:idx]
					}
					msg.MoveDestination = strings.TrimSpace(moveDestination)
				}
			default:
				// Skip unknown command elements silently
			}
		}

		// Move to next element
		offset += 8 + int(length)

		// Ensure even alignment (DICOM elements should be even-length)
		if length%2 == 1 {
			offset++ // Skip padding byte
		}
	}

	slog.Debug("Parsed DIMSE command",
		"command_field", fmt.Sprintf("0x%04x", msg.CommandField),
		"message_id", msg.MessageID)
	return msg, nil
}

// createDIMSECommand creates a DIMSE command as bytes
func createDIMSECommand(msg *types.Message) []byte {
	var result []byte

	// Command Field (0000,0100)
	result = append(result, 0x00, 0x00, 0x00, 0x01) // Tag
	result = append(result, 0x02, 0x00, 0x00, 0x00) // Length = 2
	cmdBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(cmdBytes, msg.CommandField)
	result = append(result, cmdBytes...)

	// Message ID Being Responded To (0000,0120)
	if msg.MessageIDBeingRespondedTo > 0 {
		result = append(result, 0x00, 0x00, 0x20, 0x01) // Tag
		result = append(result, 0x02, 0x00, 0x00, 0x00) // Length = 2
		msgIDBytes := make([]byte, 2)
		binary.LittleEndian.PutUint16(msgIDBytes, msg.MessageIDBeingRespondedTo)
		result = append(result, msgIDBytes...)
	}

	// Command Data Set Type (0000,0800)
	result = append(result, 0x00, 0x00, 0x00, 0x08) // Tag
	result = append(result, 0x02, 0x00, 0x00, 0x00) // Length = 2
	dataSetTypeBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(dataSetTypeBytes, msg.CommandDataSetType)
	result = append(result, dataSetTypeBytes...)

	// Status (0000,0900)
	result = append(result, 0x00, 0x00, 0x00, 0x09) // Tag
	result = append(result, 0x02, 0x00, 0x00, 0x00) // Length = 2
	statusBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(statusBytes, msg.Status)
	result = append(result, statusBytes...)

	// Affected SOP Class UID (0000,0002)
	if msg.AffectedSOPClassUID != "" {
		result = append(result, 0x00, 0x00, 0x02, 0x00) // Tag
		sopClassUIDBytes := []byte(msg.AffectedSOPClassUID)
		// Ensure even length
		if len(sopClassUIDBytes)%2 == 1 {
			sopClassUIDBytes = append(sopClassUIDBytes, 0x00) // Null pad
		}
		lengthBytes := make([]byte, 4)
		binary.LittleEndian.PutUint32(lengthBytes, uint32(len(sopClassUIDBytes)))
		result = append(result, lengthBytes...)
		result = append(result, sopClassUIDBytes...)
	}

	return result
}
