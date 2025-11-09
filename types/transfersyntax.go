package types

// DICOM Transfer Syntax UIDs as defined in DICOM Part 5, Section 8 and Part 6, Annex A.4
// https://dicom.nema.org/medical/dicom/current/output/chtml/part05/chapter_8.html

// Uncompressed Transfer Syntaxes
const (
	// ImplicitVRLittleEndian - Default Transfer Syntax for DICOM
	// Uses implicit VR encoding with little endian byte ordering
	ImplicitVRLittleEndian = "1.2.840.10008.1.2"

	// ExplicitVRLittleEndian - Explicit VR with little endian byte ordering
	// Recommended for general use due to explicit data types
	ExplicitVRLittleEndian = "1.2.840.10008.1.2.1"

	// ExplicitVRBigEndian - Explicit VR with big endian byte ordering (retired)
	// Rarely used, included for completeness
	ExplicitVRBigEndian = "1.2.840.10008.1.2.2"

	// DeflatedExplicitVRLittleEndian - Deflate compression with explicit VR
	// Uses zlib/deflate compression on top of explicit VR encoding
	DeflatedExplicitVRLittleEndian = "1.2.840.10008.1.2.1.99"
)

// JPEG Lossy Compression Transfer Syntaxes
const (
	// JPEGBaseline8Bit - JPEG Baseline (Process 1)
	// Default lossy JPEG compression, 8-bit samples
	JPEGBaseline8Bit = "1.2.840.10008.1.2.4.50"

	// JPEGExtended12Bit - JPEG Extended (Process 2 & 4)
	// Lossy JPEG compression, 8-12 bit samples
	JPEGExtended12Bit = "1.2.840.10008.1.2.4.51"

	// JPEGSpectralSelectionNonHierarchical68 - JPEG Extended (Process 3 & 5)
	JPEGSpectralSelectionNonHierarchical68 = "1.2.840.10008.1.2.4.52"

	// JPEGSpectralSelectionNonHierarchical79 - JPEG Spectral Selection (Process 6 & 8)
	JPEGSpectralSelectionNonHierarchical79 = "1.2.840.10008.1.2.4.53"

	// JPEGFullProgressionNonHierarchical1012 - JPEG Full Progression (Process 10 & 12)
	JPEGFullProgressionNonHierarchical1012 = "1.2.840.10008.1.2.4.54"

	// JPEGFullProgressionNonHierarchical1113 - JPEG Full Progression (Process 11 & 13)
	JPEGFullProgressionNonHierarchical1113 = "1.2.840.10008.1.2.4.55"
)

// JPEG Lossless Compression Transfer Syntaxes
const (
	// JPEGLossless - JPEG Lossless (Process 14)
	JPEGLossless = "1.2.840.10008.1.2.4.57"

	// JPEGLosslessSV1 - JPEG Lossless (Process 14, Selection Value 1)
	// Most commonly used lossless JPEG variant
	JPEGLosslessSV1 = "1.2.840.10008.1.2.4.70"

	// JPEGLosslessNonHierarchical1517 - JPEG Lossless (Process 15)
	JPEGLosslessNonHierarchical1517 = "1.2.840.10008.1.2.4.58"

	// JPEGLosslessNonHierarchical1618 - JPEG Lossless (Process 16)
	JPEGLosslessNonHierarchical1618 = "1.2.840.10008.1.2.4.59"
)

// JPEG 2000 Transfer Syntaxes
const (
	// JPEG2000Lossless - JPEG 2000 Image Compression (Lossless Only)
	// Modern lossless compression, better compression than JPEG lossless
	JPEG2000Lossless = "1.2.840.10008.1.2.4.90"

	// JPEG2000 - JPEG 2000 Image Compression (lossy or lossless)
	// Supports both lossy and lossless compression
	JPEG2000 = "1.2.840.10008.1.2.4.91"

	// JPEG2000Part2MultiComponentLossless - JPEG 2000 Part 2 Multi-component (Lossless)
	JPEG2000Part2MultiComponentLossless = "1.2.840.10008.1.2.4.92"

	// JPEG2000Part2MultiComponent - JPEG 2000 Part 2 Multi-component
	JPEG2000Part2MultiComponent = "1.2.840.10008.1.2.4.93"
)

