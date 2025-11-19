package main

import (
	"context"
	"encoding/binary"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/caio-sobreiro/dicomnet/client"
	"github.com/caio-sobreiro/dicomnet/dicom"
	"github.com/caio-sobreiro/dicomnet/interfaces"
	"github.com/caio-sobreiro/dicomnet/server"
	"github.com/caio-sobreiro/dicomnet/types"
)

// DicomInstance represents a stored DICOM instance
type DicomInstance struct {
	SOPClassUID    string
	SOPInstanceUID string
	StudyUID       string
	SeriesUID      string
	TransferSyntax string // Transfer syntax the data is stored in
	Data           []byte
}

type sampleHandler struct {
	instances map[string]*DicomInstance // Key: SOPInstanceUID
	mu        sync.RWMutex
}

func responseTransferSyntax(meta interfaces.MessageContext) string {
	if meta.TransferSyntaxUID != "" {
		return meta.TransferSyntaxUID
	}
	return dicom.TransferSyntaxExplicitVRLittleEndian
}

func (s *sampleHandler) HandleDIMSE(ctx context.Context, msg *types.Message, data []byte, meta interfaces.MessageContext) (*types.Message, *dicom.Dataset, error) {
	slog.InfoContext(ctx, "Received DIMSE command", "command_field", fmt.Sprintf("0x%04X", msg.CommandField), "message_id", msg.MessageID)

	switch msg.CommandField {
	case types.CEchoRQ:
		response := &types.Message{
			CommandField:              types.CEchoRSP,
			MessageIDBeingRespondedTo: msg.MessageID,
			AffectedSOPClassUID:       msg.AffectedSOPClassUID,
			CommandDataSetType:        0x0101,
			Status:                    types.StatusSuccess,
		}
		slog.InfoContext(ctx, "Responding to C-ECHO with success", "message_id", msg.MessageID)
		return response, nil, nil

	case types.CFindRQ:
		// C-FIND uses the streaming handler - this path shouldn't be hit
		// but provide a fallback response
		response := &types.Message{
			CommandField:              types.CFindRSP,
			MessageIDBeingRespondedTo: msg.MessageID,
			AffectedSOPClassUID:       msg.AffectedSOPClassUID,
			CommandDataSetType:        0x0101, // No dataset in final response
			Status:                    types.StatusSuccess,
		}
		slog.InfoContext(ctx, "C-FIND handled via non-streaming path (unexpected)", "message_id", msg.MessageID)
		return response, nil, nil

	case types.CMoveRQ:
		return s.handleCMove(ctx, msg, data, meta)

	default:
		response := &types.Message{
			CommandField:              types.ResponseCommandFor(msg.CommandField),
			MessageIDBeingRespondedTo: msg.MessageID,
			AffectedSOPClassUID:       msg.AffectedSOPClassUID,
			CommandDataSetType:        0x0101,
			Status:                    types.StatusFailure,
		}
		slog.WarnContext(ctx, "Unsupported DIMSE command", "command_field", fmt.Sprintf("0x%04X", msg.CommandField))
		return response, nil, nil
	}
}

func (s *sampleHandler) HandleDIMSEStreaming(ctx context.Context, msg *types.Message, data []byte, meta interfaces.MessageContext, responder interfaces.ResponseSender) error {
	switch msg.CommandField {
	case types.CFindRQ:
		return s.handleCFindStreaming(ctx, msg, data, meta, responder)
	case types.CMoveRQ:
		return s.handleCMoveStreaming(ctx, msg, data, meta, responder)
	case types.CGetRQ:
		return s.handleCGetStreaming(ctx, msg, data, meta, responder)
	default:
		// Fall back to non-streaming handler
		response, dataset, err := s.HandleDIMSE(ctx, msg, data, meta)
		if err != nil {
			return err
		}
		return responder.SendResponse(response, dataset, responseTransferSyntax(meta))
	}
}

