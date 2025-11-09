# client - DICOM Client (SCU)

DICOM client implementation for establishing associations and performing SCU operations.

## Features

- **Association Management** - Connect/disconnect from DICOM servers
- **Presentation Context Negotiation** - Support for multiple SOP classes
- **C-STORE SCU** - Send DICOM instances to remote SCP

## Usage

### Establishing a Connection

```go
import "github.com/caio-sobreiro/dicomnet/client"

assoc, err := client.Connect("hostname:4242", client.Config{
    CallingAETitle: "CLIENT_AE",
    CalledAETitle:  "SERVER_AE",
    MaxPDULength:   16384,
    // Optional: Specify preferred transfer syntaxes
    PreferredTransferSyntaxes: []string{
        "1.2.840.10008.1.2.4.90", // JPEG 2000 Lossless
        "1.2.840.10008.1.2.1",    // Explicit VR Little Endian
        "1.2.840.10008.1.2",      // Implicit VR Little Endian
    },
})
if err != nil {
    log.Fatal(err)
}
defer assoc.Close()
```

### Sending C-STORE

```go
resp, err := assoc.SendCStore(&client.CStoreRequest{
    SOPClassUID:    "1.2.840.10008.5.1.4.1.1.2", // CT Image Storage
    SOPInstanceUID: "1.2.3.4.5.6.7.8.9",
    Data:           dicomDataset, // Raw DICOM dataset (no Part 10 header)
    MessageID:      1,
})
if err != nil {
    log.Fatal(err)
}
if resp.Status != 0x0000 {
    log.Printf("C-STORE failed with status: 0x%04X", resp.Status)
}
```

## Implementation Details

- Uses **Implicit VR Little Endian** for DIMSE commands
- Supports **Explicit VR Little Endian** for datasets (default)
- **Dynamic transfer syntax negotiation** - proposes preferred syntaxes first, falls back to standard formats
- Handles PDU fragmentation for large datasets
- Presentation contexts for CT, MR, and Secondary Capture
- Shared DIMSE utilities with server implementation (`dimse` package)

## Status

✅ Association establishment
✅ C-ECHO SCU
✅ C-FIND SCU
✅ C-STORE SCU
✅ C-MOVE SCU
✅ C-CANCEL SCU
⏳ C-GET SCU (future)
