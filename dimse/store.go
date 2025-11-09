package dimse

import (
	"encoding/binary"
	"fmt"
	"io"
	"strings"

	"github.com/caio-sobreiro/dicomnet/pdu"
	"github.com/caio-sobreiro/dicomnet/types"
)

// CStoreRequest represents a C-STORE request
type CStoreRequest struct {
	SOPClassUID    string
	SOPInstanceUID string
	Data           []byte
	MessageID      uint16
}

// CStoreResponse represents a C-STORE response
type CStoreResponse struct {
	Status         uint16
	MessageID      uint16
	SOPClassUID    string
	SOPInstanceUID string
}

// Connection interface for sending/receiving DICOM data
type Connection interface {
	io.ReadWriter
}

// SendCStore sends a C-STORE request and waits for response
func SendCStore(conn Connection, presContextID byte, maxPDULength uint32, req *CStoreRequest) (*CStoreResponse, error) {
	// Build C-STORE-RQ command
	command := &types.Message{
		CommandField:           CStoreRQ,
		MessageID:              req.MessageID,
		Priority:               0x0002, // Medium priority (must be non-zero to be encoded)
		CommandDataSetType:     0x0000, // Dataset present
		AffectedSOPClassUID:    req.SOPClassUID,
		AffectedSOPInstanceUID: req.SOPInstanceUID,
	}

	// Encode command
	commandData, err := EncodeCommand(command)
	if err != nil {
		return nil, fmt.Errorf("failed to encode command: %w", err)
	}

	// Send C-STORE-RQ with dataset
	if err := SendDIMSEMessage(conn, presContextID, maxPDULength, commandData, req.Data); err != nil {
		return nil, fmt.Errorf("failed to send C-STORE: %w", err)
	}

	// Receive C-STORE-RSP
	msg, _, err := ReceiveDIMSEMessage(conn)
	if err != nil {
		return nil, fmt.Errorf("failed to receive C-STORE-RSP: %w", err)
	}

	if msg.CommandField != CStoreRSP {
		return nil, fmt.Errorf("unexpected command: 0x%04x (expected C-STORE-RSP)", msg.CommandField)
	}

	return &CStoreResponse{
		Status:         msg.Status,
		MessageID:      msg.MessageIDBeingRespondedTo,
		SOPClassUID:    msg.AffectedSOPClassUID,
		SOPInstanceUID: msg.AffectedSOPInstanceUID,
	}, nil
}

// SendDIMSEMessage sends a DIMSE message with optional dataset
func SendDIMSEMessage(conn Connection, presContextID byte, maxPDULength uint32, commandData []byte, datasetData []byte) error {
	// Send command in P-DATA-TF
	if err := SendPDataTF(conn, presContextID, maxPDULength, commandData, true, true); err != nil {
		return err
	}

	// Send dataset if present
	if len(datasetData) > 0 {
		if err := SendPDataTF(conn, presContextID, maxPDULength, datasetData, false, true); err != nil {
			return err
		}
	}

	return nil
}

// SendPDataTF sends a P-DATA-TF PDU
func SendPDataTF(conn Connection, presContextID byte, maxPDULength uint32, data []byte, isCommand bool, isLast bool) error {
	// Calculate max data per PDV (PDU length - PDU header - PDV header)
	maxPDVData := int(maxPDULength) - 6 - 6

	offset := 0
	for offset < len(data) {
		// Calculate chunk size
		chunkSize := len(data) - offset
		lastFragment := true
		if chunkSize > maxPDVData {
			chunkSize = maxPDVData
			lastFragment = false
		}

		// Build PDV (Presentation Data Value)
		pdvLength := uint32(chunkSize + 2) // +2 for PDV header
		pdv := make([]byte, 0, pdvLength+4)

		// PDV length (4 bytes)
		pdvLengthBytes := make([]byte, 4)
		binary.BigEndian.PutUint32(pdvLengthBytes, pdvLength)
		pdv = append(pdv, pdvLengthBytes...)

		// Presentation Context ID (1 byte)
		pdv = append(pdv, presContextID)

		// Message Control Header (1 byte)
		// Bit 0: 0=data, 1=command
		// Bit 1: 0=not last, 1=last fragment
		controlHeader := byte(0)
		if isCommand {
			controlHeader |= 0x01
		}
		if lastFragment && isLast {
			controlHeader |= 0x02
		}
		pdv = append(pdv, controlHeader)

		// Data fragment
		pdv = append(pdv, data[offset:offset+chunkSize]...)

		// Build P-DATA-TF PDU
		pduHeader := make([]byte, 6)
		pduHeader[0] = pdu.TypePDataTF
		pduHeader[1] = 0x00
		binary.BigEndian.PutUint32(pduHeader[2:6], uint32(len(pdv)))

		// Combine PDU header and PDV into single write for atomicity
		fullPDU := append(pduHeader, pdv...)

		// Send complete PDU
		if _, err := conn.Write(fullPDU); err != nil {
			return fmt.Errorf("failed to write PDU: %w", err)
		}

		offset += chunkSize
	}

	return nil
}