func (s *sampleHandler) handleCFindStreaming(ctx context.Context, msg *types.Message, data []byte, meta interfaces.MessageContext, responder interfaces.ResponseSender) error {
	slog.InfoContext(ctx, "Handling C-FIND request", "message_id", msg.MessageID)

	// Create mock study result
	dataset := dicom.NewDataset()
	dataset.AddElement(dicom.Tag{Group: 0x0008, Element: 0x0052}, dicom.VR_CS, "STUDY")
	dataset.AddElement(dicom.Tag{Group: 0x0010, Element: 0x0010}, dicom.VR_PN, "DOE^JOHN")
	dataset.AddElement(dicom.Tag{Group: 0x0010, Element: 0x0020}, dicom.VR_LO, "123456")
	dataset.AddElement(dicom.Tag{Group: 0x0020, Element: 0x000D}, dicom.VR_UI, "1.2.3.4.5.6.7.8.1")
	dataset.AddElement(dicom.Tag{Group: 0x0008, Element: 0x0020}, dicom.VR_DA, "20240101")
	dataset.AddElement(dicom.Tag{Group: 0x0008, Element: 0x0030}, dicom.VR_TM, "120000")
	dataset.AddElement(dicom.Tag{Group: 0x0008, Element: 0x0050}, dicom.VR_SH, "ACC123")
	dataset.AddElement(dicom.Tag{Group: 0x0008, Element: 0x1030}, dicom.VR_LO, "Test Study")

	// Send PENDING response with the match
	pendingResponse := &types.Message{
		CommandField:              types.CFindRSP,
		MessageIDBeingRespondedTo: msg.MessageID,
		AffectedSOPClassUID:       msg.AffectedSOPClassUID,
		CommandDataSetType:        0x0000, // Dataset present
		Status:                    types.StatusPending,
	}
	slog.InfoContext(ctx, "Sending C-FIND pending response with match", "message_id", msg.MessageID)
	if err := responder.SendResponse(pendingResponse, dataset, responseTransferSyntax(meta)); err != nil {
		return err
	}

	// Send final SUCCESS response with no dataset
	finalResponse := &types.Message{
		CommandField:              types.CFindRSP,
		MessageIDBeingRespondedTo: msg.MessageID,
		AffectedSOPClassUID:       msg.AffectedSOPClassUID,
		CommandDataSetType:        0x0101, // No dataset
		Status:                    types.StatusSuccess,
	}
	slog.InfoContext(ctx, "Sending C-FIND final success response", "message_id", msg.MessageID)
	return responder.SendResponse(finalResponse, nil, responseTransferSyntax(meta))
}

func (s *sampleHandler) handleCMoveStreaming(ctx context.Context, msg *types.Message, data []byte, meta interfaces.MessageContext, responder interfaces.ResponseSender) error {
	slog.InfoContext(ctx, "Received C-MOVE request", "move_destination", msg.MoveDestination)

	dataset := meta.Dataset
	if dataset == nil {
		var err error
		dataset, err = dicom.ParseDatasetWithTransferSyntax(data, meta.TransferSyntaxUID)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to parse C-MOVE dataset", "error", err)
			failure := buildMoveResponse(msg, types.StatusFailure, 0, 0, 0, 0)
			return responder.SendResponse(failure, nil, responseTransferSyntax(meta))
		}
	}

	logCMoveRequest(ctx, msg, dataset)

	// Find matching instances
	studyUID := dataset.GetString(dicom.Tag{Group: 0x0020, Element: 0x000D})
	seriesUID := dataset.GetString(dicom.Tag{Group: 0x0020, Element: 0x000E})
	sopUID := dataset.GetString(dicom.Tag{Group: 0x0008, Element: 0x0018})

	matchingInstances := s.findMatchingInstances(studyUID, seriesUID, sopUID)
	totalInstances := len(matchingInstances)

	slog.InfoContext(ctx, "Found matching instances", "count", totalInstances)

	if totalInstances == 0 {
		// No matches - send success with 0 completed
		final := buildMoveResponse(msg, types.StatusSuccess, 0, 0, 0, 0)
		return responder.SendResponse(final, nil, responseTransferSyntax(meta))
	}

	// Perform C-STORE sub-operations
	completed := uint16(0)
	failed := uint16(0)
	warning := uint16(0)

	for i, instance := range matchingInstances {
		remaining := uint16(totalInstances - i)

		// Send pending status before each transfer
		pending := buildMoveResponse(msg, types.StatusPending, remaining, completed, failed, warning)
		if err := responder.SendResponse(pending, nil, responseTransferSyntax(meta)); err != nil {
			return err
		}

		// Perform C-STORE to move destination
		err := s.performCStore(ctx, msg.MoveDestination, instance)
		if err != nil {
			slog.ErrorContext(ctx, "C-STORE sub-operation failed", "error", err, "sop_instance", instance.SOPInstanceUID)
			failed++
		} else {
			slog.InfoContext(ctx, "C-STORE sub-operation successful", "sop_instance", instance.SOPInstanceUID)
			completed++
		}
	}

	// Send final success response
	final := buildMoveResponse(msg, types.StatusSuccess, 0, completed, failed, warning)
	return responder.SendResponse(final, nil, responseTransferSyntax(meta))
}

