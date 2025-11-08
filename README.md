# dicomnet - DICOM Networking Core Library

Core DICOM protocol implementation (PDU, DIMSE, dataset parsing).

## Package Structure

- **`dicom/`** - DICOM dataset parsing and manipulation
- **`dimse/`** - DIMSE message handling (commands like C-STORE, C-FIND, C-MOVE, C-ECHO)
- **`pdu/`** - DICOM Upper Layer Protocol (association handling, PDU encoding/decoding)
- **`types/`** - Common types and data structures
- **`interfaces/`** - Public interfaces for service handlers

## Architecture

This library provides the core DICOM networking protocol layers:

```
Application Layer
         ↓
   DIMSE Layer (dimse/)
         ↓
  Upper Layer (pdu/)
         ↓
   Transport (TCP/IP)
```

## Usage

This package is used by:
- **`pkg/client/`** - DICOM client (SCU) implementation
- **`pkg/server/`** - DICOM server (SCP) implementation (future)
- **`internal/services/`** - Application-specific service handlers

## Usage

Currently used by the DIMSE proxy server in this repository. Will be published as a standalone library once client implementation is complete.

### Server Example

```go
import (
    "github.com/caio-sobreiro/dicomnet/pdu"
    "github.com/caio-sobreiro/dicomnet/dimse"
)

// Create DIMSE service with your handler
dimseService := dimse.NewService(yourServiceHandler)

// Create PDU layer
pduLayer := pdu.NewLayer(conn, dimseService, "YOUR_AE_TITLE")

// Handle connection
err := pduLayer.HandleConnection()
```

## License

[Add your license here]
