package client

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os/exec"
	"path/filepath"

	"github.com/google/uuid"
)

// PythonClient implements the Client interface for the Python language (via Pyright).
type PythonClient struct {
	config    config
	sessionId uuid.UUID
	conn      net.Conn // Pipe connection to Pyright
	cmd       *exec.Cmd
}

// NewPythonClient creates a new PythonClient and connects it to the Pyright language server.
func NewPythonClient(config config) (*PythonClient, error) {
	// Create a pipe for communication with Pyright
	connA, connB := net.Pipe()

	// Start the Pyright process
	cmd := exec.Command("pyright-langserver", "--stdio")
	cmd.Stdin = connB
	cmd.Stdout = connB

	// Start the command
	err := cmd.Start()
	if err != nil {
		return nil, fmt.Errorf("failed to start Pyright: %v", err)
	}

	// Check if the process is running
	if cmd.Process == nil {
		return nil, fmt.Errorf("pyright process is not running")
	}

	client := &PythonClient{
		config:    config,
		sessionId: uuid.New(),
		conn:      connA, // Go side of the pipe
		cmd:       cmd,
	}

	// Read the startup messages first - they should be logged in ReadMessage
	_, err = client.ReadResponse()
	if err != nil {
		return nil, fmt.Errorf("failed to read startup message: %v", err)
	}

	_, err = client.ReadResponse()
	if err != nil {
		return nil, fmt.Errorf("failed to read startup message: %v", err)
	}

	log.Print("Root URI: ", client.config.root)
	// Send an init message to pyright to let it know our settings, etc
	err = client.InitializePyright()
	if err != nil {
		return nil, err
	}

	return client, nil
}

func (p *PythonClient) InitializePyright() error {
	workspaceRoot := filepath.Join("file://", p.config.root)
	log.Print("Workspace Root: ", workspaceRoot)

	initRequest := map[string]any{
		"jsonrpc": "2.0",
		"id":      p.sessionId, // You can generate a unique ID if needed
		"method":  "initialize",
		"params": map[string]any{
			// Debate: this links the pyright to this process (typically an editor)
			// so when it closes, so does pyright. Could lead to some potentially weird bugs.
			// Going to leave it nil for now so it doesn't auto close...
			"processId": nil,
			"capabilities": map[string]any{
				"textDocument": map[string]any{
					"completion": map[string]any{
						"dynamicRegistration": true,
					},
					"hover": map[string]any{
						"dynamicRegistration": true,
					},
				},
			},
			"workspaceFolders": []map[string]any{
				{
					"uri":  workspaceRoot,
					"name": "Workspace Folder",
				},
			},
		},
	}

	// Marshal the request into JSON format
	err := p.SendMessage(initRequest)
	if err != nil {
		return fmt.Errorf("failed to send initialize request: %v", err)
	}

	// Get server initialization response with its capabilites
	starting_instance, err := p.ReadResponse()
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	fmt.Printf("Log Message: %#v\n", starting_instance)

	capabilites, err := p.ReadResponse()
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	fmt.Printf("LS Capabilities: %#v\n", capabilites)

	// After initialization syn/ack - the client sends an initialized _notification_ to the server
	initNotification := map[string]any{
		"method": "initialized",
		"params": map[string]any{}, // No parameters for the initialized notification
	}

	err = p.SendMessage(initNotification)
	if err != nil {
		return fmt.Errorf("failed to marshal initialized notification: %v", err)
	}

	// Response holds the capabilities of the server. Would be good to process elsewhere eventually
	return nil
}

// SendMessage sends a JSON-RPC message with the correct Content-Length header
func (p *PythonClient) SendMessage(request map[string]any) error {
	// Encode the request into JSON format
	requestBytes, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to encode request: %v", err)
	}

	// Prepare the content length header
	contentLength := fmt.Sprintf("Content-Length: %d\r\n\r\n", len(requestBytes))

	// Combine the header and the JSON payload into one single message
	message := contentLength + string(requestBytes)

	// Write the entire message in one go (header + content)
	_, err = p.conn.Write([]byte(message))
	if err != nil {
		return fmt.Errorf("failed to write message: %v", err)
	}

	return nil
}

