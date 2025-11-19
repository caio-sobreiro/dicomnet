package dicom

import (
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/caio-sobreiro/dicomnet/types"
)

// VR (Value Representation) constants
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

// Common transfer syntax UIDs
const (
	TransferSyntaxImplicitVRLittleEndian = types.ImplicitVRLittleEndian
	TransferSyntaxExplicitVRLittleEndian = types.ExplicitVRLittleEndian
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
	Tag    Tag
	VR     string
	Length uint32
	Value  interface{}
}

// Dataset represents a collection of DICOM elements
type Dataset struct {
	Elements map[Tag]*Element
}

// NewDataset creates a new empty dataset
func NewDataset() *Dataset {
	return &Dataset{
		Elements: make(map[Tag]*Element),
	}
}

// AddElement adds an element to the dataset
func (d *Dataset) AddElement(tag Tag, vr string, value interface{}) {
	element := &Element{
		Tag:   tag,
		VR:    vr,
		Value: value,
	}
	d.Elements[tag] = element
}

// GetElement returns an element by tag
func (d *Dataset) GetElement(tag Tag) (*Element, bool) {
	element, exists := d.Elements[tag]
	return element, exists
}

// GetString returns a string value for a tag
func (d *Dataset) GetString(tag Tag) string {
	if element, exists := d.Elements[tag]; exists {
		if str, ok := element.Value.(string); ok {
			return strings.TrimSpace(str)
		}
	}
	return ""
}

// GetStrings returns a slice of string values for a tag
func (d *Dataset) GetStrings(tag Tag) []string {
	if element, exists := d.Elements[tag]; exists {
		switch v := element.Value.(type) {
		case string:
			// Split by backslash for multiple values
			parts := strings.Split(v, "\\")
			result := make([]string, len(parts))
			for i, part := range parts {
				result[i] = strings.TrimSpace(part)
			}
			return result
		case []string:
			return v
		}
	}
	return nil
}

// ParseDataset parses a DICOM dataset from raw bytes (Explicit VR Little Endian)
func ParseDataset(data []byte) (*Dataset, error) {
	dataset := NewDataset()

	if len(data) == 0 {
		return dataset, nil
	}

	offset := 0
	for offset < len(data) {
		// Need at least 8 bytes for tag + VR + length
		if offset+8 > len(data) {
			break
		}

		// Read tag (4 bytes)
		group := binary.LittleEndian.Uint16(data[offset : offset+2])
		element := binary.LittleEndian.Uint16(data[offset+2 : offset+4])
		tag := Tag{Group: group, Element: element}

		// Read VR (2 bytes)
		vr := string(data[offset+4 : offset+6])

		var length uint32
		var valueOffset int

		// Determine if this is a short or long VR
		// Short VRs: AE, AS, AT, CS, DA, DS, DT, FL, FD, IS, LO, LT, PN, SH, SL, SS, ST, TM, UI, UL, US
		// Long VRs: OB, OD, OF, OL, OW, SQ, UC, UR, UT, UN, OV, SV, UV
		isLongVR := vr == "OB" || vr == "OD" || vr == "OF" || vr == "OL" || vr == "OW" ||
			vr == "SQ" || vr == "UC" || vr == "UR" || vr == "UT" || vr == "UN" ||
			vr == "OV" || vr == "SV" || vr == "UV"

		if isLongVR {
			// Long VR: Tag (4) + VR (2) + Reserved (2) + Length (4) = 12 bytes header
			if offset+12 > len(data) {
				break
			}
			// Skip 2 reserved bytes
			length = binary.LittleEndian.Uint32(data[offset+8 : offset+12])
			valueOffset = offset + 12
		} else {
			// Short VR: Tag (4) + VR (2) + Length (2) = 8 bytes header
			length = uint32(binary.LittleEndian.Uint16(data[offset+6 : offset+8]))
			valueOffset = offset + 8
		}

		// Ensure we have enough data for the value
		if valueOffset+int(length) > len(data) {
			break
		}

		// Extract value
		valueData := data[valueOffset : valueOffset+int(length)]
		value := parseElementValue(tag, valueData)

		dataset.AddElement(tag, vr, value)

		// Move to next element (including padding if odd length)
		nextOffset := valueOffset + int(length)
		if length%2 == 1 {
			nextOffset++
		}
		offset = nextOffset
	}

	return dataset, nil
}

// ParseDatasetWithTransferSyntax parses a dataset using the provided transfer syntax.
func ParseDatasetWithTransferSyntax(data []byte, transferSyntaxUID string) (*Dataset, error) {
	switch transferSyntaxUID {
	case "", TransferSyntaxExplicitVRLittleEndian:
		return ParseDataset(data)
	case TransferSyntaxImplicitVRLittleEndian:
		return parseImplicitVRDataset(data)
	default:
		return ParseDataset(data)
	}
}

