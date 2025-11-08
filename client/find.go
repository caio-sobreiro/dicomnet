package client

import (
	"fmt"
	"log/slog"

	"github.com/caio-sobreiro/dicomnet/dicom"
	"github.com/caio-sobreiro/dicomnet/dimse"
	"github.com/caio-sobreiro/dicomnet/types"
)

const studyRootFindSOPClassUID = "1.2.840.10008.5.1.4.1.2.2.1"

// CFindRequest encapsulates the information required to perform a C-FIND query.
type CFindRequest struct {
	SOPClassUID string
	MessageID   uint16
	Priority    uint16
	Dataset     *dicom.Dataset
}

// CFindResponse represents a single C-FIND response from the SCP.
type CFindResponse struct {
	Status    uint16
	MessageID uint16
	Dataset   *dicom.Dataset
}

// SendCFind performs a DICOM C-FIND query and returns all responses in order.
func (a *Association) SendCFind(req *CFindRequest) ([]*CFindResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("c-find request cannot be nil")
	}

	if req.Dataset == nil {
		return nil, fmt.Errorf("c-find request requires a dataset")
	}

	sopClass := req.SOPClassUID
	if sopClass == "" {
		sopClass = studyRootFindSOPClassUID
	}

	messageID := req.MessageID
	if messageID == 0 {
		messageID = 1
	}

	priority := req.Priority
	if priority == 0 {
		priority = 0x0000 // Medium priority per DICOM PS3.7
	}

	presContextID, err := a.GetPresentationContextID(sopClass)
	if err != nil {
		return nil, err
	}

	command := &types.Message{
		CommandField:        dimse.CFindRQ,
		MessageID:           messageID,
		CommandDataSetType:  0x0000, // Dataset present
		Priority:            priority,
		AffectedSOPClassUID: sopClass,
	}

	commandData, err := encodeCommand(command)
	if err != nil {
		return nil, fmt.Errorf("failed to encode C-FIND command: %w", err)
	}

	datasetData := req.Dataset.EncodeDataset()

	if err := a.sendDIMSEMessage(presContextID, commandData, datasetData); err != nil {
		return nil, fmt.Errorf("failed to send C-FIND request: %w", err)
	}

	var responses []*CFindResponse

	for {
		msg, data, err := a.receiveDIMSEMessage()
		if err != nil {
			return nil, err
		}

		if msg.CommandField != dimse.CFindRSP {
			return nil, fmt.Errorf("unexpected command: 0x%04x (expected C-FIND-RSP)", msg.CommandField)
		}

		var dataset *dicom.Dataset
		if len(data) > 0 {
			dataset, err = dicom.ParseDataset(data)
			if err != nil {
				slog.Warn("Failed to parse C-FIND response dataset",
					"error", err,
					"message_id", msg.MessageIDBeingRespondedTo,
					"status", fmt.Sprintf("0x%04X", msg.Status))
			}
		}

		responses = append(responses, &CFindResponse{
			Status:    msg.Status,
			MessageID: msg.MessageIDBeingRespondedTo,
			Dataset:   dataset,
		})

		if msg.Status != dimse.StatusPending {
			break
		}
	}

	return responses, nil
}
