package dimse

import (
	"context"
	"encoding/binary"
	"fmt"
	"log/slog"

	"github.com/caio-sobreiro/dicomnet/interfaces"
	"github.com/caio-sobreiro/dicomnet/types"
)

// Command types
const (
	CStoreRQ  = 0x0001
	CStoreRSP = 0x8001
	CFindRQ   = 0x0020
	CFindRSP  = 0x8020
	CMoveRQ   = 0x0021
	CMoveRSP  = 0x8021
	CEchoRQ   = 0x0030
	CEchoRSP  = 0x8030
)

// Status codes
const (
	StatusSuccess = 0x0000
	StatusPending = 0xFF00
	StatusFailure = 0xC000
)

// PDULayer interface for sending responses
type PDULayer interface {
	SendDIMSEResponse(presContextID byte, commandData []byte) error
	SendDIMSEResponseWithDataset(presContextID byte, commandData []byte, datasetData []byte) error
}

// Service manages DIMSE operations and message routing
type Service struct {
	handler     interfaces.ServiceHandler
	commandData []byte
	datasetData []byte
	currentMsg  *types.Message
	logger      *slog.Logger
}

// responseHandler implements ResponseSender for streaming responses
type responseHandler struct {
	service       *Service
	presContextID byte
	pduLayer      PDULayer
}

// SendResponse implements ResponseSender interface
func (r *responseHandler) SendResponse(msg *types.Message, data []byte) error {
	return r.service.sendDIMSEResponse(msg, data, r.presContextID, r.pduLayer)
}

// NewService creates a new DIMSE service with a handler
func NewService(handler interfaces.ServiceHandler, logger *slog.Logger) *Service {
	if logger == nil {
		logger = slog.Default()
	}
	return &Service{
		handler: handler,
		logger:  logger,
	}
}

// HandleDIMSEMessage processes DIMSE messages and routes to appropriate service
func (d *Service) HandleDIMSEMessage(presContextID byte, msgCtrlHeader byte, data []byte, pduLayer PDULayer) error {
	// Create context for this message handling
	ctx := context.Background()

	d.logger.Debug("Processing DIMSE message",
		"context_id", presContextID,
		"control_header", fmt.Sprintf("0x%02x", msgCtrlHeader))

	// Check message control header
	// 0x01 = command, more fragments
	// 0x02 = dataset, last fragment
	// 0x03 = command, last fragment
	// 0x00 = dataset, more fragments

	isCommand := (msgCtrlHeader & 0x01) != 0
	isLastFragment := (msgCtrlHeader & 0x02) != 0

	if isCommand {
		// This is command data
		d.logger.Debug("Received command data", "size_bytes", len(data))
		if isLastFragment {
			// Complete command in one fragment
			d.commandData = data
			msg, err := parseDIMSECommand(data, d.logger)
			if err != nil {
				return fmt.Errorf("failed to parse DIMSE command: %v", err)
			}
			d.currentMsg = msg

			// If CommandDataSetType indicates no dataset, process immediately
			if msg.CommandDataSetType == 0x0101 {
				return d.processCompleteMessage(ctx, presContextID, pduLayer)
			}
		} else {
			// Multi-fragment command (accumulate)
			d.commandData = append(d.commandData, data...)
		}
	} else {
		// This is dataset data
		d.logger.Debug("Received dataset data", "size_bytes", len(data))
		if isLastFragment {
			// Complete dataset received
			d.datasetData = append(d.datasetData, data...)
			return d.processCompleteMessage(ctx, presContextID, pduLayer)
		} else {
			// Multi-fragment dataset (accumulate)
			d.datasetData = append(d.datasetData, data...)
		}
	}

	return nil
}

