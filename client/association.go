package client

import (
	"encoding/binary"
	"fmt"
	"io"
	"log/slog"
	"net"
	"strings"
	"time"

	"github.com/caio-sobreiro/dicomnet/pdu"
)

// Association represents a client-side DICOM association
type Association struct {
	conn                      net.Conn
	callingAETitle            string
	calledAETitle             string
	maxPDULength              uint32
	presentationCtxs          map[byte]*PresentationContext
	logger                    *slog.Logger
	preferredTransferSyntaxes []string
}

// PresentationContext holds negotiated presentation context info
type PresentationContext struct {
	ID             byte
	AbstractSyntax string
	TransferSyntax string
	Accepted       bool
}

// Config holds client configuration
type Config struct {
	CallingAETitle            string
	CalledAETitle             string
	MaxPDULength              uint32
	ConnectTimeout            time.Duration // Timeout for establishing connection (default: 30s)
	ReadTimeout               time.Duration // Timeout for read operations (default: 60s)
	WriteTimeout              time.Duration // Timeout for write operations (default: 60s)
	Logger                    *slog.Logger  // Logger for the association (default: slog.Default())
	PreferredTransferSyntaxes []string      // Transfer syntaxes to propose (default: Explicit VR, Implicit VR)
}

// Connect establishes a DICOM association with a remote SCP
func Connect(address string, config Config) (*Association, error) {
	if config.MaxPDULength == 0 {
		config.MaxPDULength = 16384 // Default 16KB
	}
	if config.ConnectTimeout == 0 {
		config.ConnectTimeout = 30 * time.Second
	}
	if config.ReadTimeout == 0 {
		config.ReadTimeout = 60 * time.Second
	}
	if config.WriteTimeout == 0 {
		config.WriteTimeout = 60 * time.Second
	}

	// Establish TCP connection with timeout
	dialer := &net.Dialer{
		Timeout: config.ConnectTimeout,
	}
	conn, err := dialer.Dial("tcp", address)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	// Set initial read/write timeouts
	if err := conn.SetReadDeadline(time.Now().Add(config.ReadTimeout)); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to set read deadline: %w", err)
	}
	if err := conn.SetWriteDeadline(time.Now().Add(config.WriteTimeout)); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to set write deadline: %w", err)
	}

	// Set logger
	logger := config.Logger
	if logger == nil {
		logger = slog.Default()
	}

	// Set default transfer syntaxes if not provided
	transferSyntaxes := config.PreferredTransferSyntaxes
	if len(transferSyntaxes) == 0 {
		transferSyntaxes = []string{
			"1.2.840.10008.1.2.1", // Explicit VR Little Endian (default)
			"1.2.840.10008.1.2",   // Implicit VR Little Endian
		}
	}

	assoc := &Association{
		conn:                      conn,
		callingAETitle:            config.CallingAETitle,
		calledAETitle:             config.CalledAETitle,
		maxPDULength:              config.MaxPDULength,
		presentationCtxs:          make(map[byte]*PresentationContext),
		logger:                    logger,
		preferredTransferSyntaxes: transferSyntaxes,
	}

	// Send association request
	if err := assoc.sendAssociateRQ(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to send A-ASSOCIATE-RQ: %w", err)
	}

	// Wait for association accept
	if err := assoc.receiveAssociateAC(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to receive A-ASSOCIATE-AC: %w", err)
	}

	logger.Info("DICOM association established",
		"remote_addr", address,
		"calling_ae", config.CallingAETitle,
		"called_ae", config.CalledAETitle)

	return assoc, nil
}

// Close gracefully closes the association
func (a *Association) Close() error {
	// Send release request
	if err := a.sendReleaseRQ(); err != nil {
		a.logger.Warn("Failed to send release request", "error", err)
	}

	// Wait for release response (with timeout handled by TCP)
	a.receiveReleaseRP()

	return a.conn.Close()
}

