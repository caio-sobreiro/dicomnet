package client

import (
	"encoding/binary"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/caio-sobreiro/dicomnet/dimse"
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

// SendCStore sends a C-STORE request and waits for response
func (a *Association) SendCStore(req *CStoreRequest) (*CStoreResponse, error) {
	// Find presentation context for this SOP Class
	presContextID, err := a.GetPresentationContextID(req.SOPClassUID)
	if err != nil {
		return nil, fmt.Errorf("no presentation context for SOP class %s: %w", req.SOPClassUID, err)
	}

	// Build C-STORE-RQ command
	command := &types.Message{
		CommandField:           dimse.CStoreRQ,
		MessageID:              req.MessageID,
		Priority:               0x0000, // Medium
		CommandDataSetType:     0x0000, // Dataset present
		AffectedSOPClassUID:    req.SOPClassUID,
		AffectedSOPInstanceUID: req.SOPInstanceUID,
	}

	// Encode command
	commandData, err := encodeCommand(command)
	if err != nil {
		return nil, fmt.Errorf("failed to encode command: %w", err)
	}

	// Send C-STORE-RQ with dataset
	if err := a.sendDIMSEMessage(presContextID, commandData, req.Data); err != nil {
		return nil, fmt.Errorf("failed to send C-STORE: %w", err)
	}

	slog.Debug("Sent C-STORE-RQ",
		"sop_class", req.SOPClassUID,
		"sop_instance", req.SOPInstanceUID,
		"data_size", len(req.Data))

	// Receive C-STORE-RSP
	resp, err := a.receiveCStoreResponse()
	if err != nil {
		return nil, fmt.Errorf("failed to receive C-STORE-RSP: %w", err)
	}

	return resp, nil
}

// sendDIMSEMessage sends a DIMSE message with optional dataset
func (a *Association) sendDIMSEMessage(presContextID byte, commandData []byte, datasetData []byte) error {
	// Send command in P-DATA-TF
	if err := a.sendPDataTF(presContextID, commandData, true, true); err != nil {
		return err
	}

	// Send dataset if present
	if len(datasetData) > 0 {
		if err := a.sendPDataTF(presContextID, datasetData, false, true); err != nil {
			return err
		}
	}

	return nil
}

// sendPDataTF sends a P-DATA-TF PDU
func (a *Association) sendPDataTF(presContextID byte, data []byte, isCommand bool, isLast bool) error {
	// Calculate max data per PDV (PDU length - PDU header - PDV header)
	maxPDVData := int(a.maxPDULength) - 6 - 6

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
		if _, err := a.conn.Write(fullPDU); err != nil {
			return fmt.Errorf("failed to write PDU: %w", err)
		}

		offset += chunkSize
	}

	return nil
}

// receiveCStoreResponse receives and parses C-STORE-RSP
func (a *Association) receiveCStoreResponse() (*CStoreResponse, error) {
	msg, _, err := a.receiveDIMSEMessage()
	if err != nil {
		return nil, err
	}

	if msg.CommandField != dimse.CStoreRSP {
		return nil, fmt.Errorf("unexpected command: 0x%04x (expected C-STORE-RSP)", msg.CommandField)
	}

	return &CStoreResponse{
		Status:         msg.Status,
		MessageID:      msg.MessageIDBeingRespondedTo,
		SOPClassUID:    msg.AffectedSOPClassUID,
		SOPInstanceUID: msg.AffectedSOPInstanceUID,
	}, nil
}