// EncodeCommand encodes a DIMSE command message using Implicit VR Little Endian
func EncodeCommand(msg *types.Message) ([]byte, error) {
	buf := make([]byte, 0, 256)

	// Command Group Length (0000,0000) - will calculate later
	buf = AppendImplicitElement(buf, 0x0000, 0x0000, make([]byte, 4)) // Placeholder
	lengthPos := len(buf) - 4

	// Affected SOP Class UID (0000,0002) - optional
	if msg.AffectedSOPClassUID != "" {
		sopClassBytes := []byte(msg.AffectedSOPClassUID)
		if len(sopClassBytes)%2 == 1 {
			sopClassBytes = append(sopClassBytes, 0x00) // Pad to even
		}
		buf = AppendImplicitElement(buf, 0x0000, 0x0002, sopClassBytes)
	}

	// Requested SOP Class UID (0000,0003) - optional
	if msg.RequestedSOPClassUID != "" {
		sopClassBytes := []byte(msg.RequestedSOPClassUID)
		if len(sopClassBytes)%2 == 1 {
			sopClassBytes = append(sopClassBytes, 0x00) // Pad to even
		}
		buf = AppendImplicitElement(buf, 0x0000, 0x0003, sopClassBytes)
	}

	// Command Field (0000,0100) - required
	cmdBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(cmdBytes, msg.CommandField)
	buf = AppendImplicitElement(buf, 0x0000, 0x0100, cmdBytes)

	// Message ID (0000,0110) - optional (not in responses)
	if msg.MessageID != 0 {
		msgIDBytes := make([]byte, 2)
		binary.LittleEndian.PutUint16(msgIDBytes, msg.MessageID)
		buf = AppendImplicitElement(buf, 0x0000, 0x0110, msgIDBytes)
	}

	// Message ID Being Responded To (0000,0120) - optional (in responses and C-CANCEL)
	if msg.MessageIDBeingRespondedTo != 0 {
		msgIDBytes := make([]byte, 2)
		binary.LittleEndian.PutUint16(msgIDBytes, msg.MessageIDBeingRespondedTo)
		buf = AppendImplicitElement(buf, 0x0000, 0x0120, msgIDBytes)
	}

	// Move Destination (0000,0600) - optional (for C-MOVE)
	if msg.MoveDestination != "" {
		moveDestBytes := []byte(msg.MoveDestination)
		if len(moveDestBytes)%2 == 1 {
			moveDestBytes = append(moveDestBytes, 0x20) // Pad with space
		}
		buf = AppendImplicitElement(buf, 0x0000, 0x0600, moveDestBytes)
	}

	// Priority (0000,0700) - optional
	if msg.Priority != 0 {
		priorityBytes := make([]byte, 2)
		binary.LittleEndian.PutUint16(priorityBytes, msg.Priority)
		buf = AppendImplicitElement(buf, 0x0000, 0x0700, priorityBytes)
	}

	// Command Data Set Type (0000,0800) - required
	datasetTypeBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(datasetTypeBytes, msg.CommandDataSetType)
	buf = AppendImplicitElement(buf, 0x0000, 0x0800, datasetTypeBytes)

	// Status (0000,0900) - optional (in responses)
	if msg.Status != 0 {
		statusBytes := make([]byte, 2)
		binary.LittleEndian.PutUint16(statusBytes, msg.Status)
		buf = AppendImplicitElement(buf, 0x0000, 0x0900, statusBytes)
	}

	// Affected SOP Instance UID (0000,1000) - optional
	if msg.AffectedSOPInstanceUID != "" {
		sopInstBytes := []byte(msg.AffectedSOPInstanceUID)
		if len(sopInstBytes)%2 == 1 {
			sopInstBytes = append(sopInstBytes, 0x00) // Pad to even
		}
		buf = AppendImplicitElement(buf, 0x0000, 0x1000, sopInstBytes)
	}

	// C-MOVE response counters (optional, only for C-MOVE-RSP)
	if msg.NumberOfRemainingSuboperations != nil {
		remaining := make([]byte, 2)
		binary.LittleEndian.PutUint16(remaining, *msg.NumberOfRemainingSuboperations)
		buf = AppendImplicitElement(buf, 0x0000, 0x1020, remaining)
	}

	if msg.NumberOfCompletedSuboperations != nil {
		completed := make([]byte, 2)
		binary.LittleEndian.PutUint16(completed, *msg.NumberOfCompletedSuboperations)
		buf = AppendImplicitElement(buf, 0x0000, 0x1021, completed)
	}

	if msg.NumberOfFailedSuboperations != nil {
		failed := make([]byte, 2)
		binary.LittleEndian.PutUint16(failed, *msg.NumberOfFailedSuboperations)
		buf = AppendImplicitElement(buf, 0x0000, 0x1022, failed)
	}

	if msg.NumberOfWarningSuboperations != nil {
		warning := make([]byte, 2)
		binary.LittleEndian.PutUint16(warning, *msg.NumberOfWarningSuboperations)
		buf = AppendImplicitElement(buf, 0x0000, 0x1023, warning)
	}

	// Update Command Group Length
	groupLength := uint32(len(buf) - lengthPos - 4)
	binary.LittleEndian.PutUint32(buf[lengthPos:lengthPos+4], groupLength)

	return buf, nil
}

