package types

import "testing"

func TestGetTransferSyntaxInfo(t *testing.T) {
	tests := []struct {
		name           string
		uid            string
		wantName       string
		wantCompressed bool
		wantLossless   bool
		wantRetired    bool
	}{
		{
			name:           "Implicit VR Little Endian",
			uid:            ImplicitVRLittleEndian,
			wantName:       "Implicit VR Little Endian",
			wantCompressed: false,
			wantLossless:   true,
			wantRetired:    false,
		},
		{
			name:           "Explicit VR Little Endian",
			uid:            ExplicitVRLittleEndian,
			wantName:       "Explicit VR Little Endian",
			wantCompressed: false,
			wantLossless:   true,
			wantRetired:    false,
		},
		{
			name:           "Explicit VR Big Endian (retired)",
			uid:            ExplicitVRBigEndian,
			wantName:       "Explicit VR Big Endian",
			wantCompressed: false,
			wantLossless:   true,
			wantRetired:    true,
		},
		{
			name:           "JPEG 2000 Lossless",
			uid:            JPEG2000Lossless,
			wantName:       "JPEG 2000 Lossless Only",
			wantCompressed: true,
			wantLossless:   true,
			wantRetired:    false,
		},
		{
			name:           "JPEG 2000 Lossy",
			uid:            JPEG2000,
			wantName:       "JPEG 2000",
			wantCompressed: true,
			wantLossless:   false,
			wantRetired:    false,
		},
		{
			name:           "JPEG Baseline",
			uid:            JPEGBaseline8Bit,
			wantName:       "JPEG Baseline (Process 1)",
			wantCompressed: true,
			wantLossless:   false,
			wantRetired:    false,
		},
		{
			name:           "JPEG Lossless SV1",
			uid:            JPEGLosslessSV1,
			wantName:       "JPEG Lossless, Non-Hierarchical, First-Order Prediction",
			wantCompressed: true,
			wantLossless:   true,
			wantRetired:    false,
		},
		{
			name:           "RLE Lossless",
			uid:            RLELossless,
			wantName:       "RLE Lossless",
			wantCompressed: true,
			wantLossless:   true,
			wantRetired:    false,
		},
		{
			name:           "Unknown Transfer Syntax",
			uid:            "1.2.3.4.5.6.7.8.9",
			wantName:       "Unknown",
			wantCompressed: false,
			wantLossless:   true,
			wantRetired:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := GetTransferSyntaxInfo(tt.uid)

			if info.Name != tt.wantName {
				t.Errorf("GetTransferSyntaxInfo(%s).Name = %s, want %s",
					tt.uid, info.Name, tt.wantName)
			}
			if info.IsCompressed != tt.wantCompressed {
				t.Errorf("GetTransferSyntaxInfo(%s).IsCompressed = %v, want %v",
					tt.uid, info.IsCompressed, tt.wantCompressed)
			}
			if info.IsLossless != tt.wantLossless {
				t.Errorf("GetTransferSyntaxInfo(%s).IsLossless = %v, want %v",
					tt.uid, info.IsLossless, tt.wantLossless)
			}
			if info.IsRetired != tt.wantRetired {
				t.Errorf("GetTransferSyntaxInfo(%s).IsRetired = %v, want %v",
					tt.uid, info.IsRetired, tt.wantRetired)
			}
			if info.UID != tt.uid {
				t.Errorf("GetTransferSyntaxInfo(%s).UID = %s, want %s",
					tt.uid, info.UID, tt.uid)
			}
		})
	}
}

