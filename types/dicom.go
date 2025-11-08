// Package types contains all DICOM-related type definitions
package types

import "fmt"

// VR (Value Representation) constants for DICOM data elements
const (
	VR_AE = "AE" // Application Entity
	VR_AS = "AS" // Age String
	VR_AT = "AT" // Attribute Tag
	VR_CS = "CS" // Code String
	VR_DA = "DA" // Date
	VR_DS = "DS" // Decimal String
	VR_DT = "DT" // Date Time
	VR_FL = "FL" // Floating Point Single
	VR_FD = "FD" // Floating Point Double
	VR_IS = "IS" // Integer String
	VR_LO = "LO" // Long String
	VR_LT = "LT" // Long Text
	VR_OB = "OB" // Other Byte
	VR_OD = "OD" // Other Double
	VR_OF = "OF" // Other Float
	VR_OL = "OL" // Other Long
	VR_OV = "OV" // Other Very Long
	VR_OW = "OW" // Other Word
	VR_PN = "PN" // Person Name
	VR_SH = "SH" // Short String
	VR_SL = "SL" // Signed Long
	VR_SQ = "SQ" // Sequence of Items
	VR_SS = "SS" // Signed Short
	VR_ST = "ST" // Short Text
	VR_SV = "SV" // Signed Very Long
	VR_TM = "TM" // Time
	VR_UC = "UC" // Unlimited Characters
	VR_UI = "UI" // Unique Identifier
	VR_UL = "UL" // Unsigned Long
	VR_UN = "UN" // Unknown
	VR_UR = "UR" // Universal Resource
	VR_US = "US" // Unsigned Short
	VR_UT = "UT" // Unlimited Text
	VR_UV = "UV" // Unsigned Very Long
)

// Tag represents a DICOM tag (group, element)
type Tag struct {
	Group   uint16
	Element uint16
}

// String returns the tag as a string in (GGGG,EEEE) format
func (t Tag) String() string {
	return fmt.Sprintf("(%04x,%04x)", t.Group, t.Element)
}

// Element represents a DICOM data element
type Element struct {
	Tag   Tag
	VR    string
	Value interface{}
}

// Dataset represents a collection of DICOM elements
type Dataset struct {
	Elements map[Tag]*Element
}
