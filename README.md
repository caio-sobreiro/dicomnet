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

## License

[Add your license here]
