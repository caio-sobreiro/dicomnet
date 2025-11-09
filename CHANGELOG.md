# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.4.0] - 2025-11-09

### Added
- **C-GET implementation** (both SCU and SCP) for retrieving instances on same association
  - Client-side: `client/get.go` with `SendCGet()` method
  - Server-side: `dimse/service.go` with `CGetResponder` interface for C-STORE sub-operations
  - Server automatically detects C-GET requests and provides appropriate responder
  - Sub-operation counter tracking (remaining, completed, failed, warning)
- C-GET SOP Class support in server:
  - Patient Root Query/Retrieve Information Model - GET (1.2.840.10008.5.1.4.1.2.1.3)
  - Study Root Query/Retrieve Information Model - GET (1.2.840.10008.5.1.4.1.2.2.3)
  - Patient/Study Only Query/Retrieve Information Model - GET (1.2.840.10008.5.1.4.1.2.3.3)
- Enhanced `dimse.createDIMSECommand()` to support:
  - MessageID field (0000,0110) for requests
  - AffectedSOPInstanceUID field (0000,1000) for C-STORE operations
- C-GET handler in sample_server with `handleCGetStreaming()` implementation
- Documentation updates with C-GET vs C-MOVE comparison table in README

### Changed
- Enhanced test helper `buildCommandDataset()` to support sub-operation counter fields
- Updated README with C-GET feature and architectural differences from C-MOVE

### Testing
- Comprehensive C-GET client tests with sub-operation counter validation
- Verified with Orthanc PACS - successfully retrieved 3 synthetic instances via C-GET

## [0.3.1] - 2025-11-09

### Added
- **Comprehensive SOP Class support** with 150+ SOP Class UID constants in `types/sopclass.go`
  - Storage SOP Classes: CT, MR, US, NM, PET, RT, CR, DX, XA, visible light, ophthalmic, etc.
  - Query/Retrieve: Study Root, Patient Root, Patient/Study Only (FIND, MOVE, GET)
  - Worklist: Modality Worklist, General Purpose Worklist
  - MPPS: Modality Performed Procedure Step
  - Other: Verification, Storage Commitment, Encapsulated Documents (PDF, CDA, STL, OBJ, MTL)
- Helper functions for SOP Class identification:
  - `GetSOPClassInfo(uid)` - Returns metadata (name, category) for any SOP Class UID
  - `IsStorageSOPClass(uid)` - Checks if UID is a storage SOP class
  - `IsQueryRetrieveSOPClass(uid)` - Checks if UID is a query/retrieve SOP class
- **Dynamic SOP Class negotiation** in client
  - New `SOPClasses []string` field in `client.Config` for customizable SOP Class proposal
  - Default to 38 commonly used SOP Classes for broad compatibility
  - Automatic presentation context generation based on configured SOP Classes
- Comprehensive documentation in `docs/SOP_CLASS_SUPPORT.md` with usage examples

## [0.3.1] - 2025-11-09

### Added
- **Services package** (`services/`) with reusable DICOM service implementations
  - `EchoService` - Complete C-ECHO verification service implementation
  - `Registry` - Generic service registry/router for dispatching DIMSE messages to handlers
  - `ResponseBuilder` and helper functions for creating standard DIMSE responses
    - `NewCEchoResponse`, `NewCFindPendingResponse`, `NewCFindSuccessResponse`, `NewCFindErrorResponse`
    - `NewCMoveSuccessResponse`, `NewCMovePendingResponse`, `NewCMoveErrorResponse`
    - `NewCStoreResponse`
  - Support for both single-response and streaming (multi-response) operations
  - Dynamic handler registration/unregistration
  - Comprehensive test coverage (34 tests)
  - Package documentation (README.md)
- **DICOM Part 10 utilities** (`dicom/part10.go`)
  - `StripPart10Header` - Extracts dataset from DICOM Part 10 files
  - `HasPart10Header` - Checks if data contains Part 10 header
  - Useful for preparing datasets for DIMSE operations
- `CreateErrorResponse` utility function for standardized error responses
- GitHub Copilot instructions documentation (`.github/copilot-instructions.md`)

### Fixed
- VR comment formatting in Part 10 test file

## [0.3.0] - 2025-11-09

