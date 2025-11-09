package types

// DICOM Application Context UID
// The Application Context defines the DICOM application-level message exchange rules.
const ApplicationContextUID = "1.2.840.10008.3.1.1.1"

// DICOM SOP Class UIDs as defined in DICOM Part 4, Annex B
// https://dicom.nema.org/medical/dicom/current/output/chtml/part04/sect_B.5.html

// Verification Service
const (
	VerificationSOPClass = "1.2.840.10008.1.1"
)

// Storage Service - Image Storage SOP Classes
const (
	// Computed Radiography
	ComputedRadiographyImageStorage = "1.2.840.10008.5.1.4.1.1.1"

	// Digital Radiography
	DigitalXRayImageStorageForPresentation            = "1.2.840.10008.5.1.4.1.1.1.1"
	DigitalXRayImageStorageForProcessing              = "1.2.840.10008.5.1.4.1.1.1.1.1"
	DigitalMammographyXRayImageStorageForPresentation = "1.2.840.10008.5.1.4.1.1.1.2"
	DigitalMammographyXRayImageStorageForProcessing   = "1.2.840.10008.5.1.4.1.1.1.2.1"
	DigitalIntraOralXRayImageStorageForPresentation   = "1.2.840.10008.5.1.4.1.1.1.3"
	DigitalIntraOralXRayImageStorageForProcessing     = "1.2.840.10008.5.1.4.1.1.1.3.1"

	// Computed Tomography
	CTImageStorage                        = "1.2.840.10008.5.1.4.1.1.2"
	EnhancedCTImageStorage                = "1.2.840.10008.5.1.4.1.1.2.1"
	LegacyConvertedEnhancedCTImageStorage = "1.2.840.10008.5.1.4.1.1.2.2"

	// Ultrasound
	UltrasoundMultiFrameImageStorage = "1.2.840.10008.5.1.4.1.1.3.1"
	UltrasoundImageStorage           = "1.2.840.10008.5.1.4.1.1.6.1"
	EnhancedUSVolumeStorage          = "1.2.840.10008.5.1.4.1.1.6.2"

	// Magnetic Resonance
	MRImageStorage                        = "1.2.840.10008.5.1.4.1.1.4"
	EnhancedMRImageStorage                = "1.2.840.10008.5.1.4.1.1.4.1"
	MRSpectroscopyStorage                 = "1.2.840.10008.5.1.4.1.1.4.2"
	EnhancedMRColorImageStorage           = "1.2.840.10008.5.1.4.1.1.4.3"
	LegacyConvertedEnhancedMRImageStorage = "1.2.840.10008.5.1.4.1.1.4.4"

	// Nuclear Medicine
	NuclearMedicineImageStorage = "1.2.840.10008.5.1.4.1.1.20"

	// Secondary Capture and Multi-frame
	SecondaryCaptureImageStorage                        = "1.2.840.10008.5.1.4.1.1.7"
	MultiFrameGrayscaleByteSecondaryCaptureImageStorage = "1.2.840.10008.5.1.4.1.1.7.1"
	MultiFrameGrayscaleWordSecondaryCaptureImageStorage = "1.2.840.10008.5.1.4.1.1.7.2"
	MultiFrameTrueColorSecondaryCaptureImageStorage     = "1.2.840.10008.5.1.4.1.1.7.3"
	MultiFrameSingleBitSecondaryCaptureImageStorage     = "1.2.840.10008.5.1.4.1.1.7.4"

	// X-Ray Angiographic
	XRayAngiographicImageStorage      = "1.2.840.10008.5.1.4.1.1.12.1"
	EnhancedXAImageStorage            = "1.2.840.10008.5.1.4.1.1.12.1.1"
	XRayRadiofluoroscopicImageStorage = "1.2.840.10008.5.1.4.1.1.12.2"
	EnhancedXRFImageStorage           = "1.2.840.10008.5.1.4.1.1.12.2.1"

	// X-Ray 3D
	XRay3DAngiographicImageStorage                  = "1.2.840.10008.5.1.4.1.1.13.1.1"
	XRay3DCraniofacialImageStorage                  = "1.2.840.10008.5.1.4.1.1.13.1.2"
	BreastTomosynthesisImageStorage                 = "1.2.840.10008.5.1.4.1.1.13.1.3"
	BreastProjectionXRayImageStorageForPresentation = "1.2.840.10008.5.1.4.1.1.13.1.4"
	BreastProjectionXRayImageStorageForProcessing   = "1.2.840.10008.5.1.4.1.1.13.1.5"

	// Intravascular Optical Coherence Tomography
	IntravascularOpticalCoherenceTomographyImageStorageForPresentation = "1.2.840.10008.5.1.4.1.1.14.1"
	IntravascularOpticalCoherenceTomographyImageStorageForProcessing   = "1.2.840.10008.5.1.4.1.1.14.2"

	// Positron Emission Tomography
	PETImageStorage                        = "1.2.840.10008.5.1.4.1.1.128"
	EnhancedPETImageStorage                = "1.2.840.10008.5.1.4.1.1.130"
	LegacyConvertedEnhancedPETImageStorage = "1.2.840.10008.5.1.4.1.1.128.1"

	// RT (Radiation Therapy)
	RTImageStorage                   = "1.2.840.10008.5.1.4.1.1.481.1"
	RTDoseStorage                    = "1.2.840.10008.5.1.4.1.1.481.2"
	RTStructureSetStorage            = "1.2.840.10008.5.1.4.1.1.481.3"
	RTBeamsTreatmentRecordStorage    = "1.2.840.10008.5.1.4.1.1.481.4"
	RTPlanStorage                    = "1.2.840.10008.5.1.4.1.1.481.5"
	RTBrachyTreatmentRecordStorage   = "1.2.840.10008.5.1.4.1.1.481.6"
	RTTreatmentSummaryRecordStorage  = "1.2.840.10008.5.1.4.1.1.481.7"
	RTIonPlanStorage                 = "1.2.840.10008.5.1.4.1.1.481.8"
	RTIonBeamsTreatmentRecordStorage = "1.2.840.10008.5.1.4.1.1.481.9"

	// Visible Light
	VLEndoscopicImageStorage                  = "1.2.840.10008.5.1.4.1.1.77.1.1"
	VLMicroscopicImageStorage                 = "1.2.840.10008.5.1.4.1.1.77.1.2"
	VLSlideCoordinatesMicroscopicImageStorage = "1.2.840.10008.5.1.4.1.1.77.1.3"
	VLPhotographicImageStorage                = "1.2.840.10008.5.1.4.1.1.77.1.4"
	VLWholeSlideMicroscopyImageStorage        = "1.2.840.10008.5.1.4.1.1.77.1.6"

	// Ophthalmic
	OphthalmicPhotography8BitImageStorage                             = "1.2.840.10008.5.1.4.1.1.77.1.5.1"
	OphthalmicPhotography16BitImageStorage                            = "1.2.840.10008.5.1.4.1.1.77.1.5.2"
	OphthalmicTomographyImageStorage                                  = "1.2.840.10008.5.1.4.1.1.77.1.5.4"
	WideFieldOphthalmicPhotographyStereographicProjectionImageStorage = "1.2.840.10008.5.1.4.1.1.77.1.5.6"
	WideFieldOphthalmicPhotography3DCoordinatesImageStorage           = "1.2.840.10008.5.1.4.1.1.77.1.5.7"
	OphthalmicOpticalCoherenceTomographyEnFaceImageStorage            = "1.2.840.10008.5.1.4.1.1.77.1.5.8"
	OphthalmicOpticalCoherenceTomographyBscanVolumeAnalysisStorage    = "1.2.840.10008.5.1.4.1.1.77.1.5.9"

	// Encapsulated Documents
	EncapsulatedPDFStorage = "1.2.840.10008.5.1.4.1.1.104.1"
	EncapsulatedCDAStorage = "1.2.840.10008.5.1.4.1.1.104.2"
	EncapsulatedSTLStorage = "1.2.840.10008.5.1.4.1.1.104.3"
	EncapsulatedOBJStorage = "1.2.840.10008.5.1.4.1.1.104.4"
	EncapsulatedMTLStorage = "1.2.840.10008.5.1.4.1.1.104.5"
)