// ReadResponse reads the response from Pyright and handles multiple messages.
func (p *PythonClient) ReadResponse() (map[string]any, error) {
	// Read the content length header first
	var contentLengthHeader string
	_, err := fmt.Fscanf(p.conn, "Content-Length: %s\r\n\r\n", &contentLengthHeader)
	if err != nil {
		return nil, fmt.Errorf("failed to read content length header: %v", err)
	}

	// Parse the content length (it should be a number)
	var contentLength int
	_, err = fmt.Sscanf(contentLengthHeader, "%d", &contentLength)
	if err != nil {
		return nil, fmt.Errorf("failed to parse content length: %v", err)
	}

	// Read the actual JSON response based on content length
	responseBytes := make([]byte, contentLength)
	_, err = p.conn.Read(responseBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	// Decode the response JSON
	var response map[string]any
	err = json.Unmarshal(responseBytes, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	// Make sure it's not a server notification
	// Maybe set logging method here for verbose or not
	if method, ok := response["method"].(string); ok {
		if method == "window/logMessage" {
			if params, ok := response["params"].(map[string]any); ok {
				if message, ok := params["message"].(string); ok {
					fmt.Printf("Log Message: %s\n", message)
				}
			}
		} else {
			return response, fmt.Errorf("unexpected notification method: %s", method)
		}
	}

	// Return the response
	return response, nil
}

// ReadResponses reads the response from Pyright and handles multiple messages.
func (p *PythonClient) ReadResponses() ([]map[string]any, error) {
	var responses []map[string]any

	reader := bufio.NewReader(p.conn) // Use bufio.Reader to handle the connection

	for {
		// Check if data is available to read before attempting to read it
		line, err := reader.Peek(1) // Peek the first byte
		if err != nil {
			if err == io.EOF {
				break // Exit loop on EOF
			}
			return nil, fmt.Errorf("failed to peek data: %v", err)
		}

		// If there's no data to read, continue
		if len(line) == 0 {
			break // Exit the loop if there's no more data
		}

		// Read the content length header first
		var contentLengthHeader string
		_, err = fmt.Fscanf(reader, "Content-Length: %s\r\n\r\n", &contentLengthHeader)
		if err != nil {
			// If there's no content length header, exit the loop gracefully
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("failed to read content length header: %v", err)
		}

		// Parse the content length (it should be a number)
		var contentLength int
		_, err = fmt.Sscanf(contentLengthHeader, "%d", &contentLength)
		if err != nil {
			return nil, fmt.Errorf("failed to parse content length: %v", err)
		}

		// Read the actual JSON response based on content length
		responseBytes := make([]byte, contentLength)
		_, err = reader.Read(responseBytes)
		if err != nil {
			if err == io.EOF {
				break // Exit if we hit EOF while reading the response
			}
			return nil, fmt.Errorf("failed to read response: %v", err)
		}

		// Decode the response JSON
		var response map[string]any
		err = json.Unmarshal(responseBytes, &response)
		if err != nil {
			return nil, fmt.Errorf("failed to decode response: %v", err)
		}

		responses = append(responses, response)
	}

	// Return all the responses
	return responses, nil
}

// HoverAction sends a Hover action request to Pyright.
func (p *PythonClient) HoverAction(ctx context.Context, params any) (any, error) {
	// Create a Hover request
	log.Print(params)
	request := map[string]any{
		"jsonrpc": "2.0",
		"id":      p.sessionId.String(),
		"method":  "textDocument/hover",
		"params":  params,
	}

	// Send the message
	err := p.SendMessage(request)
	if err != nil {
		return nil, fmt.Errorf("failed to send hover request: %v", err)
	}

	// Read the response from Pyright
	// Expect a response back (not a log message, so keep reading until we get a log message)
	for {
		response, err := p.ReadResponse()
		if err != nil {
			return nil, fmt.Errorf("failed to receive response to hover request: %v", err)
		}
		if response["id"] == p.sessionId.String() {
			return response, nil
		}
	}
}

// CompletionAction sends a Completion action request to Pyright.
func (p *PythonClient) CompletionAction(ctx context.Context, params any) (any, error) {
	// Create a Completion request
	request := map[string]any{
		"jsonrpc": "2.0",
		"id":      p.sessionId.String(),
		"method":  "textDocument/completion",
		"params":  params,
	}

	// Send the message
	err := p.SendMessage(request)
	if err != nil {
		return nil, fmt.Errorf("failed to send completion request: %v", err)
	}

	for {
		response, err := p.ReadResponse()
		if err != nil {
			return nil, fmt.Errorf("failed to receive response to hover request: %v", err)
		}
		if response["id"] == p.sessionId.String() {
			return response, nil
		}
	}
}