// sendAssociateRQ sends an A-ASSOCIATE-RQ PDU
func (a *Association) sendAssociateRQ() error {
	// Build A-ASSOCIATE-RQ PDU
	// This is a simplified version - in production you'd want more presentation contexts
	buf := make([]byte, 0, 1024)

	// Protocol version (2 bytes) = 0x0001
	buf = append(buf, 0x00, 0x01)

	// Reserved (2 bytes)
	buf = append(buf, 0x00, 0x00)

	// Called AE Title (16 bytes, space-padded)
	calledAE := make([]byte, 16)
	copy(calledAE, a.calledAETitle)
	for i := len(a.calledAETitle); i < 16; i++ {
		calledAE[i] = ' '
	}
	buf = append(buf, calledAE...)

	// Calling AE Title (16 bytes, space-padded)
	callingAE := make([]byte, 16)
	copy(callingAE, a.callingAETitle)
	for i := len(a.callingAETitle); i < 16; i++ {
		callingAE[i] = ' '
	}
	buf = append(buf, callingAE...)

	// Reserved (32 bytes)
	buf = append(buf, make([]byte, 32)...)

	// Application Context Item
	buf = append(buf, 0x10)                               // Item type
	buf = append(buf, 0x00)                               // Reserved
	buf = append(buf, 0x00, 0x15)                         // Length
	buf = append(buf, []byte("1.2.840.10008.3.1.1.1")...) // Application Context UID

	// Presentation Context Item - CT Image Storage
	buf = a.addPresentationContext(buf, 1, "1.2.840.10008.5.1.4.1.1.2")

	// Presentation Context Item - MR Image Storage
	buf = a.addPresentationContext(buf, 3, "1.2.840.10008.5.1.4.1.1.4")

	// Presentation Context Item - Secondary Capture
	buf = a.addPresentationContext(buf, 5, "1.2.840.10008.5.1.4.1.1.7")

	// Presentation Context Item - Verification SOP Class (C-ECHO)
	buf = a.addPresentationContext(buf, 7, "1.2.840.10008.1.1")

	// Presentation Context Item - Study Root Query/Retrieve Information Model - FIND (C-FIND)
	buf = a.addPresentationContext(buf, 9, "1.2.840.10008.5.1.4.1.2.2.1")

	// User Information Item
	buf = a.addUserInformation(buf)

	// Write PDU header
	pduHeader := make([]byte, 6)
	pduHeader[0] = pdu.TypeAssociateRQ
	pduHeader[1] = 0x00 // Reserved
	binary.BigEndian.PutUint32(pduHeader[2:6], uint32(len(buf)))

	// Send PDU
	if _, err := a.conn.Write(pduHeader); err != nil {
		return err
	}
	if _, err := a.conn.Write(buf); err != nil {
		return err
	}

	return nil
}

// addPresentationContext adds a presentation context to the buffer
func (a *Association) addPresentationContext(buf []byte, contextID byte, abstractSyntax string) []byte {
	pcStart := len(buf)

	// Presentation Context Item
	buf = append(buf, 0x20)             // Item type
	buf = append(buf, 0x00)             // Reserved
	buf = append(buf, 0x00, 0x00)       // Length placeholder
	buf = append(buf, contextID)        // Presentation context ID
	buf = append(buf, 0x00, 0x00, 0x00) // Reserved

	// Abstract Syntax Sub-Item
	buf = append(buf, 0x30)                            // Item type
	buf = append(buf, 0x00)                            // Reserved
	buf = append(buf, 0x00, byte(len(abstractSyntax))) // Length
	buf = append(buf, []byte(abstractSyntax)...)

	// Transfer Syntax Sub-Items - use configured transfer syntaxes (order matters - first is preferred)
	for _, ts := range a.preferredTransferSyntaxes {
		buf = append(buf, 0x40)                // Item type
		buf = append(buf, 0x00)                // Reserved
		buf = append(buf, 0x00, byte(len(ts))) // Length
		buf = append(buf, []byte(ts)...)
	}

	// Update Presentation Context length
	pcLength := len(buf) - pcStart - 4
	binary.BigEndian.PutUint16(buf[pcStart+2:pcStart+4], uint16(pcLength))

	// Store presentation context for later use (with first transfer syntax as default)
	a.presentationCtxs[contextID] = &PresentationContext{
		ID:             contextID,
		AbstractSyntax: abstractSyntax,
		TransferSyntax: "",
		Accepted:       false,
	}

	return buf
}

// addUserInformation adds user information to the buffer
func (a *Association) addUserInformation(buf []byte) []byte {
	uiStart := len(buf)

	// User Information Item
	buf = append(buf, 0x50)       // Item type
	buf = append(buf, 0x00)       // Reserved
	buf = append(buf, 0x00, 0x00) // Length placeholder

	// Maximum Length Sub-Item
	buf = append(buf, 0x51)       // Item type
	buf = append(buf, 0x00)       // Reserved
	buf = append(buf, 0x00, 0x04) // Length
	maxLengthBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(maxLengthBytes, a.maxPDULength)
	buf = append(buf, maxLengthBytes...)

	// Implementation Class UID Sub-Item
	implClassUID := "1.2.840.10008.1.2.1"
	buf = append(buf, 0x52)                          // Item type
	buf = append(buf, 0x00)                          // Reserved
	buf = append(buf, 0x00, byte(len(implClassUID))) // Length
	buf = append(buf, []byte(implClassUID)...)

	// Implementation Version Name Sub-Item
	implVersion := "DIMSE-POC-0.1"
	buf = append(buf, 0x55)                         // Item type
	buf = append(buf, 0x00)                         // Reserved
	buf = append(buf, 0x00, byte(len(implVersion))) // Length
	buf = append(buf, []byte(implVersion)...)

	// Update User Information length
	uiLength := len(buf) - uiStart - 4
	binary.BigEndian.PutUint16(buf[uiStart+2:uiStart+4], uint16(uiLength))

	return buf
}