// Query/Retrieve Service SOP Classes
const (
	// Study Root Query/Retrieve
	StudyRootQueryRetrieveInformationModelFind = "1.2.840.10008.5.1.4.1.2.2.1"
	StudyRootQueryRetrieveInformationModelMove = "1.2.840.10008.5.1.4.1.2.2.2"
	StudyRootQueryRetrieveInformationModelGet  = "1.2.840.10008.5.1.4.1.2.2.3"

	// Patient Root Query/Retrieve
	PatientRootQueryRetrieveInformationModelFind = "1.2.840.10008.5.1.4.1.2.1.1"
	PatientRootQueryRetrieveInformationModelMove = "1.2.840.10008.5.1.4.1.2.1.2"
	PatientRootQueryRetrieveInformationModelGet  = "1.2.840.10008.5.1.4.1.2.1.3"

	// Patient/Study Only Query/Retrieve
	PatientStudyOnlyQueryRetrieveInformationModelFind = "1.2.840.10008.5.1.4.1.2.3.1"
	PatientStudyOnlyQueryRetrieveInformationModelMove = "1.2.840.10008.5.1.4.1.2.3.2"
	PatientStudyOnlyQueryRetrieveInformationModelGet  = "1.2.840.10008.5.1.4.1.2.3.3"

	// Composite Instance Root Retrieve
	CompositeInstanceRootRetrieveMove = "1.2.840.10008.5.1.4.1.2.4.2"
	CompositeInstanceRootRetrieveGet  = "1.2.840.10008.5.1.4.1.2.4.3"

	// Composite Instance Retrieve Without Bulk Data
	CompositeInstanceRetrieveWithoutBulkDataGet = "1.2.840.10008.5.1.4.1.2.5.3"

	// Defined Procedure Protocol Query/Retrieve
	DefinedProcedureProtocolInformationModelFind = "1.2.840.10008.5.1.4.20.1"
	DefinedProcedureProtocolInformationModelMove = "1.2.840.10008.5.1.4.20.2"
	DefinedProcedureProtocolInformationModelGet  = "1.2.840.10008.5.1.4.20.3"
)

