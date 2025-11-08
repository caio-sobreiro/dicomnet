package types

import "testing"

func TestDIMSECommandConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant uint16
		expected uint16
	}{
		{"C-STORE-RQ", CStoreRQ, 0x0001},
		{"C-STORE-RSP", CStoreRSP, 0x8001},
		{"C-FIND-RQ", CFindRQ, 0x0020},
		{"C-FIND-RSP", CFindRSP, 0x8020},
		{"C-MOVE-RQ", CMoveRQ, 0x0021},
		{"C-MOVE-RSP", CMoveRSP, 0x8021},
		{"C-ECHO-RQ", CEchoRQ, 0x0030},
		{"C-ECHO-RSP", CEchoRSP, 0x8030},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.expected {
				t.Errorf("%s = 0x%04x, want 0x%04x", tt.name, tt.constant, tt.expected)
			}
		})
	}
}

func TestDIMSEStatusConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant uint16
		expected uint16
	}{
		{"Success", StatusSuccess, 0x0000},
		{"Pending", StatusPending, 0xFF00},
		{"Failure", StatusFailure, 0xC000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.expected {
				t.Errorf("Status%s = 0x%04x, want 0x%04x", tt.name, tt.constant, tt.expected)
			}
		})
	}
}

func TestMessage_Creation(t *testing.T) {
	tests := []struct {
		name    string
		message Message
	}{
		{
			name: "C-FIND Request",
			message: Message{
				CommandField:              CFindRQ,
				MessageID:                 1,
				AffectedSOPClassUID:       "1.2.840.10008.5.1.4.1.2.1.1",
				CommandDataSetType:        0x0001,
				Status:                    0,
				MessageIDBeingRespondedTo: 0,
			},
		},
		{
			name: "C-FIND Response Success",
			message: Message{
				CommandField:              CFindRSP,
				MessageID:                 0,
				AffectedSOPClassUID:       "1.2.840.10008.5.1.4.1.2.1.1",
				CommandDataSetType:        0x0000,
				Status:                    StatusSuccess,
				MessageIDBeingRespondedTo: 1,
			},
		},
		{
			name: "C-FIND Response Pending",
			message: Message{
				CommandField:              CFindRSP,
				MessageID:                 0,
				CommandDataSetType:        0x0001,
				Status:                    StatusPending,
				MessageIDBeingRespondedTo: 1,
			},
		},
		{
			name: "C-ECHO Request",
			message: Message{
				CommandField:        CEchoRQ,
				MessageID:           2,
				AffectedSOPClassUID: "1.2.840.10008.1.1",
				CommandDataSetType:  0x0101,
			},
		},
		{
			name: "C-ECHO Response",
			message: Message{
				CommandField:              CEchoRSP,
				Status:                    StatusSuccess,
				CommandDataSetType:        0x0101,
				MessageIDBeingRespondedTo: 2,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := tt.message

			// Verify command field is set
			if msg.CommandField == 0 {
				t.Error("CommandField should be set")
			}

			// Verify response messages have MessageIDBeingRespondedTo
			if msg.CommandField&0x8000 != 0 && msg.MessageIDBeingRespondedTo == 0 {
				// Response but no MessageIDBeingRespondedTo - only warn if it's not the first test case
				if tt.name != "C-FIND Response Success" {
					t.Logf("Warning: Response message should typically have MessageIDBeingRespondedTo")
				}
			}
		})
	}
}

func TestMessage_IsRequest(t *testing.T) {
	tests := []struct {
		name         string
		commandField uint16
		isRequest    bool
	}{
		{"C-FIND Request", CFindRQ, true},
		{"C-FIND Response", CFindRSP, false},
		{"C-ECHO Request", CEchoRQ, true},
		{"C-ECHO Response", CEchoRSP, false},
		{"C-STORE Request", CStoreRQ, true},
		{"C-STORE Response", CStoreRSP, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Response commands have bit 15 set (0x8000)
			isResponse := tt.commandField&0x8000 != 0
			isRequest := !isResponse

			if isRequest != tt.isRequest {
				t.Errorf("Command 0x%04x isRequest = %v, want %v",
					tt.commandField, isRequest, tt.isRequest)
			}
		})
	}
}

func TestMessage_ZeroValues(t *testing.T) {
	msg := &Message{}

	if msg.CommandField != 0 {
		t.Errorf("Zero Message CommandField = 0x%04x, want 0x0000", msg.CommandField)
	}
	if msg.MessageID != 0 {
		t.Errorf("Zero Message MessageID = %d, want 0", msg.MessageID)
	}
	if msg.AffectedSOPClassUID != "" {
		t.Errorf("Zero Message AffectedSOPClassUID = %q, want empty", msg.AffectedSOPClassUID)
	}
	if msg.Status != 0 {
		t.Errorf("Zero Message Status = 0x%04x, want 0x0000", msg.Status)
	}
}
