package types

// DIMSE Command types
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

// DIMSE Status codes
const (
	StatusSuccess = 0x0000
	StatusPending = 0xFF00
	StatusFailure = 0xC000
)

// Message represents a parsed DIMSE command
type Message struct {
	CommandField              uint16
	MessageID                 uint16
	AffectedSOPClassUID       string
	AffectedSOPInstanceUID    string
	RequestedSOPClassUID      string
	Priority                  uint16
	CommandDataSetType        uint16
	Status                    uint16
	MessageIDBeingRespondedTo uint16
	MoveDestination           string // For C-MOVE-RQ: the AE title of the move destination
	TransferSyntaxUID         string // Negotiated transfer syntax for associated dataset

	// C-MOVE and C-GET response counters
	NumberOfRemainingSuboperations *uint16
	NumberOfCompletedSuboperations *uint16
	NumberOfFailedSuboperations    *uint16
	NumberOfWarningSuboperations   *uint16
}

// ResponseCommandFor maps a DIMSE request command to its corresponding response command.
func ResponseCommandFor(request uint16) uint16 {
	switch request {
	case CStoreRQ:
		return CStoreRSP
	case CGetRQ:
		return CGetRSP
	case CFindRQ:
		return CFindRSP
	case CMoveRQ:
		return CMoveRSP
	case CEchoRQ:
		return CEchoRSP
	default:
		return request | 0x8000
	}
}