// Worklist Management Service SOP Classes
const (
	ModalityWorklistInformationModelFind         = "1.2.840.10008.5.1.4.31"
	GeneralPurposeWorklistInformationModelFind   = "1.2.840.10008.5.1.4.32.1"
	GeneralPurposeScheduledProcedureStepSOPClass = "1.2.840.10008.5.1.4.32.2"
	GeneralPurposePerformedProcedureStepSOPClass = "1.2.840.10008.5.1.4.32.3"
)

// Modality Performed Procedure Step
const (
	ModalityPerformedProcedureStepSOPClass             = "1.2.840.10008.3.1.2.3.3"
	ModalityPerformedProcedureStepRetrieveSOPClass     = "1.2.840.10008.3.1.2.3.4"
	ModalityPerformedProcedureStepNotificationSOPClass = "1.2.840.10008.3.1.2.3.5"
)

// Storage Commitment
const (
	StorageCommitmentPushModelSOPClass = "1.2.840.10008.1.20.1"
	StorageCommitmentPullModelSOPClass = "1.2.840.10008.1.20.2"
)

// Unified Procedure Step
const (
	UnifiedProcedureStepPushSOPClass  = "1.2.840.10008.5.1.4.34.6.1"
	UnifiedProcedureStepWatchSOPClass = "1.2.840.10008.5.1.4.34.6.2"
	UnifiedProcedureStepPullSOPClass  = "1.2.840.10008.5.1.4.34.6.3"
	UnifiedProcedureStepEventSOPClass = "1.2.840.10008.5.1.4.34.6.4"
	UnifiedProcedureStepQuerySOPClass = "1.2.840.10008.5.1.4.34.6.5"
)