// JPEG-LS Transfer Syntaxes
const (
	// JPEGLSLossless - JPEG-LS Lossless Image Compression
	// Lossless compression with good performance
	JPEGLSLossless = "1.2.840.10008.1.2.4.80"

	// JPEGLSNearLossless - JPEG-LS Lossy (Near-Lossless) Image Compression
	// Near-lossless with controlled error bounds
	JPEGLSNearLossless = "1.2.840.10008.1.2.4.81"
)

// RLE Transfer Syntax
const (
	// RLELossless - RLE Lossless Compression
	// Simple run-length encoding, lossless compression
	RLELossless = "1.2.840.10008.1.2.5"
)

// MPEG Video Transfer Syntaxes
const (
	// MPEG2MainProfile - MPEG2 Main Profile @ Main Level
	MPEG2MainProfile = "1.2.840.10008.1.2.4.100"

	// MPEG2MainProfileHighLevel - MPEG2 Main Profile @ High Level
	MPEG2MainProfileHighLevel = "1.2.840.10008.1.2.4.101"

	// MPEG4AVCH264HighProfile - MPEG-4 AVC/H.264 High Profile / Level 4.1
	MPEG4AVCH264HighProfile = "1.2.840.10008.1.2.4.102"

	// MPEG4AVCH264BDCompatibleHighProfile - MPEG-4 AVC/H.264 BD-compatible High Profile / Level 4.1
	MPEG4AVCH264BDCompatibleHighProfile = "1.2.840.10008.1.2.4.103"

	// MPEG4AVCH264HighProfileLevel42 - MPEG-4 AVC/H.264 High Profile / Level 4.2 For 2D Video
	MPEG4AVCH264HighProfileLevel42 = "1.2.840.10008.1.2.4.104"

	// MPEG4AVCH264HighProfileLevel42Stereo - MPEG-4 AVC/H.264 High Profile / Level 4.2 For 3D Video
	MPEG4AVCH264HighProfileLevel42Stereo = "1.2.840.10008.1.2.4.105"

	// MPEG4AVCH264StereoHighProfile - MPEG-4 AVC/H.264 Stereo High Profile / Level 4.2
	MPEG4AVCH264StereoHighProfile = "1.2.840.10008.1.2.4.106"

	// HEVCH265MainProfileLevel51 - HEVC/H.265 Main Profile / Level 5.1
	HEVCH265MainProfileLevel51 = "1.2.840.10008.1.2.4.107"

	// HEVCH265Main10ProfileLevel51 - HEVC/H.265 Main 10 Profile / Level 5.1
	HEVCH265Main10ProfileLevel51 = "1.2.840.10008.1.2.4.108"
)

// JPIP Transfer Syntaxes (Referenced and Deflate)
const (
	// JPIPReferenced - JPIP Referenced
	JPIPReferenced = "1.2.840.10008.1.2.4.94"

	// JPIPReferencedDeflate - JPIP Referenced Deflate
	JPIPReferencedDeflate = "1.2.840.10008.1.2.4.95"
)

// High-Throughput JPEG 2000 Transfer Syntaxes
const (
	// HTJ2KLossless - High-Throughput JPEG 2000 Image Compression (Lossless Only)
	HTJ2KLossless = "1.2.840.10008.1.2.4.201"

	// HTJ2KLosslessRPCL - High-Throughput JPEG 2000 with RPCL Options (Lossless Only)
	HTJ2KLosslessRPCL = "1.2.840.10008.1.2.4.202"

	// HTJ2K - High-Throughput JPEG 2000
	HTJ2K = "1.2.840.10008.1.2.4.203"
)