// AppendImplicitElement appends a DICOM element using Implicit VR (no VR field)
func AppendImplicitElement(buf []byte, group, element uint16, value []byte) []byte {
	// Group (2 bytes, little endian)
	buf = append(buf, byte(group), byte(group>>8))
	// Element (2 bytes, little endian)
	buf = append(buf, byte(element), byte(element>>8))
	// Length (4 bytes, little endian)
	length := uint32(len(value))
	buf = append(buf, byte(length), byte(length>>8), byte(length>>16), byte(length>>24))
	// Value
	buf = append(buf, value...)
	return buf
}

// DecodeCommand decodes a DIMSE command message
func DecodeCommand(data []byte) (*types.Message, error) {
	msg := &types.Message{
		CommandDataSetType: 0x0101, // Default to "no dataset present"
	}
	offset := 0

	for offset+8 <= len(data) {
		group := binary.LittleEndian.Uint16(data[offset : offset+2])
		element := binary.LittleEndian.Uint16(data[offset+2 : offset+4])
		length := binary.LittleEndian.Uint32(data[offset+4 : offset+8])

		if offset+8+int(length) > len(data) {
			break
		}

		value := data[offset+8 : offset+8+int(length)]

		switch {
		case group == 0x0000 && element == 0x0002:
			msg.AffectedSOPClassUID = strings.TrimRight(string(value), "\x00 ")
		case group == 0x0000 && element == 0x0003:
			msg.RequestedSOPClassUID = strings.TrimRight(string(value), "\x00 ")
		case group == 0x0000 && element == 0x0100:
			if len(value) >= 2 {
				msg.CommandField = binary.LittleEndian.Uint16(value[:2])
			}
		case group == 0x0000 && element == 0x0110:
			if len(value) >= 2 {
				msg.MessageID = binary.LittleEndian.Uint16(value[:2])
			}
		case group == 0x0000 && element == 0x0120:
			if len(value) >= 2 {
				msg.MessageIDBeingRespondedTo = binary.LittleEndian.Uint16(value[:2])
			}
		case group == 0x0000 && element == 0x0600:
			msg.MoveDestination = strings.TrimRight(string(value), "\x00 ")
		case group == 0x0000 && element == 0x0700:
			if len(value) >= 2 {
				msg.Priority = binary.LittleEndian.Uint16(value[:2])
			}
		case group == 0x0000 && element == 0x0800:
			if len(value) >= 2 {
				msg.CommandDataSetType = binary.LittleEndian.Uint16(value[:2])
			}
		case group == 0x0000 && element == 0x0900:
			if len(value) >= 2 {
				msg.Status = binary.LittleEndian.Uint16(value[:2])
			}
		case group == 0x0000 && element == 0x1000:
			msg.AffectedSOPInstanceUID = strings.TrimRight(string(value), "\x00 ")
		case group == 0x0000 && element == 0x1020:
			if len(value) >= 2 {
				val := binary.LittleEndian.Uint16(value[:2])
				msg.NumberOfRemainingSuboperations = &val
			}
		case group == 0x0000 && element == 0x1021:
			if len(value) >= 2 {
				val := binary.LittleEndian.Uint16(value[:2])
				msg.NumberOfCompletedSuboperations = &val
			}
		case group == 0x0000 && element == 0x1022:
			if len(value) >= 2 {
				val := binary.LittleEndian.Uint16(value[:2])
				msg.NumberOfFailedSuboperations = &val
			}
		case group == 0x0000 && element == 0x1023:
			if len(value) >= 2 {
				val := binary.LittleEndian.Uint16(value[:2])
				msg.NumberOfWarningSuboperations = &val
			}
		}

		offset += 8 + int(length)
	}

	return msg, nil
}