func (s *sampleHandler) handleCGetStreaming(ctx context.Context, msg *types.Message, data []byte, meta interfaces.MessageContext, responder interfaces.ResponseSender) error {
	slog.InfoContext(ctx, "Received C-GET request")

	dataset := meta.Dataset
	if dataset == nil {
		var err error
		dataset, err = dicom.ParseDatasetWithTransferSyntax(data, meta.TransferSyntaxUID)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to parse C-GET dataset", "error", err)
			failure := buildGetResponse(msg, types.StatusFailure, 0, 0, 0, 0)
			return responder.SendResponse(failure, nil, responseTransferSyntax(meta))
		}
	}

	logCGetRequest(ctx, msg, dataset)

	// Find matching instances
	studyUID := dataset.GetString(dicom.Tag{Group: 0x0020, Element: 0x000D})
	seriesUID := dataset.GetString(dicom.Tag{Group: 0x0020, Element: 0x000E})
	sopUID := dataset.GetString(dicom.Tag{Group: 0x0008, Element: 0x0018})

	matchingInstances := s.findMatchingInstances(studyUID, seriesUID, sopUID)
	totalInstances := len(matchingInstances)

	slog.InfoContext(ctx, "Found matching instances", "count", totalInstances)

	if totalInstances == 0 {
		// No matches - send success with 0 completed
		final := buildGetResponse(msg, types.StatusSuccess, 0, 0, 0, 0)
		return responder.SendResponse(final, nil, responseTransferSyntax(meta))
	}

	// Check if responder supports C-STORE sub-operations
	cgetResponder, ok := responder.(interfaces.CGetResponder)
	if !ok {
		slog.ErrorContext(ctx, "Responder does not support C-GET operations")
		failure := buildGetResponse(msg, types.StatusFailure, 0, 0, 0, 0)
		return responder.SendResponse(failure, nil, responseTransferSyntax(meta))
	}

	// Perform C-STORE sub-operations on the same association
	completed := uint16(0)
	failed := uint16(0)
	warning := uint16(0)

	for i, instance := range matchingInstances {
		remaining := uint16(totalInstances - i)

		// Send pending status before each transfer
		pending := buildGetResponse(msg, types.StatusPending, remaining, completed, failed, warning)
		if err := responder.SendResponse(pending, nil, responseTransferSyntax(meta)); err != nil {
			return err
		}

		// Perform C-STORE on the same association
		err := cgetResponder.SendCStore(instance.SOPClassUID, instance.SOPInstanceUID, instance.Data)
		if err != nil {
			slog.ErrorContext(ctx, "C-STORE sub-operation failed", "error", err, "sop_instance", instance.SOPInstanceUID)
			failed++
		} else {
			slog.InfoContext(ctx, "C-STORE sub-operation successful", "sop_instance", instance.SOPInstanceUID)
			completed++
		}
	}

	// Send final success response
	final := buildGetResponse(msg, types.StatusSuccess, 0, completed, failed, warning)
	return responder.SendResponse(final, nil, responseTransferSyntax(meta))
}

func (s *sampleHandler) findMatchingInstances(studyUID, seriesUID, sopUID string) []*DicomInstance {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var matches []*DicomInstance
	for _, instance := range s.instances {
		// Match based on query level
		if sopUID != "" {
			// Instance level query
			if instance.SOPInstanceUID == sopUID {
				matches = append(matches, instance)
			}
		} else if seriesUID != "" {
			// Series level query
			if instance.SeriesUID == seriesUID {
				matches = append(matches, instance)
			}
		} else if studyUID != "" {
			// Study level query
			if instance.StudyUID == studyUID {
				matches = append(matches, instance)
			}
		}
	}
	return matches
}

