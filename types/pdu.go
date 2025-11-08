package types

// PDU type constants
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
