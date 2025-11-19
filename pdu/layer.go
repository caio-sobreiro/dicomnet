package pdu

import (
	"encoding/binary"
	"fmt"
	"io"
	"log/slog"
	"net"
	"strings"

	"github.com/caio-sobreiro/dicomnet/types"
)

// PDU types
const (
	TypeAssociateRQ = 0x01
	TypeAssociateAC = 0x02
	TypeAssociateRJ = 0x03
	TypePDataTF     = 0x04
	TypeReleaseRQ   = 0x05
	TypeReleaseRP   = 0x06
	TypeAbort       = 0x07
)

// PDU represents a Protocol Data Unit
type PDU struct {
	Type   byte
	Length uint32
	Data   []byte
}

// Layer handles the DICOM Upper Layer Protocol
type Layer struct {
	conn           net.Conn
	associationCtx *AssociationContext
	dimseHandler   DIMSEHandler
	serverAETitle  string
	logger         *slog.Logger
}

// AssociationContext holds association state
type AssociationContext struct {
	CalledAETitle    string
	CallingAETitle   string
	MaxPDULength     uint32
	PresentationCtxs map[byte]*PresentationContext
}

// PresentationContext represents a negotiated presentation context
type PresentationContext struct {
	ID             byte
	Result         byte
	AbstractSyntax string
	TransferSyntax string
}

const (
	presentationResultAcceptance           byte = 0x00
	presentationResultRejectAbstractSyntax byte = 0x03
	presentationResultRejectTransferSyntax byte = 0x04
)

var supportedAbstractSyntaxes = map[string]bool{
	types.VerificationSOPClass:                              true, // Verification SOP Class (C-ECHO)
	types.PatientRootQueryRetrieveInformationModelFind:      true, // Patient Root Q/R - FIND
	types.StudyRootQueryRetrieveInformationModelFind:        true, // Study Root Q/R - FIND
	types.PatientStudyOnlyQueryRetrieveInformationModelFind: true, // Patient/Study Only Q/R - FIND
	types.PatientRootQueryRetrieveInformationModelMove:      true, // Patient Root Q/R - MOVE
	types.StudyRootQueryRetrieveInformationModelMove:        true, // Study Root Q/R - MOVE
	types.PatientStudyOnlyQueryRetrieveInformationModelMove: true, // Patient/Study Only Q/R - MOVE
	types.PatientRootQueryRetrieveInformationModelGet:       true, // Patient Root Q/R - GET
	types.StudyRootQueryRetrieveInformationModelGet:         true, // Study Root Q/R - GET
	types.PatientStudyOnlyQueryRetrieveInformationModelGet:  true, // Patient/Study Only Q/R - GET
}

var supportedTransferSyntaxes = map[string]bool{
	types.ImplicitVRLittleEndian: true, // Implicit VR Little Endian
	types.ExplicitVRLittleEndian: true, // Explicit VR Little Endian
}

func normalizeUID(raw []byte) string {
	value := string(raw)
	value = strings.TrimRight(value, "\x00 ")
	return value
}

func supportsAbstractSyntax(uid string) bool {
	if supportedAbstractSyntaxes[uid] {
		return true
	}
	// Accept all storage SOP classes (C-STORE)
	if types.IsStorageSOPClass(uid) {
		return true
	}
	return false
}

func supportsTransferSyntax(uid string) bool {
	return supportedTransferSyntaxes[uid]
}

