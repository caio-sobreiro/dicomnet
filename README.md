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

## Logging

The library uses Go's standard `log/slog` package for logging. You can inject a custom logger when creating a server using the `WithLogger` option:

```go
import "log/slog"

logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
server := server.New("MY_AE_TITLE", handler, server.WithLogger(logger))
```

If no logger is provided, the library will use `slog.Default()`.

## Usage

Currently used by the DIMSE proxy server in this repository. Will be published as a standalone library once client implementation is complete.

### Server Example

```go
import (
    "log/slog"
    "github.com/caio-sobreiro/dicomnet/server"
)

// Create server with your handler
srv := server.New("YOUR_AE_TITLE", yourServiceHandler, server.WithLogger(logger))

// Start listening
err := srv.ListenAndServe(ctx, ":11112", "YOUR_AE_TITLE", handler)
```

## Features

### Transfer Syntaxes
- ✅ Implicit VR Little Endian (1.2.840.10008.1.2)
- ✅ Explicit VR Little Endian (1.2.840.10008.1.2.1)

### DIMSE Operations
- ✅ C-ECHO (verification)
- ✅ C-FIND (query/retrieve)
- ✅ C-STORE (storage)
- ✅ C-MOVE (retrieve with sub-operations)
- ✅ C-CANCEL (cancel pending operations)
- ⏳ C-GET (planned)

### Client Features
- ✅ Configurable timeouts (connect, read, write)
- ✅ Logger injection support
- ✅ Custom error types for better error handling

### Server Features
- ✅ Configurable timeouts (read, write)
- ✅ Logger injection support
- ✅ Streaming response support for C-FIND/C-MOVE
- ✅ Dynamic transfer syntax negotiation (proposes native format first)
- ✅ Sample server with synthetic DICOM data generation

## Sample Server

A sample DICOM SCP server is provided for testing and reference:

### Running with Docker Compose (recommended)

Test against Orthanc PACS:
```bash
docker compose up
```

This starts:
- **sample_server**: DICOM SCP with synthetic data (port 4242, internal)
- **orthanc**: Production PACS for validation (port 8080, exposed for UI/API)

Access Orthanc UI at http://localhost:8080 (user: `orthanc`, password: `orthanc`)

Test C-MOVE operation:
```bash
curl -u orthanc:orthanc -X POST http://localhost:8080/modalities/SAMPLE_SCP/move \
  -H "Content-Type: application/json" \
  -d '{"Level":"Study","Resources":[{"StudyInstanceUID":"1.2.840.999.999.1.1.1.1"}],"TargetAet":"ORTHANC"}'
```

### Running standalone

Generate synthetic data (no files needed):
```bash
go run ./cmd/sample_server --synthetic --port 4242
```

Or load from DICOM file:
```bash
go run ./cmd/sample_server --dicom path/to/file.dcm --port 4242
```

The server supports:
- C-ECHO (verification)
- C-FIND (study/series/instance queries)
- C-MOVE (retrieve with C-STORE sub-operations)

## Testing

### Unit Tests

Run all unit tests:
```bash
go test ./...
```

### Integration Testing

We use **Orthanc** (production PACS) as the test client instead of our own client implementation. This catches more real-world issues as Orthanc is stricter and more widely deployed than testing against our own (potentially more permissive) client.

## License

MIT License - see LICENSE file for details