func (s *sampleHandler) performCStore(ctx context.Context, destination string, instance *DicomInstance) error {
	// Create client connection to move destination
	// Propose transfer syntaxes with the instance's native transfer syntax first
	config := client.Config{
		CallingAETitle:            "SAMPLE_SCP",
		CalledAETitle:             destination,
		MaxPDULength:              16384,
		PreferredTransferSyntaxes: s.buildTransferSyntaxList(instance.TransferSyntax),
	}

	assoc, err := client.Connect("orthanc:4242", config)
	if err != nil {
		return fmt.Errorf("failed to connect to destination: %w", err)
	}
	defer assoc.Close()

	// Send C-STORE
	storeReq := &client.CStoreRequest{
		SOPClassUID:    instance.SOPClassUID,
		SOPInstanceUID: instance.SOPInstanceUID,
		Data:           instance.Data,
		MessageID:      1,
	}

	resp, err := assoc.SendCStore(storeReq)
	if err != nil {
		return fmt.Errorf("C-STORE failed: %w", err)
	}

	if resp.Status != 0x0000 {
		return fmt.Errorf("C-STORE returned error status: 0x%04X", resp.Status)
	}

	return nil
}

// buildTransferSyntaxList creates a prioritized list of transfer syntaxes
// with the instance's native transfer syntax first, followed by common ones
func (s *sampleHandler) buildTransferSyntaxList(nativeTS string) []string {
	// Start with the native transfer syntax
	syntaxes := []string{nativeTS}

	// Add common transfer syntaxes as fallbacks (only if different from native)
	common := []string{
		types.ExplicitVRLittleEndian, // Explicit VR Little Endian
		types.ImplicitVRLittleEndian, // Implicit VR Little Endian
		types.JPEG2000Lossless,       // JPEG 2000 Lossless Only
		types.JPEG2000,               // JPEG 2000
	}

	for _, ts := range common {
		if ts != nativeTS {
			syntaxes = append(syntaxes, ts)
		}
	}

	return syntaxes
}

func (s *sampleHandler) handleCMove(ctx context.Context, msg *types.Message, data []byte, meta interfaces.MessageContext) (*types.Message, *dicom.Dataset, error) {
	dataset := meta.Dataset
	if dataset == nil {
		var err error
		dataset, err = dicom.ParseDatasetWithTransferSyntax(data, meta.TransferSyntaxUID)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to parse C-MOVE dataset", "error", err)
			failure := buildMoveResponse(msg, types.StatusFailure, 0, 0, 0, 0)
			return failure, nil, nil
		}
	}

	logCMoveRequest(ctx, msg, dataset)

	response := buildMoveResponse(msg, types.StatusSuccess, 0, 0, 0, 0)
	return response, nil, nil
}

func (s *sampleHandler) loadDicomFile(filepath string) error {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("failed to read DICOM file: %w", err)
	}

	// Parse DICOM file to extract metadata
	// Skip the 128-byte preamble and "DICM" prefix
	if len(data) < 132 {
		return fmt.Errorf("file too small to be valid DICOM")
	}
	if string(data[128:132]) != "DICM" {
		return fmt.Errorf("missing DICM prefix")
	}

	// Parse the dataset starting after preamble
	dataset, err := dicom.ParseDataset(data[132:])
	if err != nil {
		return fmt.Errorf("failed to parse DICOM dataset: %w", err)
	}

	// Extract Transfer Syntax UID (0002,0010) from the file meta information
	// This is in the first part of the file (before the dataset)
	transferSyntax := types.ExplicitVRLittleEndian // Default to Explicit VR Little Endian
	if len(data) > 132 {
		// Try to find transfer syntax in meta information (group 0x0002)
		// For now, we'll use a simple approach - in production you'd parse the meta info properly
		tsTag := []byte{0x02, 0x00, 0x10, 0x00} // (0002,0010) Transfer Syntax UID
		for i := 132; i < len(data)-20 && i < 300; i++ {
			if data[i] == tsTag[0] && data[i+1] == tsTag[1] &&
				data[i+2] == tsTag[2] && data[i+3] == tsTag[3] {
				// Found transfer syntax tag, read the value
				vr := string(data[i+4 : i+6])
				if vr == "UI" {
					length := binary.LittleEndian.Uint16(data[i+6 : i+8])
					if i+8+int(length) <= len(data) {
						transferSyntax = strings.TrimRight(string(data[i+8:i+8+int(length)]), "\x00 ")
						break
					}
				}
			}
		}
	}

	instance := &DicomInstance{
		SOPClassUID:    dataset.GetString(dicom.Tag{Group: 0x0008, Element: 0x0016}),
		SOPInstanceUID: dataset.GetString(dicom.Tag{Group: 0x0008, Element: 0x0018}),
		StudyUID:       dataset.GetString(dicom.Tag{Group: 0x0020, Element: 0x000D}),
		SeriesUID:      dataset.GetString(dicom.Tag{Group: 0x0020, Element: 0x000E}),
		TransferSyntax: transferSyntax,
		Data:           data[132:], // Store only the dataset, not the preamble
	}

	s.mu.Lock()
	s.instances[instance.SOPInstanceUID] = instance
	s.mu.Unlock()

	slog.Info("Loaded DICOM instance",
		"sop_class", instance.SOPClassUID,
		"sop_instance", instance.SOPInstanceUID,
		"study_uid", instance.StudyUID,
		"series_uid", instance.SeriesUID,
		"transfer_syntax", instance.TransferSyntax,
		"size_bytes", len(data))

	return nil
}

