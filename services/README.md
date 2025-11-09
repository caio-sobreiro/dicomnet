# DICOM Services Package

This package provides reusable DICOM service implementations that follow the DICOM standard and have no external backend dependencies.

## Components

### EchoService

A complete implementation of the DICOM C-ECHO verification service (PS3.4). C-ECHO is used to verify connectivity and application-level communication between DICOM Application Entities.

**Features:**
- Stateless, no external dependencies
- Implements `interfaces.ServiceHandler`
- Full health check support

**Usage:**
```go
import "github.com/caio-sobreiro/dicomnet/services"

echoService := services.NewEchoService()
response, data, err := echoService.HandleDIMSE(ctx, msg, data)
```

### Registry

A flexible service registry/router that dispatches incoming DIMSE messages to appropriate service handlers based on command fields.

**Features:**
- Dynamic handler registration
- Support for both single-response and streaming handlers
- Automatic fallback for non-streaming handlers
- Command field routing

**Usage:**
```go
import (
    "github.com/caio-sobreiro/dicomnet/services"
    "github.com/caio-sobreiro/dicomnet/dimse"
)

// Create registry
registry := services.NewRegistry()

// Register handlers
registry.RegisterHandler(dimse.CEchoRQ, services.NewEchoService())
registry.RegisterHandler(dimse.CFindRQ, myFindService)

// Handle messages
response, data, err := registry.HandleDIMSE(ctx, msg, data)

// Or use streaming for multi-response operations
err := registry.HandleDIMSEStreaming(ctx, msg, data, responder)
```

## Migration from Local Implementations

The C-ECHO service has been moved from application-specific implementations to this reusable package. To migrate:

1. Replace `echo.Service` with `services.EchoService`
2. Replace `echo.New()` with `services.NewEchoService()`
3. Replace `.Handle()` with `.HandleDIMSE()`

## Design Principles

1. **No Backend Dependencies**: Services in this package are protocol-level implementations only
2. **Interface Compliance**: All services implement `interfaces.ServiceHandler`
3. **Optional Streaming**: Services can optionally implement `interfaces.StreamingServiceHandler`
4. **Stateless When Possible**: Prefer stateless designs for better scalability

## Future Additions

Additional DICOM services can be added to this package if they meet the criteria:
- Pure protocol implementation
- No external backend/database dependencies
- Follows DICOM standard specifications
- Reusable across different applications