func parsePresentationContext(data []byte, logger *slog.Logger) (*PresentationContext, error) {
	if len(data) < 4 {
		return nil, fmt.Errorf("presentation context too short: %d", len(data))
	}

	ctxID := data[0]
	subOffset := 4 // Skip reserved bytes
	var abstractSyntax string
	var transferSyntaxes []string

	for subOffset+4 <= len(data) {
		subItemType := data[subOffset]
		subItemLength := binary.BigEndian.Uint16(data[subOffset+2 : subOffset+4])
		valueStart := subOffset + 4
		valueEnd := valueStart + int(subItemLength)
		if valueEnd > len(data) {
			return nil, fmt.Errorf("presentation context %d sub-item exceeds length", ctxID)
		}

		value := data[valueStart:valueEnd]
		switch subItemType {
		case 0x30: // Abstract Syntax
			abstractSyntax = normalizeUID(value)
		case 0x40: // Transfer Syntax
			transferSyntaxes = append(transferSyntaxes, normalizeUID(value))
		}

		subOffset = valueEnd
	}

	if abstractSyntax == "" {
		return nil, fmt.Errorf("presentation context %d missing abstract syntax", ctxID)
	}

	if logger != nil {
		logger.Debug("Parsing presentation context",
			"context_id", ctxID,
			"abstract_syntax", abstractSyntax,
			"proposed_transfer_syntaxes", transferSyntaxes,
			"num_proposed", len(transferSyntaxes))
	}

	result := presentationResultRejectAbstractSyntax
	selectedTransfer := ""

	if supportsAbstractSyntax(abstractSyntax) {
		for _, ts := range transferSyntaxes {
			if supportsTransferSyntax(ts) {
				selectedTransfer = ts
				result = presentationResultAcceptance
				break
			}
		}
		if result != presentationResultAcceptance {
			result = presentationResultRejectTransferSyntax
		}
	}

	if logger != nil {
		logger.Debug("Presentation context negotiation result",
			"context_id", ctxID,
			"abstract_syntax", abstractSyntax,
			"selected_transfer_syntax", selectedTransfer,
			"result", result)
	}

	// Validation: accepted contexts MUST have a transfer syntax
	if result == presentationResultAcceptance && selectedTransfer == "" {
		// This should never happen - it means we accepted but didn't select a transfer syntax
		// Force rejection to avoid protocol violation
		result = presentationResultRejectTransferSyntax
	}

	return &PresentationContext{
		ID:             ctxID,
		Result:         result,
		AbstractSyntax: abstractSyntax,
		TransferSyntax: selectedTransfer,
	}, nil
}

func parseUserInformation(data []byte) (uint32, error) {
	offset := 0
	var maxPDULength uint32

	for offset+4 <= len(data) {
		subItemType := data[offset]
		subItemLength := binary.BigEndian.Uint16(data[offset+2 : offset+4])
		valueStart := offset + 4
		valueEnd := valueStart + int(subItemLength)
		if valueEnd > len(data) {
			return 0, fmt.Errorf("user information sub-item exceeds length")
		}

		if subItemType == 0x51 && subItemLength == 4 {
			maxPDULength = binary.BigEndian.Uint32(data[valueStart:valueEnd])
		}

		offset = valueEnd
	}

	return maxPDULength, nil
}

// DIMSEHandler interface for handling DIMSE messages
type DIMSEHandler interface {
	HandleDIMSEMessage(presContextID byte, msgCtrlHeader byte, data []byte, pduLayer *Layer) error
}

// NewLayer creates a new PDU layer handler
func NewLayer(conn net.Conn, dimseHandler DIMSEHandler, serverAETitle string, logger *slog.Logger) *Layer {
	if logger == nil {
		logger = slog.Default()
	}
	return &Layer{
		conn:          conn,
		dimseHandler:  dimseHandler,
		serverAETitle: serverAETitle,
		logger:        logger,
	}
}

// HandleConnection manages the complete DICOM connection lifecycle
func (p *Layer) HandleConnection() error {
	defer p.conn.Close()
	p.logger.Info("New DICOM connection", "remote_addr", p.conn.RemoteAddr())

	// Handle association establishment
	if err := p.handleAssociationPhase(); err != nil {
		return fmt.Errorf("association failed: %v", err)
	}

	// Handle DIMSE messages
	for {
		pdu, err := p.readPDU()
		if err != nil {
			if err == io.EOF {
				p.logger.Info("Connection closed by client", "remote_addr", p.conn.RemoteAddr())
			} else {
				p.logger.Warn("Error reading PDU", "error", err, "remote_addr", p.conn.RemoteAddr())
			}
			break
		}

		if err := p.handlePDU(pdu); err != nil {
			if err == io.EOF {
				break // Normal termination
			}
			return fmt.Errorf("error handling PDU: %v", err)
		}
	}

	return nil
}

