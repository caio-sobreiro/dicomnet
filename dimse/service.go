package dimse

import (
	"context"
	"encoding/binary"
	"fmt"
	"log/slog"

	"github.com/caio-sobreiro/dicomnet/dicom"
	"github.com/caio-sobreiro/dicomnet/interfaces"
	"github.com/caio-sobreiro/dicomnet/types"
)

// Command types
const (
	CStoreRQ  = 0x0001
	CStoreRSP = 0x8001
	CGetRQ    = 0x0010
	CGetRSP   = 0x8010
	CFindRQ   = 0x0020
	CFindRSP  = 0x8020
	CMoveRQ   = 0x0021
	CMoveRSP  = 0x8021
	CEchoRQ   = 0x0030
	CEchoRSP  = 0x8030
	CCancelRQ = 0x0FFF
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
	GetTransferSyntax(presContextID byte) (string, error)
}

// Service manages DIMSE operations and message routing
type Service struct {
	handler     interfaces.ServiceHandler
	commandData []byte
	datasetData []byte
	currentMsg  *types.Message
	logger      *slog.Logger
	transferUID string
	contextID   byte
}

// responseHandler implements ResponseSender for streaming responses
type responseHandler struct {
	service               *Service
	presContextID         byte
	pduLayer              PDULayer
	defaultTransferSyntax string
}

// SendResponse implements ResponseSender interface
func (r *responseHandler) SendResponse(msg *types.Message, dataset *dicom.Dataset, transferSyntaxUID string) error {
	tsUID := transferSyntaxUID
	if tsUID == "" {
		tsUID = r.defaultTransferSyntax
	}

	var datasetBytes []byte
	var err error
	if dataset != nil {
		datasetBytes, err = dicom.EncodeDatasetWithTransferSyntax(dataset, tsUID)
		if err != nil {
			return fmt.Errorf("failed to encode dataset with transfer syntax %s: %w", tsUID, err)
		}
	}

	// Propagate transfer syntax to message for downstream consumers
	msg.TransferSyntaxUID = tsUID

	return r.service.sendDIMSEResponse(msg, datasetBytes, r.presContextID, r.pduLayer)
}

// cGetResponder implements CGetResponder for C-GET operations
type cGetResponder struct {
	responseHandler
	messageIDCounter uint16
}

// SendCStore implements CGetResponder interface - sends C-STORE sub-operation on same association
func (c *cGetResponder) SendCStore(sopClassUID, sopInstanceUID string, data []byte) error {
	c.messageIDCounter++

	// Build C-STORE-RQ command
	command := &types.Message{
		CommandField:           CStoreRQ,
		MessageID:              c.messageIDCounter,
		Priority:               0x0002, // Medium priority
		AffectedSOPClassUID:    sopClassUID,
		AffectedSOPInstanceUID: sopInstanceUID,
		CommandDataSetType:     0x0000, // Dataset present
	}

	commandData := c.service.createDIMSECommand(command)

	// Send C-STORE-RQ with dataset on the same association
	if err := c.pduLayer.SendDIMSEResponseWithDataset(c.presContextID, commandData, data); err != nil {
		return fmt.Errorf("failed to send C-STORE sub-operation: %w", err)
	}

	// Note: In a full implementation, we should wait for C-STORE-RSP
	// For now, we'll assume success
	return nil
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
	tsUID, err := pduLayer.GetTransferSyntax(presContextID)
	if err != nil {
		d.logger.Warn("Failed to retrieve transfer syntax for presentation context",
			"context_id", presContextID,
			"error", err)
	}
	if tsUID != "" {
		d.transferUID = tsUID
	}
	d.contextID = presContextID

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

	tsUID := d.transferUID
	if tsUID == "" {
		if negotiatedTS, err := pduLayer.GetTransferSyntax(presContextID); err == nil {
			tsUID = negotiatedTS
		} else {
			d.logger.WarnContext(ctx, "Unable to determine transfer syntax for presentation context",
				"context_id", presContextID,
				"error", err)
		}
	}
	d.currentMsg.TransferSyntaxUID = tsUID

	var parsedDataset *dicom.Dataset
	if len(d.datasetData) > 0 {
		var err error
		parsedDataset, err = dicom.ParseDatasetWithTransferSyntax(d.datasetData, tsUID)
		if err != nil {
			d.logger.WarnContext(ctx, "Failed to parse dataset with negotiated transfer syntax",
				"transfer_syntax", tsUID,
				"error", err)
		} else {
			d.logger.DebugContext(ctx, "Parsed dataset using transfer syntax",
				"transfer_syntax", tsUID)
		}
	}

	meta := interfaces.MessageContext{
		PresentationContextID: presContextID,
		TransferSyntaxUID:     tsUID,
		Dataset:               parsedDataset,
	}

	defer d.resetState()

	if streamingHandler, ok := d.handler.(interfaces.StreamingServiceHandler); ok {
		d.logger.DebugContext(ctx, "Using streaming handler for multi-response operation")

		responder := d.buildResponder(presContextID, pduLayer, tsUID)
		return streamingHandler.HandleDIMSEStreaming(ctx, d.currentMsg, d.datasetData, meta, responder)
	}

	responseMsg, responseDataset, err := d.handler.HandleDIMSE(ctx, d.currentMsg, d.datasetData, meta)
	if err != nil {
		return fmt.Errorf("service handler failed: %w", err)
	}

	responseTS := responseMsg.TransferSyntaxUID
	if responseTS == "" {
		responseTS = tsUID
	}

	var encodedDataset []byte
	if responseDataset != nil {
		var encodeErr error
		encodedDataset, encodeErr = dicom.EncodeDatasetWithTransferSyntax(responseDataset, responseTS)
		if encodeErr != nil {
			return fmt.Errorf("failed to encode response dataset using transfer syntax %s: %w", responseTS, encodeErr)
		}
	}

	responseMsg.TransferSyntaxUID = responseTS
	return d.sendDIMSEResponse(responseMsg, encodedDataset, presContextID, pduLayer)
}