// Hanging Protocol
const (
	HangingProtocolStorage              = "1.2.840.10008.5.1.4.38.1"
	HangingProtocolInformationModelFind = "1.2.840.10008.5.1.4.38.2"
	HangingProtocolInformationModelMove = "1.2.840.10008.5.1.4.38.3"
	HangingProtocolInformationModelGet  = "1.2.840.10008.5.1.4.38.4"
)

// Color Palette
const (
	ColorPaletteStorage              = "1.2.840.10008.5.1.4.39.1"
	ColorPaletteInformationModelFind = "1.2.840.10008.5.1.4.39.2"
	ColorPaletteInformationModelMove = "1.2.840.10008.5.1.4.39.3"
	ColorPaletteInformationModelGet  = "1.2.840.10008.5.1.4.39.4"
)

// Implant Template
const (
	GenericImplantTemplateStorage               = "1.2.840.10008.5.1.4.43.1"
	GenericImplantTemplateInformationModelFind  = "1.2.840.10008.5.1.4.43.2"
	GenericImplantTemplateInformationModelMove  = "1.2.840.10008.5.1.4.43.3"
	GenericImplantTemplateInformationModelGet   = "1.2.840.10008.5.1.4.43.4"
	ImplantAssemblyTemplateStorage              = "1.2.840.10008.5.1.4.44.1"
	ImplantAssemblyTemplateInformationModelFind = "1.2.840.10008.5.1.4.44.2"
	ImplantAssemblyTemplateInformationModelMove = "1.2.840.10008.5.1.4.44.3"
	ImplantAssemblyTemplateInformationModelGet  = "1.2.840.10008.5.1.4.44.4"
	ImplantTemplateGroupStorage                 = "1.2.840.10008.5.1.4.45.1"
	ImplantTemplateGroupInformationModelFind    = "1.2.840.10008.5.1.4.45.2"
	ImplantTemplateGroupInformationModelMove    = "1.2.840.10008.5.1.4.45.3"
	ImplantTemplateGroupInformationModelGet     = "1.2.840.10008.5.1.4.45.4"
)

// SOPClassInfo provides human-readable information about a SOP Class UID
type SOPClassInfo struct {
	UID         string
	Name        string
	Category    string
	Description string
}

// GetSOPClassInfo returns information about a SOP Class UID
func GetSOPClassInfo(uid string) *SOPClassInfo {
	info, ok := sopClassRegistry[uid]
	if !ok {
		return &SOPClassInfo{
			UID:      uid,
			Name:     "Unknown",
			Category: "Unknown",
		}
	}
	return &info
}

// IsStorageSOPClass returns true if the UID is a storage SOP class
func IsStorageSOPClass(uid string) bool {
	info := GetSOPClassInfo(uid)
	return info.Category == "Storage"
}

// IsQueryRetrieveSOPClass returns true if the UID is a query/retrieve SOP class
func IsQueryRetrieveSOPClass(uid string) bool {
	info := GetSOPClassInfo(uid)
	return info.Category == "Query/Retrieve"
}