// TransferSyntaxInfo provides metadata about a transfer syntax
type TransferSyntaxInfo struct {
	UID                string
	Name               string
	IsCompressed       bool
	IsLossless         bool
	IsRetired          bool
	SupportsEncapsulated bool
	Description        string
}

// GetTransferSyntaxInfo returns information about a transfer syntax UID
func GetTransferSyntaxInfo(uid string) *TransferSyntaxInfo {
	info, ok := transferSyntaxRegistry[uid]
	if !ok {
		return &TransferSyntaxInfo{
			UID:          uid,
			Name:         "Unknown",
			IsCompressed: false,
			IsLossless:   true,
			Description:  "Unknown transfer syntax",
		}
	}
	return &info
}

// IsCompressed returns true if the transfer syntax uses compression
func IsCompressed(uid string) bool {
	info := GetTransferSyntaxInfo(uid)
	return info.IsCompressed
}

// IsLossless returns true if the transfer syntax is lossless
// Note: Uncompressed transfer syntaxes are considered lossless
func IsLossless(uid string) bool {
	info := GetTransferSyntaxInfo(uid)
	return info.IsLossless
}

// IsRetired returns true if the transfer syntax is retired
func IsRetired(uid string) bool {
	info := GetTransferSyntaxInfo(uid)
	return info.IsRetired
}