// generateSyntheticInstance creates a synthetic DICOM instance in memory
func (s *sampleHandler) generateSyntheticInstance(sopInstanceUID, studyUID, seriesUID string) error {
	// Build a minimal DICOM dataset with required tags
	// Using Implicit VR Little Endian
	buf := make([]byte, 0, 512)

	// Helper to append elements in Implicit VR format
	appendElement := func(group, element uint16, vr string, value []byte) {
		// Tag
		buf = append(buf, byte(group), byte(group>>8), byte(element), byte(element>>8))
		// Length (4 bytes in Implicit VR)
		length := uint32(len(value))
		buf = append(buf, byte(length), byte(length>>8), byte(length>>16), byte(length>>24))
		// Value
		buf = append(buf, value...)
	}

	// SOP Class UID (0008,0016) - CT Image Storage
	sopClassUID := types.CTImageStorage
	appendElement(0x0008, 0x0016, "UI", []byte(sopClassUID))

	// SOP Instance UID (0008,0018)
	appendElement(0x0008, 0x0018, "UI", []byte(sopInstanceUID))

	// Study Date (0008,0020)
	appendElement(0x0008, 0x0020, "DA", []byte("20250109"))

	// Study Time (0008,0030)
	appendElement(0x0008, 0x0030, "TM", []byte("120000"))

	// Modality (0008,0060)
	appendElement(0x0008, 0x0060, "CS", []byte("CT"))

	// Patient Name (0010,0010)
	appendElement(0x0010, 0x0010, "PN", []byte("TEST^PATIENT"))

	// Patient ID (0010,0020)
	appendElement(0x0010, 0x0020, "LO", []byte("12345"))

	// Study Instance UID (0020,000D)
	appendElement(0x0020, 0x000D, "UI", []byte(studyUID))

	// Series Instance UID (0020,000E)
	appendElement(0x0020, 0x000E, "UI", []byte(seriesUID))

	// Instance Number (0020,0013)
	appendElement(0x0020, 0x0013, "IS", []byte("1"))

	// Rows (0028,0010) - minimal image dimensions
	rows := make([]byte, 2)
	binary.LittleEndian.PutUint16(rows, 512)
	appendElement(0x0028, 0x0010, "US", rows)

	// Columns (0028,0011)
	cols := make([]byte, 2)
	binary.LittleEndian.PutUint16(cols, 512)
	appendElement(0x0028, 0x0011, "US", cols)

	// Bits Allocated (0028,0100)
	bits := make([]byte, 2)
	binary.LittleEndian.PutUint16(bits, 16)
	appendElement(0x0028, 0x0100, "US", bits)

	// Pixel Data (7FE0,0010) - empty for now
	appendElement(0x7FE0, 0x0010, "OW", []byte{})

	instance := &DicomInstance{
		SOPInstanceUID: sopInstanceUID,
		StudyUID:       studyUID,
		SeriesUID:      seriesUID,
		TransferSyntax: types.ImplicitVRLittleEndian, // Implicit VR Little Endian
		Data:           buf,
	}

	s.mu.Lock()
	s.instances[instance.SOPInstanceUID] = instance
	s.mu.Unlock()

	slog.Info("Generated synthetic DICOM instance",
		"sop_class", instance.SOPClassUID,
		"sop_instance", instance.SOPInstanceUID,
		"study_uid", instance.StudyUID,
		"series_uid", instance.SeriesUID,
		"transfer_syntax", instance.TransferSyntax,
		"size_bytes", len(buf))

	return nil
}