func TestIsCompressed(t *testing.T) {
	tests := []struct {
		name string
		uid  string
		want bool
	}{
		{"Implicit VR", ImplicitVRLittleEndian, false},
		{"Explicit VR", ExplicitVRLittleEndian, false},
		{"Explicit VR Big Endian", ExplicitVRBigEndian, false},
		{"Deflated", DeflatedExplicitVRLittleEndian, true},
		{"JPEG Baseline", JPEGBaseline8Bit, true},
		{"JPEG Lossless", JPEGLossless, true},
		{"JPEG 2000 Lossless", JPEG2000Lossless, true},
		{"JPEG 2000", JPEG2000, true},
		{"JPEG-LS Lossless", JPEGLSLossless, true},
		{"RLE", RLELossless, true},
		{"MPEG2", MPEG2MainProfile, true},
		{"H.264", MPEG4AVCH264HighProfile, true},
		{"H.265", HEVCH265MainProfileLevel51, true},
		{"HTJ2K Lossless", HTJ2KLossless, true},
		{"Unknown", "1.2.3.4.5", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsCompressed(tt.uid)
			if got != tt.want {
				t.Errorf("IsCompressed(%s) = %v, want %v", tt.uid, got, tt.want)
			}
		})
	}
}

func TestIsLossless(t *testing.T) {
	tests := []struct {
		name string
		uid  string
		want bool
	}{
		// Uncompressed (all lossless)
		{"Implicit VR", ImplicitVRLittleEndian, true},
		{"Explicit VR", ExplicitVRLittleEndian, true},
		{"Explicit VR Big Endian", ExplicitVRBigEndian, true},
		{"Deflated", DeflatedExplicitVRLittleEndian, true},

		// Lossless compression
		{"JPEG Lossless", JPEGLossless, true},
		{"JPEG Lossless SV1", JPEGLosslessSV1, true},
		{"JPEG 2000 Lossless", JPEG2000Lossless, true},
		{"JPEG-LS Lossless", JPEGLSLossless, true},
		{"RLE Lossless", RLELossless, true},
		{"HTJ2K Lossless", HTJ2KLossless, true},

		// Lossy compression
		{"JPEG Baseline", JPEGBaseline8Bit, false},
		{"JPEG Extended", JPEGExtended12Bit, false},
		{"JPEG 2000", JPEG2000, false},
		{"JPEG-LS Near-Lossless", JPEGLSNearLossless, false},
		{"MPEG2", MPEG2MainProfile, false},
		{"H.264", MPEG4AVCH264HighProfile, false},
		{"H.265", HEVCH265MainProfileLevel51, false},
		{"HTJ2K", HTJ2K, false},

		// Unknown (defaults to lossless)
		{"Unknown", "1.2.3.4.5", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsLossless(tt.uid)
			if got != tt.want {
				t.Errorf("IsLossless(%s) = %v, want %v", tt.uid, got, tt.want)
			}
		})
	}
}

func TestIsRetired(t *testing.T) {
	tests := []struct {
		name string
		uid  string
		want bool
	}{
		{"Implicit VR", ImplicitVRLittleEndian, false},
		{"Explicit VR", ExplicitVRLittleEndian, false},
		{"Explicit VR Big Endian (retired)", ExplicitVRBigEndian, true},
		{"JPEG 2000", JPEG2000, false},
		{"Unknown", "1.2.3.4.5", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsRetired(tt.uid)
			if got != tt.want {
				t.Errorf("IsRetired(%s) = %v, want %v", tt.uid, got, tt.want)
			}
		})
	}
}

