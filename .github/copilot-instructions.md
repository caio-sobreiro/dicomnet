# GitHub Copilot Instructions for dicomnet

## Project Overview

This is a Go implementation of the DICOM networking protocol (DICOM Upper Layer Protocol). The project provides both client (SCU) and server (SCP) implementations with support for various DIMSE operations.

## Code Style and Conventions

### General Go Conventions
- Follow standard Go formatting (use `gofmt`)
- Use meaningful variable names that reflect DICOM terminology
- Add comments for exported functions and types
- Keep functions focused and single-purpose
- Prefer explicit error handling over panics

### DICOM-Specific Conventions
- Use DICOM UIDs from the standard (e.g., `1.2.840.10008.5.1.4.1.1.2` for CT Image Storage)
- Tag references should include both hex format and name in comments (e.g., `0x0000, 0x0100 // Command Field`)
- Transfer syntax UIDs should be defined as constants with descriptive names
- AE Titles should be uppercase and use underscores (e.g., `SAMPLE_SCP`)

### Package Structure
- `types/` - Core DICOM type definitions (PDU, DIMSE, query structures)
- `pdu/` - PDU layer implementation (encoding/decoding, network communication)
- `dimse/` - DIMSE service layer (command encoding, message handling)
- `dicom/` - DICOM dataset parsing and manipulation
- `client/` - SCU (client) implementations
- `server/` - SCP (server) implementations
- `errors/` - Custom error types for DICOM operations
- `interfaces/` - Shared interfaces
- `cmd/` - Command-line tools and sample servers

### Testing
- Unit tests should be in `*_test.go` files alongside the code
- Use table-driven tests for multiple test cases
- Mock network operations using in-memory buffers
- Include both success and failure scenarios
- Test files should initialize loggers (use `slog.New(slog.NewTextHandler(io.Discard, nil))` for quiet tests)

## Architecture Patterns

### PDU Layer
- Handles low-level PDU encoding/decoding
- Manages TCP connection lifecycle
- Fragments large datasets into P-DATA-TF PDUs
- Maximum PDU size is configurable (default 16384 bytes)

### DIMSE Layer
- Builds on PDU layer for DIMSE operations
- Commands use **Implicit VR Little Endian** encoding
- Datasets use **Explicit VR Little Endian** by default
- Shared utilities in `dimse/` package (e.g., `SendCStore`, `EncodeCommand`, `DecodeCommand`)

### Client Architecture
- `Association` struct manages connection state
- Each DIMSE operation (C-ECHO, C-FIND, C-STORE, C-MOVE, C-CANCEL) has dedicated methods
- Dynamic transfer syntax negotiation via `PreferredTransferSyntaxes` config field
- Proposes native transfer syntax first, falls back to standard syntaxes

### Server Architecture
- Handler interface for custom SCP implementations
- Streaming support for multi-response operations (C-FIND, C-MOVE)
- Server sends responses as they become available
- Configurable timeouts via `WithReadTimeout`, `WithWriteTimeout` options

## Common Patterns

### Error Handling
Use custom error types from `errors/` package:
```go
if err := operation(); err != nil {
    return errors.NewDIMSEError("operation failed", err)
}
```

### Logger Usage
All major components accept a logger:
```go
logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
assoc, err := client.Connect("host:port", client.Config{
    Logger: logger,
    // ... other config
})
```

### Transfer Syntax Handling
When working with DICOM files:
```go
// Detect transfer syntax from file meta information (0002,0010)
transferSyntax := dataset.GetString(0x0002, 0x0010)

// Configure client with preferred syntaxes
config.PreferredTransferSyntaxes = []string{
    transferSyntax,                    // Native format
    "1.2.840.10008.1.2.1",            // Explicit VR
    "1.2.840.10008.1.2",              // Implicit VR
}
```

## Development Workflow

### Running Tests
```bash
go test ./...                    # Run all tests
go test -v ./client              # Verbose output for specific package
go test -race ./...              # Run with race detector
```

### Local Testing with Docker
```bash
docker compose up -d             # Start sample server + Orthanc
docker compose logs -f           # View logs
docker compose down              # Stop services
```

### Integration Testing
Use Orthanc as the test client (more realistic than internal tests):
```bash
# C-ECHO
curl -X POST http://localhost:8080/modalities/SAMPLE_SCP/echo

# C-FIND
curl -X POST http://localhost:8080/modalities/SAMPLE_SCP/query \
  -d '{"Level":"Study","Query":{"PatientName":"*"}}'

# C-MOVE
curl -X POST http://localhost:8080/modalities/SAMPLE_SCP/query \
  -d '{"Level":"Study","Query":{"PatientName":"TEST*"}}' | \
  jq -r '.ID' | xargs -I {} curl -X POST \
  http://localhost:8080/queries/{}/retrieve -d '{"TargetAet":"ORTHANC"}'
```

## Version Release Process

When cutting a new version:

1. **Update CHANGELOG.md**
   - Add all new features under appropriate version section
   - Include `### Added`, `### Changed`, `### Fixed`, `### Testing` sections as needed
   - Use clear, user-facing descriptions
   - Mark breaking changes with `**BREAKING**:`
   - Set the release date (format: `YYYY-MM-DD`)

2. **Update version references**
   - Check if any code references the version number
   - Update README.md if needed

3. **Commit changes**
   ```bash
   git add CHANGELOG.md
   git commit -m "chore: prepare v0.x.0 release"
   ```

4. **Create and push tag**
   ```bash
   git tag v0.x.0
   git push origin main
   git push origin v0.x.0
   ```

5. **Verify tag on GitHub**
   - Check that tag appears in releases
   - Consider creating a GitHub release with CHANGELOG content

## Important Technical Details

### Priority Field in C-STORE
The Priority field (0000,0700) **must be non-zero** to be included in the command. Use `0x0002` (MEDIUM priority) as default. Zero priority causes some PACS systems (like Orthanc) to reject with "Command Parse Failed".

### Transfer Syntax Negotiation
- Always propose native transfer syntax first for best compatibility
- Fall back to Explicit VR (1.2.840.10008.1.2.1) and Implicit VR (1.2.840.10008.1.2)
- Check accepted transfer syntax in A-ASSOCIATE-AC before sending data

### PDU Fragmentation
- Large datasets must be fragmented into multiple P-DATA-TF PDUs
- Last fragment must have last-fragment flag set
- Each fragment includes presentation context ID

### C-MOVE Implementation
- C-MOVE responses are sent on the request association
- C-STORE sub-operations create **new associations** to destination AET
- Use shared `dimse.SendCStore()` for sub-operations
- Dynamic transfer syntax negotiation ensures compatibility

### Synthetic DICOM Data
Sample server can generate instances in memory:
```bash
./sample_server --synthetic  # Generates 3 CT instances
```
Useful for testing without real DICOM files (only 252 bytes per instance).

## Common Pitfalls to Avoid

1. **Don't expose internal ports unnecessarily** - Docker services on same network don't need port exposure
2. **Don't use zero for Priority field** - Will be omitted from command, causing failures
3. **Don't forget transfer syntax in presentation context** - Must match dataset encoding
4. **Don't mix VR encodings** - Commands are always Implicit VR, datasets typically Explicit VR
5. **Don't test with only our own client** - Use production PACS (Orthanc) to catch real-world issues

## Resources

- DICOM Standard: https://www.dicomstandard.org/
- Orthanc Book: https://orthanc.uclouvain.be/book/
- Transfer Syntax Registry: Part 6, Annex A of DICOM standard