// transferSyntaxRegistry maps transfer syntax UIDs to their information
var transferSyntaxRegistry = map[string]TransferSyntaxInfo{
	// Uncompressed
	ImplicitVRLittleEndian: {
		UID:          ImplicitVRLittleEndian,
		Name:         "Implicit VR Little Endian",
		IsCompressed: false,
		IsLossless:   true,
		IsRetired:    false,
		Description:  "Default DICOM transfer syntax with implicit VR encoding",
	},
	ExplicitVRLittleEndian: {
		UID:          ExplicitVRLittleEndian,
		Name:         "Explicit VR Little Endian",
		IsCompressed: false,
		IsLossless:   true,
		IsRetired:    false,
		Description:  "Explicit VR encoding with little endian byte order",
	},
	ExplicitVRBigEndian: {
		UID:          ExplicitVRBigEndian,
		Name:         "Explicit VR Big Endian",
		IsCompressed: false,
		IsLossless:   true,
		IsRetired:    true,
		Description:  "Explicit VR encoding with big endian byte order (retired)",
	},
	DeflatedExplicitVRLittleEndian: {
		UID:                DeflatedExplicitVRLittleEndian,
		Name:               "Deflated Explicit VR Little Endian",
		IsCompressed:       true,
		IsLossless:         true,
		IsRetired:          false,
		SupportsEncapsulated: false,
		Description:        "Deflate/zlib compression with explicit VR encoding",
	},

	// JPEG Lossy
	JPEGBaseline8Bit: {
		UID:                JPEGBaseline8Bit,
		Name:               "JPEG Baseline (Process 1)",
		IsCompressed:       true,
		IsLossless:         false,
		IsRetired:          false,
		SupportsEncapsulated: true,
		Description:        "JPEG lossy compression, 8-bit samples",
	},
	JPEGExtended12Bit: {
		UID:                JPEGExtended12Bit,
		Name:               "JPEG Extended (Process 2 & 4)",
		IsCompressed:       true,
		IsLossless:         false,
		IsRetired:          false,
		SupportsEncapsulated: true,
		Description:        "JPEG lossy compression, 8-12 bit samples",
	},

	// JPEG Lossless
	JPEGLossless: {
		UID:                JPEGLossless,
		Name:               "JPEG Lossless (Process 14)",
		IsCompressed:       true,
		IsLossless:         true,
		IsRetired:          false,
		SupportsEncapsulated: true,
		Description:        "JPEG lossless compression",
	},
	JPEGLosslessSV1: {
		UID:                JPEGLosslessSV1,
		Name:               "JPEG Lossless, Non-Hierarchical, First-Order Prediction",
		IsCompressed:       true,
		IsLossless:         true,
		IsRetired:          false,
		SupportsEncapsulated: true,
		Description:        "JPEG lossless compression with prediction (most common)",
	},

	// JPEG 2000
	JPEG2000Lossless: {
		UID:                JPEG2000Lossless,
		Name:               "JPEG 2000 Lossless Only",
		IsCompressed:       true,
		IsLossless:         true,
		IsRetired:          false,
		SupportsEncapsulated: true,
		Description:        "JPEG 2000 lossless compression",
	},
	JPEG2000: {
		UID:                JPEG2000,
		Name:               "JPEG 2000",
		IsCompressed:       true,
		IsLossless:         false,
		IsRetired:          false,
		SupportsEncapsulated: true,
		Description:        "JPEG 2000 lossy or lossless compression",
	},

	// JPEG-LS
	JPEGLSLossless: {
		UID:                JPEGLSLossless,
		Name:               "JPEG-LS Lossless",
		IsCompressed:       true,
		IsLossless:         true,
		IsRetired:          false,
		SupportsEncapsulated: true,
		Description:        "JPEG-LS lossless compression",
	},
	JPEGLSNearLossless: {
		UID:                JPEGLSNearLossless,
		Name:               "JPEG-LS Near-Lossless",
		IsCompressed:       true,
		IsLossless:         false,
		IsRetired:          false,
		SupportsEncapsulated: true,
		Description:        "JPEG-LS near-lossless compression with bounded error",
	},

	// RLE
	RLELossless: {
		UID:                RLELossless,
		Name:               "RLE Lossless",
		IsCompressed:       true,
		IsLossless:         true,
		IsRetired:          false,
		SupportsEncapsulated: true,
		Description:        "Run-Length Encoding lossless compression",
	},

	// MPEG
	MPEG2MainProfile: {
		UID:                MPEG2MainProfile,
		Name:               "MPEG2 Main Profile @ Main Level",
		IsCompressed:       true,
		IsLossless:         false,
		IsRetired:          false,
		SupportsEncapsulated: true,
		Description:        "MPEG-2 video compression",
	},
	MPEG4AVCH264HighProfile: {
		UID:                MPEG4AVCH264HighProfile,
		Name:               "MPEG-4 AVC/H.264 High Profile",
		IsCompressed:       true,
		IsLossless:         false,
		IsRetired:          false,
		SupportsEncapsulated: true,
		Description:        "H.264 video compression",
	},
	HEVCH265MainProfileLevel51: {
		UID:                HEVCH265MainProfileLevel51,
		Name:               "HEVC/H.265 Main Profile",
		IsCompressed:       true,
		IsLossless:         false,
		IsRetired:          false,
		SupportsEncapsulated: true,
		Description:        "H.265/HEVC video compression",
	},

	// High-Throughput JPEG 2000
	HTJ2KLossless: {
		UID:                HTJ2KLossless,
		Name:               "High-Throughput JPEG 2000 Lossless",
		IsCompressed:       true,
		IsLossless:         true,
		IsRetired:          false,
		SupportsEncapsulated: true,
		Description:        "HTJ2K lossless compression (fast JPEG 2000 variant)",
	},
	HTJ2K: {
		UID:                HTJ2K,
		Name:               "High-Throughput JPEG 2000",
		IsCompressed:       true,
		IsLossless:         false,
		IsRetired:          false,
		SupportsEncapsulated: true,
		Description:        "HTJ2K lossy or lossless compression",
	},
}

// GetCommonTransferSyntaxes returns a list of commonly supported transfer syntaxes
// in recommended negotiation order (uncompressed first, then lossless, then lossy)
func GetCommonTransferSyntaxes() []string {
	return []string{
		ExplicitVRLittleEndian, // Most widely supported
		ImplicitVRLittleEndian, // Default DICOM
		JPEG2000Lossless,       // Modern lossless
		JPEGLosslessSV1,        // Traditional lossless
		RLELossless,            // Simple lossless
		JPEG2000,               // Modern lossy/lossless
		JPEGBaseline8Bit,       // Traditional lossy
	}
}