// readPDU reads a complete PDU from the connection
func (p *Layer) readPDU() (*PDU, error) {
	// Read PDU header (6 bytes)
	header := make([]byte, 6)
	if _, err := io.ReadFull(p.conn, header); err != nil {
		return nil, err
	}

	pduType := header[0]
	pduLength := binary.BigEndian.Uint32(header[2:6])

	// Read PDU data
	pduData := make([]byte, pduLength)
	if _, err := io.ReadFull(p.conn, pduData); err != nil {
		return nil, fmt.Errorf("failed to read PDU data: %v", err)
	}

	return &PDU{
		Type:   pduType,
		Length: pduLength,
		Data:   pduData,
	}, nil
}

// handlePDU routes PDUs to appropriate handlers
func (p *Layer) handlePDU(pdu *PDU) error {
	p.logger.Debug("Received PDU", "type", fmt.Sprintf("0x%02x", pdu.Type), "length", pdu.Length)

	switch pdu.Type {
	case TypePDataTF:
		return p.handlePDataTF(pdu)
	case TypeReleaseRQ:
		return p.handleReleaseRequest()
	case TypeReleaseRP:
		p.logger.Debug("Received A-RELEASE-RP")
		return io.EOF
	case TypeAbort:
		p.logger.Info("Received A-ABORT")
		return io.EOF
	default:
		p.logger.Warn("Unhandled PDU type", "type", fmt.Sprintf("0x%02x", pdu.Type))
		return nil
	}
}

// handleAssociationPhase handles the association establishment
func (p *Layer) handleAssociationPhase() error {
	pdu, err := p.readPDU()
	if err != nil {
		return fmt.Errorf("failed to read association request: %v", err)
	}

	if pdu.Type != TypeAssociateRQ {
		return fmt.Errorf("expected A-ASSOCIATE-RQ, got PDU type: 0x%02x", pdu.Type)
	}

	return p.handleAssociateRequest(pdu)
}

// handleAssociateRequest processes A-ASSOCIATE-RQ and sends A-ASSOCIATE-AC
func (p *Layer) handleAssociateRequest(pdu *PDU) error {
	p.logger.Debug("Processing A-ASSOCIATE-RQ")

	// Initialize association context with default values (will be updated by parsing)
	p.associationCtx = &AssociationContext{
		CalledAETitle:    p.serverAETitle, // Use configured server AE title
		CallingAETitle:   "UNKNOWN",       // Default, will be updated from request
		MaxPDULength:     16384,
		PresentationCtxs: make(map[byte]*PresentationContext),
	}

	// Parse the incoming association request to get the presentation contexts
	if err := p.parseAssociationRequest(pdu); err != nil {
		p.logger.Debug("Using default presentation contexts", "reason", err)
		// Fall back to accepting common contexts
	}

	// If no contexts were parsed, add default supported contexts
	if len(p.associationCtx.PresentationCtxs) == 0 {
		p.addDefaultPresentationContexts()
	}

	// Send A-ASSOCIATE-AC
	response := p.createAssociateAccept()
	if _, err := p.conn.Write(response); err != nil {
		return fmt.Errorf("failed to send A-ASSOCIATE-AC: %v", err)
	}

	p.logger.Debug("Sent A-ASSOCIATE-AC")
	return nil
}

// handlePDataTF processes P-DATA-TF PDUs and forwards to DIMSE layer
func (p *Layer) handlePDataTF(pdu *PDU) error {
	p.logger.Debug("Processing P-DATA-TF")

	// Extract PDV from P-DATA-TF
	if len(pdu.Data) < 6 {
		return fmt.Errorf("P-DATA-TF too short")
	}

	// Parse PDV
	pdvLength := binary.BigEndian.Uint32(pdu.Data[0:4])
	if len(pdu.Data) < int(4+pdvLength) {
		return fmt.Errorf("incomplete PDV data")
	}

	pdvData := pdu.Data[4 : 4+pdvLength]
	if len(pdvData) < 2 {
		return fmt.Errorf("PDV data too short")
	}

	presContextID := pdvData[0]
	msgCtrlHeader := pdvData[1]
	dimseData := pdvData[2:]

	p.logger.Debug("Processing DIMSE message",
		"presentation_context_id", presContextID,
		"message_control_header", fmt.Sprintf("0x%02x", msgCtrlHeader))

	// Forward to DIMSE layer
	return p.dimseHandler.HandleDIMSEMessage(presContextID, msgCtrlHeader, dimseData, p)
}