func TestTransferSyntaxConstants(t *testing.T) {
	// Verify all constants are properly defined
	syntaxes := []struct {
		name string
		uid  string
	}{
		// Uncompressed
		{"ImplicitVRLittleEndian", ImplicitVRLittleEndian},
		{"ExplicitVRLittleEndian", ExplicitVRLittleEndian},
		{"ExplicitVRBigEndian", ExplicitVRBigEndian},
		{"DeflatedExplicitVRLittleEndian", DeflatedExplicitVRLittleEndian},

		// JPEG Lossy
		{"JPEGBaseline8Bit", JPEGBaseline8Bit},
		{"JPEGExtended12Bit", JPEGExtended12Bit},

		// JPEG Lossless
		{"JPEGLossless", JPEGLossless},
		{"JPEGLosslessSV1", JPEGLosslessSV1},

		// JPEG 2000
		{"JPEG2000Lossless", JPEG2000Lossless},
		{"JPEG2000", JPEG2000},
		{"JPEG2000Part2MultiComponentLossless", JPEG2000Part2MultiComponentLossless},
		{"JPEG2000Part2MultiComponent", JPEG2000Part2MultiComponent},

		// JPEG-LS
		{"JPEGLSLossless", JPEGLSLossless},
		{"JPEGLSNearLossless", JPEGLSNearLossless},

		// RLE
		{"RLELossless", RLELossless},

		// MPEG
		{"MPEG2MainProfile", MPEG2MainProfile},
		{"MPEG4AVCH264HighProfile", MPEG4AVCH264HighProfile},
		{"HEVCH265MainProfileLevel51", HEVCH265MainProfileLevel51},

		// HTJ2K
		{"HTJ2KLossless", HTJ2KLossless},
		{"HTJ2K", HTJ2K},
	}

	for _, ts := range syntaxes {
		t.Run(ts.name, func(t *testing.T) {
			if ts.uid == "" {
				t.Errorf("%s is empty", ts.name)
			}
			// All DICOM transfer syntax UIDs should start with "1.2.840.10008"
			if len(ts.uid) < 13 || ts.uid[:13] != "1.2.840.10008" {
				t.Errorf("%s = %s, should start with 1.2.840.10008", ts.name, ts.uid)
			}
		})
	}
}

func TestGetCommonTransferSyntaxes(t *testing.T) {
	syntaxes := GetCommonTransferSyntaxes()

	if len(syntaxes) == 0 {
		t.Fatal("GetCommonTransferSyntaxes() returned empty list")
	}

	// Should include explicit and implicit VR
	foundExplicit := false
	foundImplicit := false

	for _, ts := range syntaxes {
		if ts == ExplicitVRLittleEndian {
			foundExplicit = true
		}
		if ts == ImplicitVRLittleEndian {
			foundImplicit = true
		}
	}

	if !foundExplicit {
		t.Error("GetCommonTransferSyntaxes() missing Explicit VR Little Endian")
	}
	if !foundImplicit {
		t.Error("GetCommonTransferSyntaxes() missing Implicit VR Little Endian")
	}

	// First should be Explicit VR (most widely supported)
	if syntaxes[0] != ExplicitVRLittleEndian {
		t.Errorf("GetCommonTransferSyntaxes()[0] = %s, want %s",
			syntaxes[0], ExplicitVRLittleEndian)
	}
}

func TestTransferSyntaxRegistry(t *testing.T) {
	// Verify that all constants in registry have proper metadata
	requiredUIDs := []string{
		ImplicitVRLittleEndian,
		ExplicitVRLittleEndian,
		JPEG2000Lossless,
		JPEGLosslessSV1,
		RLELossless,
	}

	for _, uid := range requiredUIDs {
		info := GetTransferSyntaxInfo(uid)
		if info.Name == "Unknown" {
			t.Errorf("Transfer syntax %s missing from registry", uid)
		}
		if info.Description == "" {
			t.Errorf("Transfer syntax %s missing description", uid)
		}
	}
}

func TestTransferSyntaxInfoCompleteness(t *testing.T) {
	// Test that all registered syntaxes have complete information
	for uid, info := range transferSyntaxRegistry {
		t.Run(info.Name, func(t *testing.T) {
			if info.UID != uid {
				t.Errorf("UID mismatch: registry key = %s, info.UID = %s", uid, info.UID)
			}
			if info.Name == "" {
				t.Error("Name is empty")
			}
			if info.Description == "" {
				t.Error("Description is empty")
			}
		})
	}
}

// Benchmark tests
func BenchmarkGetTransferSyntaxInfo(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GetTransferSyntaxInfo(JPEG2000Lossless)
	}
}

func BenchmarkIsCompressed(b *testing.B) {
	for i := 0; i < b.N; i++ {
		IsCompressed(JPEG2000Lossless)
	}
}

func BenchmarkIsLossless(b *testing.B) {
	for i := 0; i < b.N; i++ {
		IsLossless(JPEGBaseline8Bit)
	}
}