// ReceiveDIMSEMessage reads a complete DIMSE message (command and optional dataset)
func ReceiveDIMSEMessage(conn Connection) (*types.Message, []byte, error) {
	var commandData []byte
	var datasetData []byte
	commandComplete := false
	datasetComplete := false
	datasetExpected := false
	var currentMsg *types.Message

	for {
		header := make([]byte, 6)
		if _, err := io.ReadFull(conn, header); err != nil {
			return nil, nil, fmt.Errorf("failed to read PDU header: %w", err)
		}

		pduType := header[0]
		pduLength := binary.BigEndian.Uint32(header[2:6])

		switch pduType {
		case pdu.TypePDataTF:
			payload := make([]byte, pduLength)
			if _, err := io.ReadFull(conn, payload); err != nil {
				return nil, nil, fmt.Errorf("failed to read PDU data: %w", err)
			}

			offset := 0
			for offset < len(payload) {
				if offset+6 > len(payload) {
					return nil, nil, fmt.Errorf("malformed PDV encountered")
				}

				pdvLength := binary.BigEndian.Uint32(payload[offset : offset+4])
				end := offset + 4 + int(pdvLength)
				if end > len(payload) {
					return nil, nil, fmt.Errorf("PDV length exceeds PDU payload")
				}

				controlHeader := payload[offset+5]
				value := payload[offset+6 : end]
				isCommand := controlHeader&0x01 != 0
				isLastFragment := controlHeader&0x02 != 0

				if isCommand {
					commandData = append(commandData, value...)
					if isLastFragment {
						commandComplete = true
						decoded, err := DecodeCommand(commandData)
						if err != nil {
							return nil, nil, fmt.Errorf("failed to decode command: %w", err)
						}
						currentMsg = decoded

						if currentMsg.CommandDataSetType != 0x0101 {
							datasetExpected = true
							if len(datasetData) == 0 {
								datasetComplete = false
							}
						} else {
							datasetExpected = false
							datasetComplete = true
						}
					}
				} else {
					datasetData = append(datasetData, value...)
					if isLastFragment {
						datasetComplete = true
					}
				}

				offset = end
			}
		case 0x07: // A-ABORT
			abortData := make([]byte, pduLength)
			if _, err := io.ReadFull(conn, abortData); err != nil {
				return nil, nil, fmt.Errorf("failed to read ABORT data: %w", err)
			}

			var source, reason byte
			if len(abortData) >= 4 {
				source = abortData[2]
				reason = abortData[3]
			}

			return nil, nil, fmt.Errorf("received A-ABORT PDU (source=%d, reason=%d)", source, reason)
		default:
			// Skip payload for unexpected PDU types to maintain stream alignment
			discard := make([]byte, pduLength)
			if _, err := io.ReadFull(conn, discard); err != nil {
				return nil, nil, fmt.Errorf("failed to read unexpected PDU payload: %w", err)
			}
			return nil, nil, fmt.Errorf("unexpected PDU type: 0x%02x", pduType)
		}

		if commandComplete && (!datasetExpected || datasetComplete) {
			return currentMsg, datasetData, nil
		}
	}
}