func (d *Service) buildResponder(presContextID byte, pduLayer PDULayer, defaultTS string) interfaces.ResponseSender {
	base := responseHandler{
		service:               d,
		presContextID:         presContextID,
		pduLayer:              pduLayer,
		defaultTransferSyntax: defaultTS,
	}

	if d.currentMsg != nil && d.currentMsg.CommandField == CGetRQ {
		return &cGetResponder{responseHandler: base}
	}

	return &base
}

func (d *Service) resetState() {
	d.commandData = nil
	d.datasetData = nil
	d.currentMsg = nil
	d.transferUID = ""
	d.contextID = 0
}

// sendDIMSEResponse sends a DIMSE response
func (d *Service) sendDIMSEResponse(msg *types.Message, data []byte, presContextID byte, pduLayer PDULayer) error {
	commandData := d.createDIMSECommand(msg)
	return pduLayer.SendDIMSEResponseWithDataset(presContextID, commandData, data)
}

// createDIMSECommand creates a DIMSE command dataset
func (d *Service) createDIMSECommand(msg *types.Message) []byte {
	var elements []byte

	// Affected SOP Class UID (0000,0002)
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

	// Message ID (0000,0110) - for requests
	if msg.MessageID > 0 && msg.MessageIDBeingRespondedTo == 0 {
		elements = append(elements, 0x00, 0x00, 0x10, 0x01) // Tag
		elements = append(elements, 0x02, 0x00, 0x00, 0x00) // Length = 2
		msgID := make([]byte, 2)
		binary.LittleEndian.PutUint16(msgID, msg.MessageID)
		elements = append(elements, msgID...)
	}

	// Message ID Being Responded To (0000,0120)
	if msg.MessageIDBeingRespondedTo > 0 {
		elements = append(elements, 0x00, 0x00, 0x20, 0x01) // Tag
		elements = append(elements, 0x02, 0x00, 0x00, 0x00) // Length = 2
		msgID := make([]byte, 2)
		binary.LittleEndian.PutUint16(msgID, msg.MessageIDBeingRespondedTo)
		elements = append(elements, msgID...)
	}

	// Affected SOP Instance UID (0000,1000) - for C-STORE
	if msg.AffectedSOPInstanceUID != "" {
		sopInstanceUID := msg.AffectedSOPInstanceUID
		if len(sopInstanceUID)%2 == 1 {
			sopInstanceUID += "\x00"
		}
		elements = append(elements, 0x00, 0x00, 0x00, 0x10) // Tag
		sopInstLen := make([]byte, 4)
		binary.LittleEndian.PutUint32(sopInstLen, uint32(len(sopInstanceUID)))
		elements = append(elements, sopInstLen...)
		elements = append(elements, []byte(sopInstanceUID)...)
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

	// C-MOVE response counters (optional, only for C-MOVE-RSP)
	if msg.NumberOfRemainingSuboperations != nil {
		// Number of Remaining Sub-operations (0000,1020)
		elements = append(elements, 0x00, 0x00, 0x20, 0x10) // Tag
		elements = append(elements, 0x02, 0x00, 0x00, 0x00) // Length = 2
		remaining := make([]byte, 2)
		binary.LittleEndian.PutUint16(remaining, *msg.NumberOfRemainingSuboperations)
		elements = append(elements, remaining...)
	}

	if msg.NumberOfCompletedSuboperations != nil {
		// Number of Completed Sub-operations (0000,1021)
		elements = append(elements, 0x00, 0x00, 0x21, 0x10) // Tag
		elements = append(elements, 0x02, 0x00, 0x00, 0x00) // Length = 2
		completed := make([]byte, 2)
		binary.LittleEndian.PutUint16(completed, *msg.NumberOfCompletedSuboperations)
		elements = append(elements, completed...)
	}

	if msg.NumberOfFailedSuboperations != nil {
		// Number of Failed Sub-operations (0000,1022)
		elements = append(elements, 0x00, 0x00, 0x22, 0x10) // Tag
		elements = append(elements, 0x02, 0x00, 0x00, 0x00) // Length = 2
		failed := make([]byte, 2)
		binary.LittleEndian.PutUint16(failed, *msg.NumberOfFailedSuboperations)
		elements = append(elements, failed...)
	}

	if msg.NumberOfWarningSuboperations != nil {
		// Number of Warning Sub-operations (0000,1023)
		elements = append(elements, 0x00, 0x00, 0x23, 0x10) // Tag
		elements = append(elements, 0x02, 0x00, 0x00, 0x00) // Length = 2
		warning := make([]byte, 2)
		binary.LittleEndian.PutUint16(warning, *msg.NumberOfWarningSuboperations)
		elements = append(elements, warning...)
	}

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