// encodeCommand encodes a DIMSE command message using Implicit VR Little Endian
func encodeCommand(msg *types.Message) ([]byte, error) {
	buf := make([]byte, 0, 256)

	// Command Group Length (0000,0000) - will calculate later
	buf = appendImplicitElement(buf, 0x0000, 0x0000, make([]byte, 4)) // Placeholder
	lengthPos := len(buf) - 4

	// Affected SOP Class UID (0000,0002)
	sopClassBytes := []byte(msg.AffectedSOPClassUID)
	if len(sopClassBytes)%2 == 1 {
		sopClassBytes = append(sopClassBytes, 0x00) // Pad to even
	}
	buf = appendImplicitElement(buf, 0x0000, 0x0002, sopClassBytes)

	// Command Field (0000,0100)
	cmdBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(cmdBytes, msg.CommandField)
	buf = appendImplicitElement(buf, 0x0000, 0x0100, cmdBytes)

	// Message ID (0000,0110)
	msgIDBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(msgIDBytes, msg.MessageID)
	buf = appendImplicitElement(buf, 0x0000, 0x0110, msgIDBytes)

	// Priority (0000,0700)
	priorityBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(priorityBytes, msg.Priority)
	buf = appendImplicitElement(buf, 0x0000, 0x0700, priorityBytes)

	// Command Data Set Type (0000,0800)
	datasetTypeBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(datasetTypeBytes, msg.CommandDataSetType)
	buf = appendImplicitElement(buf, 0x0000, 0x0800, datasetTypeBytes)

	// Affected SOP Instance UID (0000,1000)
	sopInstBytes := []byte(msg.AffectedSOPInstanceUID)
	if len(sopInstBytes)%2 == 1 {
		sopInstBytes = append(sopInstBytes, 0x00) // Pad to even
	}
	buf = appendImplicitElement(buf, 0x0000, 0x1000, sopInstBytes)

	// Update Command Group Length
	groupLength := uint32(len(buf) - lengthPos - 4)
	binary.LittleEndian.PutUint32(buf[lengthPos:lengthPos+4], groupLength)

	return buf, nil
}

// appendImplicitElement appends a DICOM element using Implicit VR (no VR field)
func appendImplicitElement(buf []byte, group, element uint16, value []byte) []byte {
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

// decodeCommand decodes a DIMSE command message
func decodeCommand(data []byte) (*types.Message, error) {
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
		case group == 0x0000 && element == 0x0600:
			msg.MoveDestination = strings.TrimRight(string(value), "\x00 ")
		}

		offset += 8 + int(length)
	}

	return msg, nil
}

// receiveDIMSEMessage reads a complete DIMSE message (command and optional dataset)
// from the association connection.
func (a *Association) receiveDIMSEMessage() (*types.Message, []byte, error) {
	var commandData []byte
	var datasetData []byte
	commandComplete := false
	datasetComplete := false
	datasetExpected := false
	var currentMsg *types.Message

	for {
		header := make([]byte, 6)
		if _, err := io.ReadFull(a.conn, header); err != nil {
			return nil, nil, fmt.Errorf("failed to read PDU header: %w", err)
		}

		pduType := header[0]
		pduLength := binary.BigEndian.Uint32(header[2:6])

		switch pduType {
		case pdu.TypePDataTF:
			payload := make([]byte, pduLength)
			if _, err := io.ReadFull(a.conn, payload); err != nil {
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
						decoded, err := decodeCommand(commandData)
						if err != nil {
							return nil, nil, fmt.Errorf("failed to decode command: %w", err)
						}
						currentMsg = decoded

						if currentMsg.CommandDataSetType != 0x0101 {
							datasetExpected = true
							if datasetData == nil || len(datasetData) == 0 {
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
			if _, err := io.ReadFull(a.conn, abortData); err != nil {
				return nil, nil, fmt.Errorf("failed to read ABORT data: %w", err)
			}

			var source, reason byte
			if len(abortData) >= 4 {
				source = abortData[2]
				reason = abortData[3]
			}

			slog.Error("Received A-ABORT from peer",
				"source", source,
				"reason", reason)

			return nil, nil, fmt.Errorf("received A-ABORT PDU (source=%d, reason=%d)", source, reason)
		default:
			// Skip payload for unexpected PDU types to maintain stream alignment
			discard := make([]byte, pduLength)
			if _, err := io.ReadFull(a.conn, discard); err != nil {
				return nil, nil, fmt.Errorf("failed to read unexpected PDU payload: %w", err)
			}
			return nil, nil, fmt.Errorf("unexpected PDU type: 0x%02x", pduType)
		}

		if commandComplete && (!datasetExpected || datasetComplete) {
			return currentMsg, datasetData, nil
		}
	}
}