// handleReleaseRequest processes A-RELEASE-RQ and sends A-RELEASE-RP
func (p *Layer) handleReleaseRequest() error {
	p.logger.Debug("Processing A-RELEASE-RQ")

	// Send A-RELEASE-RP
	response := []byte{0x06, 0x00, 0x00, 0x00, 0x00, 0x04, 0x00, 0x00, 0x00, 0x00}

	if _, err := p.conn.Write(response); err != nil {
		return fmt.Errorf("failed to send A-RELEASE-RP: %v", err)
	}

	p.logger.Debug("Sent A-RELEASE-RP")
	return io.EOF
}

// SendDIMSEResponse sends a DIMSE response via P-DATA-TF
func (p *Layer) SendDIMSEResponse(presContextID byte, commandData []byte) error {
	return p.SendDIMSEResponseWithDataset(presContextID, commandData, nil)
}

// SendDIMSEResponseWithDataset sends a DIMSE response with optional dataset via P-DATA-TF
func (p *Layer) SendDIMSEResponseWithDataset(presContextID byte, commandData []byte, datasetData []byte) error {
	// First, send the command PDV as a separate P-DATA-TF PDU
	commandPDVHeader := []byte{presContextID, 0x03} // Message Control Header = 0x03 (command, last fragment)
	commandPDVData := append(commandPDVHeader, commandData...)

	// PDV Length for command
	commandPDVLength := make([]byte, 4)
	binary.BigEndian.PutUint32(commandPDVLength, uint32(len(commandPDVData)))

	// Create command P-DATA-TF PDU
	commandPDUHeader := []byte{TypePDataTF, 0x00} // P-DATA-TF PDU type
	commandPDULength := make([]byte, 4)
	binary.BigEndian.PutUint32(commandPDULength, uint32(len(commandPDVLength)+len(commandPDVData)))

	// Assemble command PDU: PDU header + PDU length + command PDV
	commandResponse := append(commandPDUHeader, commandPDULength...)
	commandResponse = append(commandResponse, commandPDVLength...)
	commandResponse = append(commandResponse, commandPDVData...)

	// Send command PDU
	if _, err := p.conn.Write(commandResponse); err != nil {
		return fmt.Errorf("failed to send command PDU: %v", err)
	}

	// If there's dataset data, send it as a separate P-DATA-TF PDU
	if len(datasetData) > 0 {
		datasetPDVHeader := []byte{presContextID, 0x02} // Message Control Header = 0x02 (dataset, last fragment)
		datasetPDVData := append(datasetPDVHeader, datasetData...)

		// PDV Length for dataset
		datasetPDVLength := make([]byte, 4)
		binary.BigEndian.PutUint32(datasetPDVLength, uint32(len(datasetPDVData)))

		// Create dataset P-DATA-TF PDU
		datasetPDUHeader := []byte{TypePDataTF, 0x00} // P-DATA-TF PDU type
		datasetPDULength := make([]byte, 4)
		binary.BigEndian.PutUint32(datasetPDULength, uint32(len(datasetPDVLength)+len(datasetPDVData)))

		// Assemble dataset PDU: PDU header + PDU length + dataset PDV
		datasetResponse := append(datasetPDUHeader, datasetPDULength...)
		datasetResponse = append(datasetResponse, datasetPDVLength...)
		datasetResponse = append(datasetResponse, datasetPDVData...)

		// Send dataset PDU
		if _, err := p.conn.Write(datasetResponse); err != nil {
			return fmt.Errorf("failed to send dataset PDU: %v", err)
		}
	}

	return nil
}

// GetTransferSyntax returns the negotiated transfer syntax for the given presentation context.
func (p *Layer) GetTransferSyntax(presContextID byte) (string, error) {
	if p.associationCtx == nil {
		return "", fmt.Errorf("association context not initialized")
	}

	ctx, ok := p.associationCtx.PresentationCtxs[presContextID]
	if !ok {
		return "", fmt.Errorf("presentation context %d not found", presContextID)
	}

	if ctx.TransferSyntax == "" {
		return "", fmt.Errorf("no transfer syntax negotiated for presentation context %d", presContextID)
	}

	return ctx.TransferSyntax, nil
}