func parseImplicitVRDataset(data []byte) (*Dataset, error) {
	dataset := NewDataset()

	if len(data) == 0 {
		return dataset, nil
	}

	offset := 0
	for offset+8 <= len(data) {
		group := binary.LittleEndian.Uint16(data[offset : offset+2])
		element := binary.LittleEndian.Uint16(data[offset+2 : offset+4])
		tag := Tag{Group: group, Element: element}

		length := binary.LittleEndian.Uint32(data[offset+4 : offset+8])
		valueOffset := offset + 8

		if valueOffset+int(length) > len(data) {
			break
		}

		valueData := data[valueOffset : valueOffset+int(length)]
		vr := determineVR(tag)
		value := parseElementValue(tag, valueData)

		dataset.AddElement(tag, vr, value)

		nextOffset := valueOffset + int(length)
		if length%2 == 1 {
			nextOffset++
		}
		offset = nextOffset
	}

	return dataset, nil
}

// parseElementValue parses the value based on the tag and raw data
func parseElementValue(tag Tag, data []byte) interface{} {
	if len(data) == 0 {
		return ""
	}

	// For most query elements, we treat them as strings
	// Remove null padding
	value := string(data)
	if idx := strings.IndexByte(value, 0); idx != -1 {
		value = value[:idx]
	}

	return strings.TrimSpace(value)
}

// determineVR determines the VR based on the tag (simplified mapping)
func determineVR(tag Tag) string {
	// This is a simplified mapping - in practice you'd use a DICOM dictionary
	switch tag {
	case Tag{0x0008, 0x0005}: // Specific Character Set
		return VR_CS
	case Tag{0x0008, 0x0016}: // SOP Class UID
		return VR_UI
	case Tag{0x0008, 0x0018}: // SOP Instance UID
		return VR_UI
	case Tag{0x0008, 0x0020}: // Study Date
		return VR_DA
	case Tag{0x0008, 0x0030}: // Study Time
		return VR_TM
	case Tag{0x0008, 0x0050}: // Accession Number
		return VR_SH
	case Tag{0x0008, 0x0052}: // Query/Retrieve Level
		return VR_CS
	case Tag{0x0008, 0x0054}: // Retrieve AE Title
		return VR_AE
	case Tag{0x0008, 0x0060}: // Modality
		return VR_CS
	case Tag{0x0008, 0x0080}: // Institution Name
		return VR_LO
	case Tag{0x0008, 0x0090}: // Referring Physician's Name
		return VR_PN
	case Tag{0x0008, 0x1030}: // Study Description
		return VR_LO
	case Tag{0x0008, 0x103E}: // Series Description
		return VR_LO
	case Tag{0x0008, 0x1040}: // Institutional Department Name
		return VR_LO
	case Tag{0x0008, 0x1050}: // Performing Physician's Name
		return VR_PN
	case Tag{0x0008, 0x1060}: // Name of Physician(s) Reading Study
		return VR_PN
	case Tag{0x0008, 0x1070}: // Operators' Name
		return VR_PN
	case Tag{0x0010, 0x0010}: // Patient's Name
		return VR_PN
	case Tag{0x0010, 0x0020}: // Patient ID
		return VR_LO
	case Tag{0x0010, 0x0030}: // Patient's Birth Date
		return VR_DA
	case Tag{0x0010, 0x0040}: // Patient's Sex
		return VR_CS
	case Tag{0x0010, 0x1010}: // Patient's Age
		return VR_AS
	case Tag{0x0018, 0x0015}: // Body Part Examined
		return VR_CS
	case Tag{0x0020, 0x000D}: // Study Instance UID
		return VR_UI
	case Tag{0x0020, 0x000E}: // Series Instance UID
		return VR_UI
	case Tag{0x0020, 0x0010}: // Study ID
		return VR_SH
	case Tag{0x0020, 0x0011}: // Series Number
		return VR_IS
	case Tag{0x0020, 0x0013}: // Instance Number
		return VR_IS
	case Tag{0x0020, 0x0020}: // Patient Orientation
		return VR_CS
	default:
		return VR_UN // Unknown
	}
}