### Added
- **C-MOVE implementation** with C-STORE sub-operations to destination AET
- **Dynamic transfer syntax negotiation** - proposes native format first, falls back to standard syntaxes
- **Sample DICOM SCP server** (`cmd/sample_server`) with full C-ECHO/C-FIND/C-MOVE support
- **Synthetic DICOM data generation** - create valid instances in memory without files
- **Docker Compose setup** with Orthanc PACS for realistic integration testing
- Shared C-STORE SCU implementation in `dimse` package (usable by both client and server)
- Transfer syntax detection from DICOM file meta information (0002,0010)
- `PreferredTransferSyntaxes` field in client.Config for per-connection customization
- Explicit VR Little Endian (1.2.840.10008.1.2.1) transfer syntax support
- Custom error types package with DICOM-specific errors (AssociationError, DIMSEError, TimeoutError, NetworkError, PDUError, AbortError)
- Timeout configuration for client (ConnectTimeout, ReadTimeout, WriteTimeout)
- Timeout configuration for server (WithReadTimeout, WithWriteTimeout options)
- Logger injection support for client via Config.Logger field
- C-CANCEL operation support for canceling pending C-FIND and C-MOVE operations
- Comprehensive test coverage for all new features

### Changed
- **BREAKING**: Refactored C-STORE encoding/sending logic into shared `dimse` package
  - Moved `SendCStore`, `SendDIMSEMessage`, `SendPDataTF` to `dimse/store.go`
  - Moved `EncodeCommand`, `DecodeCommand` to `dimse` package
  - Client now uses shared functions via `dimse.SendCStore()`
- C-ECHO and C-FIND updated to use shared DIMSE utilities
- Enhanced server with streaming operation support for multi-response operations
- Improved PDU fragmentation and reassembly with comprehensive logging
- Default transfer syntax is now Explicit VR Little Endian (1.2.1) when not specified
- Command encoding now handles optional fields correctly (required for C-CANCEL)
- Client now uses instance logger instead of global slog package
- All test files updated to include logger initialization

### Fixed
- C-STORE Priority field (0000,0700) now properly included in command (was missing, causing Orthanc to reject)
- PDU construction for fragmented data transfers
- DIMSE command encoding for messages without all optional fields

### Testing
- Switched to **Orthanc as test client** instead of own client implementation
- Orthanc catches more real-world issues due to stricter validation
- Removed internal integration test suite in favor of Docker Compose + Orthanc approach

## [0.2.1] - 2025-11-08

### Added
- Logger injection support for server via `WithLogger` option
- PDU layer now accepts logger parameter for consistent logging
- DIMSE service accepts logger parameter
- Parser functions accept logger parameter

### Changed
- All internal logging migrated from direct `slog` calls to injected logger instances
- Logger defaults to `slog.Default()` when not provided

### Fixed
- Documentation updates for logging configuration

## [0.2.0] - 2025-11-08

### Added
- MIT License
- C-ECHO client implementation
- C-FIND client implementation with query support
- C-STORE client implementation
- Client-side association management
- Streaming service handler interface for multi-response operations
- Comprehensive test coverage for client operations

### Changed
- Improved error handling in PDU layer
- Enhanced presentation context negotiation

## [0.1.0] - Initial Release

### Added
- Core DICOM networking protocol layers (PDU, DIMSE, Dataset)
- Server implementation with context-aware lifecycle management
- Support for C-ECHO, C-FIND, C-STORE, C-MOVE operations
- Implicit VR Little Endian transfer syntax support
- Service handler interfaces
- Basic dataset parsing and encoding
- Unit tests for core functionality

[0.4.0]: https://github.com/caio-sobreiro/dicomnet/releases/tag/v0.4.0
[0.3.1]: https://github.com/caio-sobreiro/dicomnet/releases/tag/v0.3.1
[0.3.0]: https://github.com/caio-sobreiro/dicomnet/releases/tag/v0.3.0
[0.2.1]: https://github.com/caio-sobreiro/dicomnet/releases/tag/v0.2.1
[0.2.0]: https://github.com/caio-sobreiro/dicomnet/releases/tag/v0.2.0
[0.1.0]: https://github.com/caio-sobreiro/dicomnet/releases/tag/v0.1.0
