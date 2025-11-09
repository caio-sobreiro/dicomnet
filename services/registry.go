package services

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/caio-sobreiro/dicomnet/interfaces"
	"github.com/caio-sobreiro/dicomnet/types"
)

// Registry manages DICOM service handlers and routes incoming DIMSE messages.
//
// The registry acts as a dispatcher, routing DIMSE messages to the appropriate
// service handler based on the command field. It supports both single-response
// and streaming (multi-response) operations.
//
// Example usage:
//
//	registry := services.NewRegistry()
//	registry.RegisterHandler(dimse.CEchoRQ, echoService)
//	registry.RegisterHandler(dimse.CFindRQ, findService)
//
//	// In your server handler:
//	response, data, err := registry.HandleDIMSE(ctx, msg, data)
type Registry struct {
	handlers map[uint16]interfaces.ServiceHandler
}

// NewRegistry creates a new service registry.
//
// Returns an empty registry. Use RegisterHandler to add service handlers.
func NewRegistry() *Registry {
	return &Registry{
		handlers: make(map[uint16]interfaces.ServiceHandler),
	}
}

// RegisterHandler registers a service handler for a specific DIMSE command.
//
// The handler will be invoked when a message with the specified command field
// is received. Only one handler can be registered per command field; calling
// RegisterHandler again with the same command will replace the previous handler.
//
// Parameters:
//   - commandField: The DIMSE command field (e.g., dimse.CEchoRQ, dimse.CFindRQ)
//   - handler: The service handler that will process messages for this command
//
// Example:
//
//	registry.RegisterHandler(dimse.CEchoRQ, NewEchoService())
//	registry.RegisterHandler(dimse.CFindRQ, myFindService)
func (r *Registry) RegisterHandler(commandField uint16, handler interfaces.ServiceHandler) {
	r.handlers[commandField] = handler
}

// UnregisterHandler removes a service handler for a specific DIMSE command.
//
// After unregistering, messages with this command field will result in
// an "unsupported command" error.
func (r *Registry) UnregisterHandler(commandField uint16) {
	delete(r.handlers, commandField)
}

// HandleDIMSE routes DIMSE messages to the appropriate service handler.
//
// This method provides the single-response interface for DIMSE operations.
// For operations that support streaming (like C-FIND), use HandleDIMSEStreaming instead.
//
// If no handler is registered for the message's command field, returns an error.
//
// Parameters:
//   - ctx: Context for cancellation and request tracking
//   - msg: The incoming DIMSE command message
//   - data: The optional dataset associated with the message
//
// Returns:
//   - Response DIMSE message
//   - Response dataset (if any)
//   - Error if handling fails or no handler is registered
func (r *Registry) HandleDIMSE(ctx context.Context, msg *types.Message, data []byte) (*types.Message, []byte, error) {
	slog.DebugContext(ctx, "Routing DIMSE message",
		"command_field", fmt.Sprintf("0x%04x", msg.CommandField),
		"message_id", msg.MessageID)

	handler, ok := r.handlers[msg.CommandField]
	if !ok {
		slog.WarnContext(ctx, "No handler registered for DIMSE command",
			"command_field", fmt.Sprintf("0x%04x", msg.CommandField))
		return nil, nil, fmt.Errorf("unsupported DIMSE command: 0x%04x", msg.CommandField)
	}

	return handler.HandleDIMSE(ctx, msg, data)
}

// HandleDIMSEStreaming routes streaming DIMSE messages to appropriate service handlers.
//
// This method is preferred for operations that can return multiple responses,
// such as C-FIND which may return many matching results.
//
// If the registered handler implements interfaces.StreamingServiceHandler, it will
// use the streaming interface. Otherwise, it falls back to HandleDIMSE and sends
// a single response.
//
// Parameters:
//   - ctx: Context for cancellation and request tracking
//   - msg: The incoming DIMSE command message
//   - data: The optional dataset associated with the message
//   - responder: Interface for sending multiple responses
//
// Returns:
//   - Error if handling fails or no handler is registered
func (r *Registry) HandleDIMSEStreaming(ctx context.Context, msg *types.Message, data []byte, responder interfaces.ResponseSender) error {
	slog.DebugContext(ctx, "Routing streaming DIMSE message",
		"command_field", fmt.Sprintf("0x%04x", msg.CommandField),
		"message_id", msg.MessageID)

	handler, ok := r.handlers[msg.CommandField]
	if !ok {
		slog.WarnContext(ctx, "No handler registered for DIMSE command",
			"command_field", fmt.Sprintf("0x%04x", msg.CommandField))
		return fmt.Errorf("unsupported DIMSE command: 0x%04x", msg.CommandField)
	}

	// Check if handler supports streaming
	if streamingHandler, ok := handler.(interfaces.StreamingServiceHandler); ok {
		return streamingHandler.HandleDIMSEStreaming(ctx, msg, data, responder)
	}

	// Fallback to single-response handler
	responseMsg, responseData, err := handler.HandleDIMSE(ctx, msg, data)
	if err != nil {
		return err
	}
	return responder.SendResponse(responseMsg, responseData)
}

// HasHandler returns true if a handler is registered for the given command field.
func (r *Registry) HasHandler(commandField uint16) bool {
	_, ok := r.handlers[commandField]
	return ok
}

// RegisteredCommands returns a list of all command fields that have handlers registered.
func (r *Registry) RegisteredCommands() []uint16 {
	commands := make([]uint16, 0, len(r.handlers))
	for cmd := range r.handlers {
		commands = append(commands, cmd)
	}
	return commands
}

// CreateErrorResponse creates a standard DIMSE error response message.
//
// This is a utility function for creating error responses when handling fails.
// The response will have the appropriate response command field (original | 0x8000),
// the message ID being responded to, and the specified status code.
//
// Parameters:
//   - req: The original request message
//   - status: The status code for the error response
//
// Returns:
//   - Error response message
func CreateErrorResponse(req *types.Message, status uint16) *types.Message {
	return &types.Message{
		CommandField:              req.CommandField | 0x8000, // Set response bit
		MessageIDBeingRespondedTo: req.MessageID,
		AffectedSOPClassUID:       req.AffectedSOPClassUID,
		CommandDataSetType:        0x0101, // No dataset
		Status:                    status,
	}
}