// createAssociateAccept creates a proper A-ASSOCIATE-AC PDU
func (p *Layer) createAssociateAccept() []byte {
	// Fixed fields (68 bytes)
	fixedFields := make([]byte, 68)

	// Protocol version (bytes 0-1): 0x0001
	binary.BigEndian.PutUint16(fixedFields[0:2], 0x0001)

	// Use the AE titles from the association context (extracted from request)
	calledAE := p.associationCtx.CalledAETitle
	if len(calledAE) > 16 {
		calledAE = calledAE[:16]
	}
	callingAE := p.associationCtx.CallingAETitle
	if len(callingAE) > 16 {
		callingAE = callingAE[:16]
	}

	// Copy AE titles (pad with spaces to 16 bytes each)
	copy(fixedFields[4:20], fmt.Sprintf("%-16s", calledAE))   // Called AE Title
	copy(fixedFields[20:36], fmt.Sprintf("%-16s", callingAE)) // Calling AE Title

	// Application Context Item
	appContextUID := types.ApplicationContextUID
	appContextItem := []byte{0x10, 0x00} // Item type
	appContextLen := make([]byte, 2)
	binary.BigEndian.PutUint16(appContextLen, uint16(len(appContextUID)))
	appContextItem = append(appContextItem, appContextLen...)
	appContextItem = append(appContextItem, []byte(appContextUID)...)

	// Build all presentation contexts
	// Sort context IDs to ensure consistent ordering
	var contextIDs []byte
	for id := range p.associationCtx.PresentationCtxs {
		contextIDs = append(contextIDs, id)
	}
	// Simple bubble sort since we have few contexts
	for i := 0; i < len(contextIDs); i++ {
		for j := i + 1; j < len(contextIDs); j++ {
			if contextIDs[i] > contextIDs[j] {
				contextIDs[i], contextIDs[j] = contextIDs[j], contextIDs[i]
			}
		}
	}

	var allPresContextItems []byte
	for _, id := range contextIDs {
		ctx := p.associationCtx.PresentationCtxs[id]

		// WORKAROUND: Some DICOM implementations (e.g., DCMTK/Orthanc) incorrectly reject
		// A-ASSOCIATE-AC PDUs that include rejected presentation contexts, even though
		// DICOM PS3.8 Section 9.3.3.3 requires including all contexts from the RQ.
		// Skip rejected contexts to maintain compatibility.
		if ctx.Result != presentationResultAcceptance {
			p.logger.Debug("Skipping rejected context (compatibility workaround)",
				"context_id", ctx.ID,
				"result", ctx.Result)
			continue
		}

		var presContextData []byte

		// According to DICOM Part 8, Section 9.3.3.3:
		// - For accepted contexts (Result == 0x00): include ONLY Transfer Syntax
		// - For rejected contexts (Result != 0x00): include NO sub-items
		if ctx.Result == presentationResultAcceptance {
			// CRITICAL: Accepted contexts MUST have a transfer syntax
			if ctx.TransferSyntax == "" {
				p.logger.Error("Accepted presentation context missing transfer syntax",
					"context_id", ctx.ID,
					"abstract_syntax", ctx.AbstractSyntax)
				// This should never happen - reject the context instead
				ctx.Result = presentationResultRejectTransferSyntax
			} else {
				// Transfer Syntax only for accepted contexts
				transferSyntaxItem := []byte{0x40, 0x00} // Item type
				transferSyntaxLen := make([]byte, 2)
				binary.BigEndian.PutUint16(transferSyntaxLen, uint16(len(ctx.TransferSyntax)))
				transferSyntaxItem = append(transferSyntaxItem, transferSyntaxLen...)
				transferSyntaxItem = append(transferSyntaxItem, []byte(ctx.TransferSyntax)...)
				presContextData = transferSyntaxItem
			}
		}
		// For rejected contexts, presContextData remains empty (no sub-items)

		// Build this presentation context
		presContextItem := []byte{0x21, 0x00} // Item type (0x21 = Presentation Context Item - AC)
		presContextLen := make([]byte, 2)
		binary.BigEndian.PutUint16(presContextLen, uint16(4+len(presContextData)))
		presContextItem = append(presContextItem, presContextLen...)
		presContextItem = append(presContextItem, ctx.ID, ctx.Result, 0x00, 0x00)
		presContextItem = append(presContextItem, presContextData...)

		allPresContextItems = append(allPresContextItems, presContextItem...)
	}

	// User Information Item
	maxPDUItem := []byte{0x51, 0x00, 0x00, 0x04}
	maxPDUValue := make([]byte, 4)
	binary.BigEndian.PutUint32(maxPDUValue, 16384)
	maxPDUItem = append(maxPDUItem, maxPDUValue...)

	implClassUID := "1.2.3.4.5.6.7.8.9"
	implClassItem := []byte{0x52, 0x00}
	implClassLen := make([]byte, 2)
	binary.BigEndian.PutUint16(implClassLen, uint16(len(implClassUID)))
	implClassItem = append(implClassItem, implClassLen...)
	implClassItem = append(implClassItem, []byte(implClassUID)...)

	implVersionName := "DIMSE_GO_1.0"
	implVersionItem := []byte{0x55, 0x00}
	implVersionLen := make([]byte, 2)
	binary.BigEndian.PutUint16(implVersionLen, uint16(len(implVersionName)))
	implVersionItem = append(implVersionItem, implVersionLen...)
	implVersionItem = append(implVersionItem, []byte(implVersionName)...)

	userInfoData := append(maxPDUItem, implClassItem...)
	userInfoData = append(userInfoData, implVersionItem...)
	userInfoItem := []byte{0x50, 0x00}
	userInfoLen := make([]byte, 2)
	binary.BigEndian.PutUint16(userInfoLen, uint16(len(userInfoData)))
	userInfoItem = append(userInfoItem, userInfoLen...)
	userInfoItem = append(userInfoItem, userInfoData...)

	// Combine all
	variableItems := append(appContextItem, allPresContextItems...)
	variableItems = append(variableItems, userInfoItem...)
	pduData := append(fixedFields, variableItems...)

	// Create PDU header
	pduHeader := []byte{TypeAssociateAC, 0x00}
	pduLength := make([]byte, 4)
	binary.BigEndian.PutUint32(pduLength, uint32(len(pduData)))
	pduHeader = append(pduHeader, pduLength...)

	return append(pduHeader, pduData...)
}

