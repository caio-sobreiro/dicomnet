package types

import "testing"

func TestGetSOPClassInfo(t *testing.T) {
	tests := []struct {
		name     string
		uid      string
		wantName string
		wantCat  string
	}{
		{
			name:     "CT Image Storage",
			uid:      CTImageStorage,
			wantName: "CT Image Storage",
			wantCat:  "Storage",
		},
		{
			name:     "MR Image Storage",
			uid:      MRImageStorage,
			wantName: "MR Image Storage",
			wantCat:  "Storage",
		},
		{
			name:     "Verification SOP Class",
			uid:      VerificationSOPClass,
			wantName: "Verification SOP Class",
			wantCat:  "Verification",
		},
		{
			name:     "Study Root FIND",
			uid:      StudyRootQueryRetrieveInformationModelFind,
			wantName: "Study Root Query/Retrieve - FIND",
			wantCat:  "Query/Retrieve",
		},
		{
			name:     "Unknown SOP Class",
			uid:      "1.2.3.4.5.6.7.8.9",
			wantName: "Unknown",
			wantCat:  "Unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := GetSOPClassInfo(tt.uid)
			if info.Name != tt.wantName {
				t.Errorf("GetSOPClassInfo(%s).Name = %s, want %s", tt.uid, info.Name, tt.wantName)
			}
			if info.Category != tt.wantCat {
				t.Errorf("GetSOPClassInfo(%s).Category = %s, want %s", tt.uid, info.Category, tt.wantCat)
			}
			if info.UID != tt.uid {
				t.Errorf("GetSOPClassInfo(%s).UID = %s, want %s", tt.uid, info.UID, tt.uid)
			}
		})
	}
}

func TestIsStorageSOPClass(t *testing.T) {
	tests := []struct {
		name string
		uid  string
		want bool
	}{
		{"CT Image Storage", CTImageStorage, true},
		{"MR Image Storage", MRImageStorage, true},
		{"Secondary Capture", SecondaryCaptureImageStorage, true},
		{"PET Image Storage", PETImageStorage, true},
		{"RT Dose Storage", RTDoseStorage, true},
		{"Verification", VerificationSOPClass, false},
		{"Study Root FIND", StudyRootQueryRetrieveInformationModelFind, false},
		{"Modality Worklist", ModalityWorklistInformationModelFind, false},
		{"Unknown", "1.2.3.4.5.6.7.8.9", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsStorageSOPClass(tt.uid)
			if got != tt.want {
				t.Errorf("IsStorageSOPClass(%s) = %v, want %v", tt.uid, got, tt.want)
			}
		})
	}
}

func TestIsQueryRetrieveSOPClass(t *testing.T) {
	tests := []struct {
		name string
		uid  string
		want bool
	}{
		{"Study Root FIND", StudyRootQueryRetrieveInformationModelFind, true},
		{"Study Root MOVE", StudyRootQueryRetrieveInformationModelMove, true},
		{"Study Root GET", StudyRootQueryRetrieveInformationModelGet, true},
		{"Patient Root FIND", PatientRootQueryRetrieveInformationModelFind, true},
		{"Patient Root MOVE", PatientRootQueryRetrieveInformationModelMove, true},
		{"Patient Root GET", PatientRootQueryRetrieveInformationModelGet, true},
		{"CT Image Storage", CTImageStorage, false},
		{"Verification", VerificationSOPClass, false},
		{"Unknown", "1.2.3.4.5.6.7.8.9", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsQueryRetrieveSOPClass(tt.uid)
			if got != tt.want {
				t.Errorf("IsQueryRetrieveSOPClass(%s) = %v, want %v", tt.uid, got, tt.want)
			}
		})
	}
}

func TestSOPClassConstants(t *testing.T) {
	// Verify that all constants are properly defined with expected format
	sopClasses := []struct {
		name string
		uid  string
	}{
		{"VerificationSOPClass", VerificationSOPClass},
		{"CTImageStorage", CTImageStorage},
		{"MRImageStorage", MRImageStorage},
		{"UltrasoundImageStorage", UltrasoundImageStorage},
		{"SecondaryCaptureImageStorage", SecondaryCaptureImageStorage},
		{"PETImageStorage", PETImageStorage},
		{"RTImageStorage", RTImageStorage},
		{"EnhancedCTImageStorage", EnhancedCTImageStorage},
		{"EnhancedMRImageStorage", EnhancedMRImageStorage},
		{"NuclearMedicineImageStorage", NuclearMedicineImageStorage},
		{"StudyRootQueryRetrieveInformationModelFind", StudyRootQueryRetrieveInformationModelFind},
		{"StudyRootQueryRetrieveInformationModelMove", StudyRootQueryRetrieveInformationModelMove},
		{"PatientRootQueryRetrieveInformationModelFind", PatientRootQueryRetrieveInformationModelFind},
		{"ModalityWorklistInformationModelFind", ModalityWorklistInformationModelFind},
		{"EncapsulatedPDFStorage", EncapsulatedPDFStorage},
	}

	for _, tc := range sopClasses {
		t.Run(tc.name, func(t *testing.T) {
			if tc.uid == "" {
				t.Errorf("%s is empty", tc.name)
			}
			// All standard DICOM UIDs should start with "1.2.840.10008"
			if len(tc.uid) < 13 || tc.uid[:13] != "1.2.840.10008" {
				t.Errorf("%s = %s, should start with 1.2.840.10008", tc.name, tc.uid)
			}
		})
	}
}