// EncodeDataset encodes a dataset to bytes (Explicit VR Little Endian)
func (d *Dataset) EncodeDataset() []byte {
	var result []byte

	// Collect tags and sort them (DICOM requires tag ordering)
	var tags []Tag
	for tag := range d.Elements {
		tags = append(tags, tag)
	}

	// Sort tags by group, then by element
	for i := 0; i < len(tags)-1; i++ {
		for j := i + 1; j < len(tags); j++ {
			if tags[i].Group > tags[j].Group ||
				(tags[i].Group == tags[j].Group && tags[i].Element > tags[j].Element) {
				tags[i], tags[j] = tags[j], tags[i]
			}
		}
	}

	// Add elements in sorted tag order (using Explicit VR Little Endian)
	for _, tag := range tags {
		element := d.Elements[tag]

		// Tag (4 bytes - Little Endian)
		tagBytes := make([]byte, 4)
		binary.LittleEndian.PutUint16(tagBytes[0:2], tag.Group)
		binary.LittleEndian.PutUint16(tagBytes[2:4], tag.Element)
		result = append(result, tagBytes...)

		// VR (2 bytes - ASCII)
		result = append(result, []byte(element.VR)...)

		// Encode value
		valueBytes := encodeElementValue(element)

		// Add padding if odd length (DICOM requires even lengths)
		if len(valueBytes)%2 == 1 {
			valueBytes = append(valueBytes, 0x20) // Use space padding for text elements
		}

		// For Explicit VR, length encoding depends on VR type
		// Short VRs (most string types): 2-byte length
		// Long VRs (OB, OW, SQ, UN, UT): 4-byte length with 2 reserved bytes
		isLongVR := element.VR == VR_OB || element.VR == VR_OW || element.VR == VR_SQ ||
			element.VR == VR_UN || element.VR == VR_UT || element.VR == VR_OD ||
			element.VR == VR_OF || element.VR == VR_OL || element.VR == VR_OV ||
			element.VR == VR_UC || element.VR == VR_UR

		if isLongVR {
			// Long VR format: VR (2 bytes) + Reserved (2 bytes) + Length (4 bytes)
			result = append(result, 0x00, 0x00) // Reserved bytes
			lengthBytes := make([]byte, 4)
			binary.LittleEndian.PutUint32(lengthBytes, uint32(len(valueBytes)))
			result = append(result, lengthBytes...)
		} else {
			// Short VR format: VR (2 bytes) + Length (2 bytes)
			if len(valueBytes) > 65535 {
				// Value too long for short VR format - truncate or error
				valueBytes = valueBytes[:65535]
			}
			lengthBytes := make([]byte, 2)
			binary.LittleEndian.PutUint16(lengthBytes, uint16(len(valueBytes)))
			result = append(result, lengthBytes...)
		}

		// Value (already padded)
		result = append(result, valueBytes...)
	}

	return result
}

// EncodeDatasetWithTransferSyntax encodes a dataset using the provided transfer syntax.
func EncodeDatasetWithTransferSyntax(dataset *Dataset, transferSyntaxUID string) ([]byte, error) {
	if dataset == nil {
		return nil, nil
	}

	switch transferSyntaxUID {
	case "", TransferSyntaxExplicitVRLittleEndian:
		return dataset.EncodeDataset(), nil
	case TransferSyntaxImplicitVRLittleEndian:
		return encodeImplicitVRDataset(dataset), nil
	default:
		return dataset.EncodeDataset(), nil
	}
}

func encodeImplicitVRDataset(dataset *Dataset) []byte {
	var result []byte

	var tags []Tag
	for tag := range dataset.Elements {
		tags = append(tags, tag)
	}

	for i := 0; i < len(tags)-1; i++ {
		for j := i + 1; j < len(tags); j++ {
			if tags[i].Group > tags[j].Group ||
				(tags[i].Group == tags[j].Group && tags[i].Element > tags[j].Element) {
				tags[i], tags[j] = tags[j], tags[i]
			}
		}
	}

	for _, tag := range tags {
		element := dataset.Elements[tag]

		tagBytes := make([]byte, 4)
		binary.LittleEndian.PutUint16(tagBytes[0:2], tag.Group)
		binary.LittleEndian.PutUint16(tagBytes[2:4], tag.Element)
		result = append(result, tagBytes...)

		valueBytes := encodeElementValue(element)
		if len(valueBytes)%2 == 1 {
			valueBytes = append(valueBytes, 0x20)
		}

		lengthBytes := make([]byte, 4)
		binary.LittleEndian.PutUint32(lengthBytes, uint32(len(valueBytes)))
		result = append(result, lengthBytes...)
		result = append(result, valueBytes...)
	}

	return result
}

// encodeElementValue encodes an element value to bytes
func encodeElementValue(element *Element) []byte {
	switch v := element.Value.(type) {
	case string:
		// For string VRs, ensure proper encoding
		value := v
		// Remove any existing null terminators and add proper padding
		value = strings.TrimRight(value, "\x00")
		return []byte(value)
	case []string:
		joined := strings.Join(v, "\\")
		joined = strings.TrimRight(joined, "\x00")
		return []byte(joined)
	case int:
		return []byte(fmt.Sprintf("%d", v))
	case uint16:
		result := make([]byte, 2)
		binary.LittleEndian.PutUint16(result, v)
		return result
	case uint32:
		result := make([]byte, 4)
		binary.LittleEndian.PutUint32(result, v)
		return result
	default:
		return []byte(fmt.Sprintf("%v", v))
	}
}
