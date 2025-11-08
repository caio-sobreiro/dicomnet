package types

import (
	"testing"
)

func TestTag_String(t *testing.T) {
	tests := []struct {
		name     string
		tag      Tag
		expected string
	}{
		{
			name:     "Standard tag",
			tag:      Tag{Group: 0x0010, Element: 0x0010},
			expected: "(0010,0010)",
		},
		{
			name:     "Zero tag",
			tag:      Tag{Group: 0x0000, Element: 0x0000},
			expected: "(0000,0000)",
		},
		{
			name:     "High value tag",
			tag:      Tag{Group: 0xFFFF, Element: 0xFFFF},
			expected: "(ffff,ffff)",
		},
		{
			name:     "Command group tag",
			tag:      Tag{Group: 0x0000, Element: 0x0100},
			expected: "(0000,0100)",
		},
		{
			name:     "Patient name tag",
			tag:      Tag{Group: 0x0010, Element: 0x0010},
			expected: "(0010,0010)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.tag.String()
			if result != tt.expected {
				t.Errorf("Tag.String() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestElement_Creation(t *testing.T) {
	tests := []struct {
		name    string
		element Element
	}{
		{
			name: "String element",
			element: Element{
				Tag:   Tag{Group: 0x0010, Element: 0x0010},
				VR:    VR_PN,
				Value: "Doe^John",
			},
		},
		{
			name: "Integer element",
			element: Element{
				Tag:   Tag{Group: 0x0020, Element: 0x0010},
				VR:    VR_IS,
				Value: 12345,
			},
		},
		{
			name: "UID element",
			element: Element{
				Tag:   Tag{Group: 0x0008, Element: 0x0018},
				VR:    VR_UI,
				Value: "1.2.840.10008.1.1",
			},
		},
		{
			name: "Nil value element",
			element: Element{
				Tag:   Tag{Group: 0x0010, Element: 0x0020},
				VR:    VR_LO,
				Value: nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify element fields are set correctly
			if tt.element.Tag.Group == 0 && tt.element.Tag.Element == 0 {
				t.Error("Element tag not initialized")
			}
			if tt.element.VR == "" {
				t.Error("Element VR not set")
			}
		})
	}
}

func TestDataset_Creation(t *testing.T) {
	ds := &Dataset{
		Elements: make(map[Tag]*Element),
	}

	if ds.Elements == nil {
		t.Error("Dataset elements map not initialized")
	}

	// Add an element
	tag := Tag{Group: 0x0010, Element: 0x0010}
	elem := &Element{
		Tag:   tag,
		VR:    VR_PN,
		Value: "Test^Patient",
	}
	ds.Elements[tag] = elem

	// Verify retrieval
	retrieved, exists := ds.Elements[tag]
	if !exists {
		t.Error("Element not found in dataset")
	}
	if retrieved.Value != "Test^Patient" {
		t.Errorf("Element value = %v, want %v", retrieved.Value, "Test^Patient")
	}
}

func TestVRConstants(t *testing.T) {
	tests := []struct {
		name string
		vr   string
		want string
	}{
		{"Application Entity", VR_AE, "AE"},
		{"Person Name", VR_PN, "PN"},
		{"Unique Identifier", VR_UI, "UI"},
		{"Date", VR_DA, "DA"},
		{"Time", VR_TM, "TM"},
		{"Long String", VR_LO, "LO"},
		{"Short String", VR_SH, "SH"},
		{"Code String", VR_CS, "CS"},
		{"Unsigned Short", VR_US, "US"},
		{"Signed Long", VR_SL, "SL"},
		{"Sequence", VR_SQ, "SQ"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.vr != tt.want {
				t.Errorf("VR constant %s = %q, want %q", tt.name, tt.vr, tt.want)
			}
		})
	}
}

func TestTag_Equality(t *testing.T) {
	tag1 := Tag{Group: 0x0010, Element: 0x0010}
	tag2 := Tag{Group: 0x0010, Element: 0x0010}
	tag3 := Tag{Group: 0x0010, Element: 0x0020}

	if tag1 != tag2 {
		t.Error("Equal tags should be equal")
	}
	if tag1 == tag3 {
		t.Error("Different tags should not be equal")
	}
}

func TestDataset_MultipleElements(t *testing.T) {
	ds := &Dataset{
		Elements: make(map[Tag]*Element),
	}

	// Add multiple elements
	elements := []struct {
		tag   Tag
		vr    string
		value interface{}
	}{
		{Tag{0x0010, 0x0010}, VR_PN, "Doe^John"},
		{Tag{0x0010, 0x0020}, VR_LO, "12345"},
		{Tag{0x0008, 0x0060}, VR_CS, "CT"},
		{Tag{0x0020, 0x000D}, VR_UI, "1.2.840.113619.2.1"},
	}

	for _, e := range elements {
		ds.Elements[e.tag] = &Element{
			Tag:   e.tag,
			VR:    e.vr,
			Value: e.value,
		}
	}

	if len(ds.Elements) != 4 {
		t.Errorf("Dataset should have 4 elements, got %d", len(ds.Elements))
	}

	// Verify each element
	for _, e := range elements {
		elem, exists := ds.Elements[e.tag]
		if !exists {
			t.Errorf("Element %s not found", e.tag.String())
			continue
		}
		if elem.VR != e.vr {
			t.Errorf("Element %s VR = %s, want %s", e.tag.String(), elem.VR, e.vr)
		}
		if elem.Value != e.value {
			t.Errorf("Element %s value = %v, want %v", e.tag.String(), elem.Value, e.value)
		}
	}
}
