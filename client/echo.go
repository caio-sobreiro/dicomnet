package client

import (
	"fmt"

	"github.com/caio-sobreiro/dicomnet/dimse"
	"github.com/caio-sobreiro/dicomnet/types"
)

const verificationSOPClassUID = "1.2.840.10008.1.1"

// CEchoResponse represents the result of a C-ECHO operation.
type CEchoResponse struct {
	Status    uint16
	MessageID uint16
}

// SendCEcho performs a DICOM C-ECHO (verification) request and returns the response status.
func (a *Association) SendCEcho(messageID uint16) (*CEchoResponse, error) {
	if messageID == 0 {
		messageID = 1
	}

	presContextID, err := a.GetPresentationContextID(verificationSOPClassUID)
	if err != nil {
		return nil, err
	}

	command := &types.Message{
		CommandField:        dimse.CEchoRQ,
		MessageID:           messageID,
		CommandDataSetType:  0x0101, // No dataset present
		Priority:            0x0000, // Medium priority
		AffectedSOPClassUID: verificationSOPClassUID,
	}

	commandData, err := encodeCommand(command)
	if err != nil {
		return nil, fmt.Errorf("failed to encode C-ECHO command: %w", err)
	}

	if err := a.sendDIMSEMessage(presContextID, commandData, nil); err != nil {
		return nil, fmt.Errorf("failed to send C-ECHO request: %w", err)
	}

	msg, _, err := a.receiveDIMSEMessage()
	if err != nil {
		return nil, err
	}

	if msg.CommandField != dimse.CEchoRSP {
		return nil, fmt.Errorf("unexpected command: 0x%04x (expected C-ECHO-RSP)", msg.CommandField)
	}

	return &CEchoResponse{
		Status:    msg.Status,
		MessageID: msg.MessageIDBeingRespondedTo,
	}, nil
}