func main() {
	port := flag.Int("port", 4242, "TCP port to listen on")
	aeTitle := flag.String("ae", "SAMPLE_SCP", "Server AE Title")
	dicomFile := flag.String("dicom", "sample.dcm", "Path to sample DICOM file (optional)")
	generateSynthetic := flag.Bool("synthetic", false, "Generate synthetic DICOM instances instead of loading from file")
	flag.Parse()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	handler := &sampleHandler{
		instances: make(map[string]*DicomInstance),
	}

	// Load or generate DICOM instances
	if *generateSynthetic {
		// Generate synthetic instances
		studyUID := "1.2.840.999.999.1.1.1.1"
		seriesUID := "1.2.840.999.999.1.1.1.1.1"

		// Generate 3 instances in the same series
		for i := 1; i <= 3; i++ {
			sopInstanceUID := fmt.Sprintf("1.2.840.999.999.1.1.1.1.1.%d", i)
			if err := handler.generateSyntheticInstance(sopInstanceUID, studyUID, seriesUID); err != nil {
				logger.Error("Failed to generate synthetic instance", "error", err, "instance", i)
				os.Exit(1)
			}
		}
	} else if *dicomFile != "" {
		// Load from file
		if err := handler.loadDicomFile(*dicomFile); err != nil {
			logger.Error("Failed to load DICOM file", "error", err, "file", *dicomFile)
			os.Exit(1)
		}
	} else {
		logger.Error("Must specify either --dicom <file> or --synthetic")
		os.Exit(1)
	}

	address := fmt.Sprintf(":%d", *port)

	err := server.ListenAndServe(ctx, address, *aeTitle, handler, server.WithLogger(logger))
	switch {
	case err == nil:
		logger.Info("Sample server shutdown complete")
	case errors.Is(err, context.Canceled):
		logger.Info("Sample server stopped", "reason", err.Error())
	default:
		logger.Error("Sample server terminated unexpectedly", "error", err)
		os.Exit(1)
	}
}

func buildMoveResponse(req *types.Message, status uint16, remaining, completed, failed, warning uint16) *types.Message {
	// Helper to create uint16 pointers
	uint16Ptr := func(v uint16) *uint16 { return &v }

	resp := &types.Message{
		CommandField:                   types.CMoveRSP,
		MessageIDBeingRespondedTo:      req.MessageID,
		AffectedSOPClassUID:            req.AffectedSOPClassUID,
		CommandDataSetType:             0x0101,
		Status:                         status,
		NumberOfRemainingSuboperations: uint16Ptr(remaining),
		NumberOfCompletedSuboperations: uint16Ptr(completed),
		NumberOfFailedSuboperations:    uint16Ptr(failed),
		NumberOfWarningSuboperations:   uint16Ptr(warning),
	}

	return resp
}

func buildGetResponse(req *types.Message, status uint16, remaining, completed, failed, warning uint16) *types.Message {
	// Helper to create uint16 pointers
	uint16Ptr := func(v uint16) *uint16 { return &v }

	resp := &types.Message{
		CommandField:                   types.CGetRSP,
		MessageIDBeingRespondedTo:      req.MessageID,
		AffectedSOPClassUID:            req.AffectedSOPClassUID,
		CommandDataSetType:             0x0101,
		Status:                         status,
		NumberOfRemainingSuboperations: uint16Ptr(remaining),
		NumberOfCompletedSuboperations: uint16Ptr(completed),
		NumberOfFailedSuboperations:    uint16Ptr(failed),
		NumberOfWarningSuboperations:   uint16Ptr(warning),
	}

	return resp
}

func logCMoveRequest(ctx context.Context, msg *types.Message, dataset *dicom.Dataset) {
	if dataset == nil {
		slog.InfoContext(ctx, "Handling C-MOVE request",
			"move_destination", msg.MoveDestination,
			"note", "no dataset provided")
		return
	}

	slog.InfoContext(ctx, "Handling C-MOVE request",
		"move_destination", msg.MoveDestination,
		"study_uid", dataset.GetString(dicom.Tag{Group: 0x0020, Element: 0x000D}),
		"series_uid", dataset.GetString(dicom.Tag{Group: 0x0020, Element: 0x000E}),
		"sop_uid", dataset.GetString(dicom.Tag{Group: 0x0008, Element: 0x0018}))
}

func logCGetRequest(ctx context.Context, msg *types.Message, dataset *dicom.Dataset) {
	if dataset == nil {
		slog.InfoContext(ctx, "Handling C-GET request", "note", "no dataset provided")
		return
	}

	slog.InfoContext(ctx, "Handling C-GET request",
		"study_uid", dataset.GetString(dicom.Tag{Group: 0x0020, Element: 0x000D}),
		"series_uid", dataset.GetString(dicom.Tag{Group: 0x0020, Element: 0x000E}),
		"sop_uid", dataset.GetString(dicom.Tag{Group: 0x0008, Element: 0x0018}))
}