// sopClassRegistry maps SOP Class UIDs to their information
var sopClassRegistry = map[string]SOPClassInfo{
	// Verification
	VerificationSOPClass: {
		UID:      VerificationSOPClass,
		Name:     "Verification SOP Class",
		Category: "Verification",
	},

	// Computed Radiography
	ComputedRadiographyImageStorage: {
		UID:      ComputedRadiographyImageStorage,
		Name:     "Computed Radiography Image Storage",
		Category: "Storage",
	},

	// CT
	CTImageStorage: {
		UID:      CTImageStorage,
		Name:     "CT Image Storage",
		Category: "Storage",
	},
	EnhancedCTImageStorage: {
		UID:      EnhancedCTImageStorage,
		Name:     "Enhanced CT Image Storage",
		Category: "Storage",
	},

	// MR
	MRImageStorage: {
		UID:      MRImageStorage,
		Name:     "MR Image Storage",
		Category: "Storage",
	},
	EnhancedMRImageStorage: {
		UID:      EnhancedMRImageStorage,
		Name:     "Enhanced MR Image Storage",
		Category: "Storage",
	},

	// Ultrasound
	UltrasoundImageStorage: {
		UID:      UltrasoundImageStorage,
		Name:     "Ultrasound Image Storage",
		Category: "Storage",
	},
	UltrasoundMultiFrameImageStorage: {
		UID:      UltrasoundMultiFrameImageStorage,
		Name:     "Ultrasound Multi-frame Image Storage",
		Category: "Storage",
	},

	// Secondary Capture
	SecondaryCaptureImageStorage: {
		UID:      SecondaryCaptureImageStorage,
		Name:     "Secondary Capture Image Storage",
		Category: "Storage",
	},

	// Nuclear Medicine
	NuclearMedicineImageStorage: {
		UID:      NuclearMedicineImageStorage,
		Name:     "Nuclear Medicine Image Storage",
		Category: "Storage",
	},

	// PET
	PETImageStorage: {
		UID:      PETImageStorage,
		Name:     "PET Image Storage",
		Category: "Storage",
	},
	EnhancedPETImageStorage: {
		UID:      EnhancedPETImageStorage,
		Name:     "Enhanced PET Image Storage",
		Category: "Storage",
	},

	// RT
	RTImageStorage: {
		UID:      RTImageStorage,
		Name:     "RT Image Storage",
		Category: "Storage",
	},
	RTDoseStorage: {
		UID:      RTDoseStorage,
		Name:     "RT Dose Storage",
		Category: "Storage",
	},
	RTStructureSetStorage: {
		UID:      RTStructureSetStorage,
		Name:     "RT Structure Set Storage",
		Category: "Storage",
	},
	RTPlanStorage: {
		UID:      RTPlanStorage,
		Name:     "RT Plan Storage",
		Category: "Storage",
	},

	// Query/Retrieve - Study Root
	StudyRootQueryRetrieveInformationModelFind: {
		UID:      StudyRootQueryRetrieveInformationModelFind,
		Name:     "Study Root Query/Retrieve - FIND",
		Category: "Query/Retrieve",
	},
	StudyRootQueryRetrieveInformationModelMove: {
		UID:      StudyRootQueryRetrieveInformationModelMove,
		Name:     "Study Root Query/Retrieve - MOVE",
		Category: "Query/Retrieve",
	},
	StudyRootQueryRetrieveInformationModelGet: {
		UID:      StudyRootQueryRetrieveInformationModelGet,
		Name:     "Study Root Query/Retrieve - GET",
		Category: "Query/Retrieve",
	},

	// Query/Retrieve - Patient Root
	PatientRootQueryRetrieveInformationModelFind: {
		UID:      PatientRootQueryRetrieveInformationModelFind,
		Name:     "Patient Root Query/Retrieve - FIND",
		Category: "Query/Retrieve",
	},
	PatientRootQueryRetrieveInformationModelMove: {
		UID:      PatientRootQueryRetrieveInformationModelMove,
		Name:     "Patient Root Query/Retrieve - MOVE",
		Category: "Query/Retrieve",
	},
	PatientRootQueryRetrieveInformationModelGet: {
		UID:      PatientRootQueryRetrieveInformationModelGet,
		Name:     "Patient Root Query/Retrieve - GET",
		Category: "Query/Retrieve",
	},

	// Worklist
	ModalityWorklistInformationModelFind: {
		UID:      ModalityWorklistInformationModelFind,
		Name:     "Modality Worklist - FIND",
		Category: "Worklist",
	},

	// MPPS
	ModalityPerformedProcedureStepSOPClass: {
		UID:      ModalityPerformedProcedureStepSOPClass,
		Name:     "Modality Performed Procedure Step",
		Category: "MPPS",
	},

	// Storage Commitment
	StorageCommitmentPushModelSOPClass: {
		UID:      StorageCommitmentPushModelSOPClass,
		Name:     "Storage Commitment Push Model",
		Category: "Storage Commitment",
	},

	// Encapsulated Documents
	EncapsulatedPDFStorage: {
		UID:      EncapsulatedPDFStorage,
		Name:     "Encapsulated PDF Storage",
		Category: "Storage",
	},
	EncapsulatedCDAStorage: {
		UID:      EncapsulatedCDAStorage,
		Name:     "Encapsulated CDA Storage",
		Category: "Storage",
	},
}
