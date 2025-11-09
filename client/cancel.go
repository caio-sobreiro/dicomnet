package client

import (
	"fmt"

	"github.com/caio-sobreiro/dicomnet/dimse"
	"github.com/caio-sobreiro/dicomnet/types"
)

// SendCCancel sends a C-CANCEL-RQ to cancel a pending C-FIND or C-MOVE operation.
// The messageID parameter must match the MessageID of the operation being canceled.
// C-CANCEL does not have a response - it's a notification to the SCP to stop sending
// pending responses for the specified operation.
func (a *Association) SendCCancel(messageID uint16, sopClassUID string) error {
	if messageID == 0 {
		return fmt.Errorf("messageID must be non-zero for C-CANCEL")
	}

	if sopClassUID == "" {
		return fmt.Errorf("sopClassUID must be provided for C-CANCEL")
	}

	presContextID, err := a.GetPresentationContextID(sopClassUID)
	if err != nil {
		return err
	}

	command := &types.Message{
		CommandField:              dimse.CCancelRQ,
		MessageIDBeingRespondedTo: messageID,
		CommandDataSetType:        0x0101, // No dataset present
	}

	commandData, err := dimse.EncodeCommand(command)
	if err != nil {
		return fmt.Errorf("failed to encode C-CANCEL command: %w", err)
	}

	if err := dimse.SendDIMSEMessage(a.conn, presContextID, a.maxPDULength, commandData, nil); err != nil {
		return fmt.Errorf("failed to send C-CANCEL request: %w", err)
	}

	a.logger.Debug("C-CANCEL sent", "messageID", messageID, "sopClassUID", sopClassUID)

	return nil
}
