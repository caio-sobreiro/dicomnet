package client

import (
	"fmt"

	"github.com/caio-sobreiro/dicomnet/dicom"
	"github.com/caio-sobreiro/dicomnet/dimse"
	"github.com/caio-sobreiro/dicomnet/types"
)

// CGetRequest encapsulates the information required to perform a C-GET operation.
type CGetRequest struct {
	SOPClassUID string
	MessageID   uint16
	Priority    uint16
	Dataset     *dicom.Dataset // Query identifying which instances to retrieve
}

// CGetResponse represents a single C-GET response from the SCP.
type CGetResponse struct {
	Status                         uint16
	MessageID                      uint16
	NumberOfRemainingSuboperations *uint16
	NumberOfCompletedSuboperations *uint16
	NumberOfFailedSuboperations    *uint16
	NumberOfWarningSuboperations   *uint16
}

// SendCGet performs a DICOM C-GET operation to retrieve instances.
// The SCP will send C-STORE operations on the same association for each matching instance.
//
// Returns responses indicating the progress and final status of the retrieval.
// Note: The caller must handle incoming C-STORE requests on this association.
func (a *Association) SendCGet(req *CGetRequest) ([]*CGetResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("c-get request cannot be nil")
	}

	if req.Dataset == nil {
		return nil, fmt.Errorf("c-get request requires a dataset")
	}

	sopClass := req.SOPClassUID
	if sopClass == "" {
		sopClass = types.StudyRootQueryRetrieveInformationModelGet
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

	// Encode the query dataset
	datasetBytes := req.Dataset.EncodeDataset()

	// Build C-GET-RQ command
	command := &types.Message{
		CommandField:        dimse.CGetRQ,
		MessageID:           messageID,
		Priority:            priority,
		AffectedSOPClassUID: sopClass,
		CommandDataSetType:  0x0000, // Dataset present
	}

	commandData, err := dimse.EncodeCommand(command)
	if err != nil {
		return nil, fmt.Errorf("failed to encode C-GET command: %w", err)
	}

	// Send C-GET-RQ with dataset
	if err := dimse.SendDIMSEMessage(a.conn, presContextID, a.maxPDULength, commandData, datasetBytes); err != nil {
		return nil, fmt.Errorf("failed to send C-GET request: %w", err)
	}

	// Collect responses
	var responses []*CGetResponse

	for {
		responseCmd, _, err := dimse.ReceiveDIMSEMessage(a.conn)
		if err != nil {
			return responses, fmt.Errorf("failed to receive C-GET response: %w", err)
		}

		if responseCmd.CommandField != dimse.CGetRSP {
			return responses, fmt.Errorf("unexpected response command: 0x%04X (expected C-GET-RSP)", responseCmd.CommandField)
		}

		response := &CGetResponse{
			Status:                         responseCmd.Status,
			MessageID:                      responseCmd.MessageIDBeingRespondedTo,
			NumberOfRemainingSuboperations: responseCmd.NumberOfRemainingSuboperations,
			NumberOfCompletedSuboperations: responseCmd.NumberOfCompletedSuboperations,
			NumberOfFailedSuboperations:    responseCmd.NumberOfFailedSuboperations,
			NumberOfWarningSuboperations:   responseCmd.NumberOfWarningSuboperations,
		}

		responses = append(responses, response)

		// Check if this is the final response
		if responseCmd.Status != dimse.StatusPending {
			break
		}
	}

	return responses, nil
}