// processCompleteMessage processes a complete DIMSE message (command + optional dataset)
func (d *Service) processCompleteMessage(ctx context.Context, presContextID byte, pduLayer PDULayer) error {
	if d.currentMsg == nil {
		return fmt.Errorf("no current message to process")
	}

	d.logger.InfoContext(ctx, "Processing complete DIMSE message",
		"command_field", fmt.Sprintf("0x%04x", d.currentMsg.CommandField),
		"message_id", d.currentMsg.MessageID,
		"dataset_size", len(d.datasetData))

	// Check if handler supports streaming (for multi-response operations like C-FIND)
	if streamingHandler, ok := d.handler.(interfaces.StreamingServiceHandler); ok {
		d.logger.DebugContext(ctx, "Using streaming handler for multi-response operation")
		responder := &responseHandler{
			service:       d,
			presContextID: presContextID,
			pduLayer:      pduLayer,
		}

		err := streamingHandler.HandleDIMSEStreaming(ctx, d.currentMsg, d.datasetData, responder)

		// Reset for next message
		d.commandData = nil
		d.datasetData = nil
		d.currentMsg = nil

		return err
	}

	// Fallback to single-response handler
	responseMsg, responseData, err := d.handler.HandleDIMSE(ctx, d.currentMsg, d.datasetData)
	if err != nil {
		return fmt.Errorf("service handler failed: %v", err)
	}

	// Send response
	err = d.sendDIMSEResponse(responseMsg, responseData, presContextID, pduLayer)

	// Reset for next message
	d.commandData = nil
	d.datasetData = nil
	d.currentMsg = nil

	return err
}

// sendDIMSEResponse sends a DIMSE response
func (d *Service) sendDIMSEResponse(msg *types.Message, data []byte, presContextID byte, pduLayer PDULayer) error {
	commandData := d.createDIMSECommand(msg)
	return pduLayer.SendDIMSEResponseWithDataset(presContextID, commandData, data)
}

// createDIMSECommand creates a DIMSE command dataset
func (d *Service) createDIMSECommand(msg *types.Message) []byte {
	var elements []byte

	// Affected SOP Class UID (0000,0002) - for echo
	if msg.AffectedSOPClassUID != "" {
		sopClassUID := msg.AffectedSOPClassUID
		if len(sopClassUID)%2 == 1 {
			sopClassUID += "\x00"
		}
		elements = append(elements, 0x00, 0x00, 0x02, 0x00) // Tag
		sopLen := make([]byte, 4)
		binary.LittleEndian.PutUint32(sopLen, uint32(len(sopClassUID)))
		elements = append(elements, sopLen...)
		elements = append(elements, []byte(sopClassUID)...)
	}

	// Command Field (0000,0100)
	elements = append(elements, 0x00, 0x00, 0x00, 0x01) // Tag
	elements = append(elements, 0x02, 0x00, 0x00, 0x00) // Length = 2
	cmdField := make([]byte, 2)
	binary.LittleEndian.PutUint16(cmdField, msg.CommandField)
	elements = append(elements, cmdField...)

	// Message ID Being Responded To (0000,0120)
	if msg.MessageIDBeingRespondedTo > 0 {
		elements = append(elements, 0x00, 0x00, 0x20, 0x01) // Tag
		elements = append(elements, 0x02, 0x00, 0x00, 0x00) // Length = 2
		msgID := make([]byte, 2)
		binary.LittleEndian.PutUint16(msgID, msg.MessageIDBeingRespondedTo)
		elements = append(elements, msgID...)
	}

	// CommandDataSetType (0000,0800)
	elements = append(elements, 0x00, 0x00, 0x00, 0x08) // Tag
	elements = append(elements, 0x02, 0x00, 0x00, 0x00) // Length = 2
	cmdDataSetType := make([]byte, 2)
	binary.LittleEndian.PutUint16(cmdDataSetType, msg.CommandDataSetType)
	elements = append(elements, cmdDataSetType...)

	// Status (0000,0900)
	elements = append(elements, 0x00, 0x00, 0x00, 0x09) // Tag
	elements = append(elements, 0x02, 0x00, 0x00, 0x00) // Length = 2
	status := make([]byte, 2)
	binary.LittleEndian.PutUint16(status, msg.Status)
	elements = append(elements, status...)

	// Add Group Length (0000,0000) at the beginning
	groupLengthValue := make([]byte, 4)
	binary.LittleEndian.PutUint32(groupLengthValue, uint32(len(elements)))

	var commandSet []byte
	commandSet = append(commandSet, 0x00, 0x00, 0x00, 0x00) // Group Length tag
	commandSet = append(commandSet, 0x04, 0x00, 0x00, 0x00) // Length = 4
	commandSet = append(commandSet, groupLengthValue...)    // Value
	commandSet = append(commandSet, elements...)

	return commandSet
}