// receiveAssociateAC receives and parses A-ASSOCIATE-AC
func (a *Association) receiveAssociateAC() error {
	// Read PDU header
	header := make([]byte, 6)
	if _, err := io.ReadFull(a.conn, header); err != nil {
		return fmt.Errorf("failed to read PDU header: %w", err)
	}

	pduType := header[0]
	pduLength := binary.BigEndian.Uint32(header[2:6])

	if pduType == pdu.TypeAssociateRJ {
		return fmt.Errorf("association rejected by peer")
	}

	if pduType != pdu.TypeAssociateAC {
		return fmt.Errorf("unexpected PDU type: 0x%02x (expected A-ASSOCIATE-AC)", pduType)
	}

	// Read PDU data
	data := make([]byte, pduLength)
	if _, err := io.ReadFull(a.conn, data); err != nil {
		return fmt.Errorf("failed to read PDU data: %w", err)
	}

	// Parse presentation context results (simplified)
	// In production, you'd want to parse all items properly
	offset := 68 // Skip fixed fields and app context
	for offset+4 <= len(data) {
		itemType := data[offset]
		itemLength := binary.BigEndian.Uint16(data[offset+2 : offset+4])
		itemEnd := offset + 4 + int(itemLength)
		if itemEnd > len(data) {
			break
		}

		if itemType == 0x21 { // Presentation Context Result
			contextID := data[offset+4]
			result := byte(0xff)
			if itemLength >= 4 {
				result = data[offset+7]
			}

			transferSyntax := ""
			subOffset := offset + 8
			for subOffset+4 <= itemEnd {
				subItemType := data[subOffset]
				subItemLength := binary.BigEndian.Uint16(data[subOffset+2 : subOffset+4])
				subItemEnd := subOffset + 4 + int(subItemLength)
				if subItemEnd > itemEnd {
					break
				}

				if subItemType == 0x40 && subItemLength > 0 {
					tsVal := string(data[subOffset+4 : subItemEnd])
					transferSyntax = strings.TrimRight(tsVal, "\x00 ")
				}

				subOffset = subItemEnd
			}

			if pc, ok := a.presentationCtxs[contextID]; ok {
				pc.Accepted = (result == 0)
				if pc.Accepted && transferSyntax != "" {
					pc.TransferSyntax = transferSyntax
				}
				a.logger.Debug("Presentation context negotiation",
					"context_id", contextID,
					"abstract_syntax", pc.AbstractSyntax,
					"result", result,
					"accepted", pc.Accepted,
					"transfer_syntax", pc.TransferSyntax)
			}
		}

		offset = itemEnd
	}

	return nil
}

// sendReleaseRQ sends an A-RELEASE-RQ PDU
func (a *Association) sendReleaseRQ() error {
	pduData := make([]byte, 6)
	pduData[0] = pdu.TypeReleaseRQ
	pduData[1] = 0x00
	binary.BigEndian.PutUint32(pduData[2:6], 4) // Length is always 4
	reserved := make([]byte, 4)

	if _, err := a.conn.Write(pduData); err != nil {
		return err
	}
	if _, err := a.conn.Write(reserved); err != nil {
		return err
	}

	return nil
}

// receiveReleaseRP receives A-RELEASE-RP (or timeout)
func (a *Association) receiveReleaseRP() error {
	header := make([]byte, 6)
	if _, err := io.ReadFull(a.conn, header); err != nil {
		return err // Connection closed or timeout
	}

	pduType := header[0]
	pduLength := binary.BigEndian.Uint32(header[2:6])

	if pduType != pdu.TypeReleaseRP {
		return fmt.Errorf("unexpected PDU type: 0x%02x", pduType)
	}

	// Read and discard PDU data
	data := make([]byte, pduLength)
	io.ReadFull(a.conn, data)

	return nil
}

// GetPresentationContextID finds a presentation context for the given abstract syntax
func (a *Association) GetPresentationContextID(abstractSyntax string) (byte, error) {
	for _, pc := range a.presentationCtxs {
		if pc.AbstractSyntax == abstractSyntax && pc.Accepted {
			return pc.ID, nil
		}
	}
	return 0, fmt.Errorf("no accepted presentation context for abstract syntax: %s", abstractSyntax)
}