// parseAssociationRequest parses an A-ASSOCIATE-RQ PDU to extract presentation contexts and AE titles
func (p *Layer) parseAssociationRequest(pdu *PDU) error {
	p.logger.Debug("Parsing association request", "pdu_length", len(pdu.Data))

	if len(pdu.Data) < 68 { // Minimum size for a basic association request
		return fmt.Errorf("association request too short")
	}

	data := pdu.Data

	// Extract AE titles from fixed fields (bytes 4-36)
	// Called AE Title (bytes 4-19) - what they're calling us
	calledAEBytes := data[4:20]
	calledAE := string(calledAEBytes)
	if idx := strings.IndexByte(calledAE, 0); idx != -1 {
		calledAE = calledAE[:idx]
	}
	calledAE = strings.TrimSpace(calledAE)

	// Calling AE Title (bytes 20-35) - who is calling us
	callingAEBytes := data[20:36]
	callingAE := string(callingAEBytes)
	if idx := strings.IndexByte(callingAE, 0); idx != -1 {
		callingAE = callingAE[:idx]
	}
	callingAE = strings.TrimSpace(callingAE)

	// Update association context with extracted AE titles
	if p.associationCtx != nil {
		p.associationCtx.CalledAETitle = calledAE
		p.associationCtx.CallingAETitle = callingAE
		p.associationCtx.PresentationCtxs = make(map[byte]*PresentationContext)
	}

	p.logger.Info("Extracted AE titles from association request",
		"calling_ae", callingAE,
		"called_ae", calledAE)

	// Parse variable items starting from offset 68
	offset := 68
	var proposedContexts int
	var acceptedContexts int

	// Parse variable items
	for offset < len(data) {
		if offset+4 > len(data) {
			break
		}

		itemType := data[offset]
		// Skip reserved byte
		itemLength := binary.BigEndian.Uint16(data[offset+2 : offset+4])
		valueStart := offset + 4
		valueEnd := valueStart + int(itemLength)
		if valueEnd > len(data) {
			return fmt.Errorf("association item exceeds PDU length")
		}
		itemData := data[valueStart:valueEnd]

		p.logger.Debug("Found association item", "type", fmt.Sprintf("0x%02x", itemType), "length", itemLength)

		switch itemType {
		case 0x10: // Application Context
			p.logger.Debug("Found application context item")
		case 0x20: // Presentation Context
			p.logger.Debug("Found presentation context item")
			proposedContexts++
			ctx, err := parsePresentationContext(itemData, p.logger)
			if err != nil {
				p.logger.Warn("Failed to parse presentation context", "error", err)
			} else if p.associationCtx != nil {
				p.associationCtx.PresentationCtxs[ctx.ID] = ctx
				if ctx.Result == presentationResultAcceptance {
					acceptedContexts++
				}
			}
		case 0x50: // User Information
			p.logger.Debug("Found user information item")
			if maxPDULength, err := parseUserInformation(itemData); err != nil {
				p.logger.Warn("Failed to parse user information", "error", err)
			} else if maxPDULength > 0 && p.associationCtx != nil {
				p.associationCtx.MaxPDULength = maxPDULength
			}
		}

		offset = valueEnd
	}

	if proposedContexts == 0 {
		p.logger.Warn("No presentation contexts found in association request")
	} else {
		p.logger.Info("Negotiated presentation contexts",
			"proposed", proposedContexts,
			"accepted", acceptedContexts,
			"max_pdu_length", p.associationCtx.MaxPDULength)
	}

	return nil
}

