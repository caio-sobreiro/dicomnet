package client

import (
	"fmt"

	"github.com/caio-sobreiro/dicomnet/dimse"
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

	a.logger.Debug("Sending C-STORE-RQ",
		"sop_class", req.SOPClassUID,
		"sop_instance", req.SOPInstanceUID,
		"data_size", len(req.Data))

	// Use shared dimse.SendCStore
	dimseReq := &dimse.CStoreRequest{
		SOPClassUID:    req.SOPClassUID,
		SOPInstanceUID: req.SOPInstanceUID,
		Data:           req.Data,
		MessageID:      req.MessageID,
	}

	dimseResp, err := dimse.SendCStore(a.conn, presContextID, a.maxPDULength, dimseReq)
	if err != nil {
		return nil, err
	}

	return &CStoreResponse{
		Status:         dimseResp.Status,
		MessageID:      dimseResp.MessageID,
		SOPClassUID:    dimseResp.SOPClassUID,
		SOPInstanceUID: dimseResp.SOPInstanceUID,
	}, nil
}