// addDefaultPresentationContexts adds the standard presentation contexts
func (p *Layer) addDefaultPresentationContexts() {
	p.logger.Debug("Adding default presentation contexts")

	// Verification SOP Class (C-ECHO)
	p.associationCtx.PresentationCtxs[1] = &PresentationContext{
		ID:             1,
		Result:         0,                   // Acceptance
		AbstractSyntax: "1.2.840.10008.1.1", // Verification SOP Class
		TransferSyntax: "1.2.840.10008.1.2", // Implicit VR Little Endian
	}

	// Patient Root Query/Retrieve Information Model - FIND
	p.associationCtx.PresentationCtxs[3] = &PresentationContext{
		ID:             3,
		Result:         0,                             // Acceptance
		AbstractSyntax: "1.2.840.10008.5.1.4.1.2.1.1", // Patient Root Q/R - FIND
		TransferSyntax: "1.2.840.10008.1.2",           // Implicit VR Little Endian
	}

	// Study Root Query/Retrieve Information Model - FIND
	p.associationCtx.PresentationCtxs[5] = &PresentationContext{
		ID:             5,
		Result:         0,                             // Acceptance
		AbstractSyntax: "1.2.840.10008.5.1.4.1.2.2.1", // Study Root Q/R - FIND
		TransferSyntax: "1.2.840.10008.1.2",           // Implicit VR Little Endian
	}

	// Patient/Study Only Query/Retrieve Information Model - FIND
	p.associationCtx.PresentationCtxs[7] = &PresentationContext{
		ID:             7,
		Result:         0,                             // Acceptance
		AbstractSyntax: "1.2.840.10008.5.1.4.1.2.3.1", // Patient/Study Only Q/R - FIND
		TransferSyntax: "1.2.840.10008.1.2",           // Implicit VR Little Endian
	}

	// Patient Root Query/Retrieve Information Model - MOVE
	p.associationCtx.PresentationCtxs[9] = &PresentationContext{
		ID:             9,
		Result:         0,
		AbstractSyntax: "1.2.840.10008.5.1.4.1.2.1.2", // Patient Root Q/R - MOVE
		TransferSyntax: "1.2.840.10008.1.2",           // Implicit VR Little Endian
	}

	// Study Root Query/Retrieve Information Model - MOVE
	p.associationCtx.PresentationCtxs[11] = &PresentationContext{
		ID:             11,
		Result:         0,
		AbstractSyntax: "1.2.840.10008.5.1.4.1.2.2.2", // Study Root Q/R - MOVE
		TransferSyntax: "1.2.840.10008.1.2",           // Implicit VR Little Endian
	}

	// Patient/Study Only Query/Retrieve Information Model - MOVE
	p.associationCtx.PresentationCtxs[13] = &PresentationContext{
		ID:             13,
		Result:         0,
		AbstractSyntax: "1.2.840.10008.5.1.4.1.2.3.2", // Patient/Study Only Q/R - MOVE
		TransferSyntax: "1.2.840.10008.1.2",           // Implicit VR Little Endian
	}

	slog.Debug("Added presentation contexts", "count", len(p.associationCtx.PresentationCtxs))
}
