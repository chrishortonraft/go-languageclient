package client

import (
	"io"
	"net"
	"testing"
	"time"
)

// Mock a net.Conn for testing (instead of using a real pipe or external process)
type mockConn struct {
	Server *End
	Client *End
}

func NewConn() *mockConn {
	// A connection consists of two pipes:
	// Client      |      Server
	//   writes   ===>  reads
	//    reads  <===   writes

	serverRead, clientWrite := io.Pipe()
	clientRead, serverWrite := io.Pipe()

	return &mockConn{
		Server: &End{
			Reader: serverRead,
			Writer: serverWrite,
		},
		Client: &End{
			Reader: clientRead,
			Writer: clientWrite,
		},
	}
}

func (c *mockConn) Close() error {
	if err := c.Server.Close(); err != nil {
		return err
	}
	if err := c.Client.Close(); err != nil {
		return err
	}
	return nil
}

type Addr struct {
	NetworkString string
	AddrString    string
}

func (a Addr) Network() string {
	return a.NetworkString
}

func (a Addr) String() string {
	return a.AddrString
}

type End struct {
	Reader *io.PipeReader
	Writer *io.PipeWriter
}

func (c End) Close() error {
	if err := c.Writer.Close(); err != nil {
		return err
	}
	if err := c.Reader.Close(); err != nil {
		return err
	}
	return nil
}

func (e End) Read(data []byte) (n int, err error)  { return e.Reader.Read(data) }
func (e End) Write(data []byte) (n int, err error) { return e.Writer.Write(data) }

func (e End) LocalAddr() net.Addr {
	return Addr{
		NetworkString: "tcp",
		AddrString:    "127.0.0.1",
	}
}

func (e End) RemoteAddr() net.Addr {
	return Addr{
		NetworkString: "tcp",
		AddrString:    "127.0.0.1",
	}
}

func (e End) SetDeadline(t time.Time) error      { return nil }
func (e End) SetReadDeadline(t time.Time) error  { return nil }
func (e End) SetWriteDeadline(t time.Time) error { return nil }

// Test the ReadResponse function
func TestReadResponse(t *testing.T) {
	// Create a mock connection with a predefined response
	conn := NewConn()

	// We need to write the content-length and JSON in separate goroutines to avoid deadlock
	go func() {
		conn.Client.Write([]byte("Content-Length: 127\r\n\r\n"))
		conn.Client.Write([]byte(`{"jsonrpc": "2.0", "method": "window/logMessage", "params": {"type": 3, "message": "Pyright language server 1.1.401 starting"}}`))
	}()

	client := &PythonClient{
		conn: conn.Server, // PythonClient should use the Server side for reading
	}

	// Call the ReadResponse function in the main goroutine
	response, err := client.ReadResponse()
	if err != nil {
		t.Fatalf("Expected no error, but got: %v", err)
	}

	// Check if the response contains the expected values
	if response == nil {
		t.Fatal("Expected response, but got nil")
	}

	// Check if the response has the correct method
	if response["method"] != "window/logMessage" {
		t.Fatalf("Expected 'window/logMessage' method, but got: %v", response["method"])
	}
}
func TestNewPyrightClient(t *testing.T) {
	cfg := config{
		language:      "python",
		root:          "main.py",
		workspaceMode: false,
	}

	_, err := NewPythonClient(cfg)
	if err != nil {
		t.Fatalf("Initialization failed")
	}
}

// Test handling multiple responses
// Test handling multiple responses with channel synchronization
func TestReadMultipleResponses(t *testing.T) {
	t.Skip("Skip testing multiple until needed")
	// Create a mock connection with multiple responses
	conn := NewConn()

	// Channel for signaling the end of writing
	done := make(chan struct{})

	// Goroutine for writing multiple responses
	go func() {
		// Write first response
		conn.Client.Write([]byte("Content-Length: 127\r\n\r\n"))
		conn.Client.Write([]byte(`{"jsonrpc": "2.0", "method": "window/logMessage", "params": {"type": 3, "message": "Pyright language server 1.1.401 starting"}}`))

		// Write second response
		conn.Client.Write([]byte("Content-Length: 160\r\n\r\n"))
		conn.Client.Write([]byte(`{"jsonrpc": "2.0", "method": "window/logMessage", "params": {"type": 3, "message": "Server root directory: file:///opt/homebrew/lib/node_modules/pyright/dist"}}`))

		// Signal that writing is done
		close(done) // Notify the reader that writing is done
	}()

	client := &PythonClient{
		conn: conn.Server, // PythonClient should use the Server side for reading
	}

	// Wait for the writer goroutine to complete
	<-done // Block until writing is complete

	// Call the ReadResponses function to read both responses
	responses, err := client.ReadResponses()
	if err != nil {
		t.Fatalf("Expected no error, but got: %v", err)
	}

	// There should be two responses
	if len(responses) != 2 {
		t.Fatalf("Expected 2 responses, but got: %d", len(responses))
	}

	// Check if the responses are correctly parsed
	if responses[0]["method"] != "window/logMessage" {
		t.Fatalf("Expected 'window/logMessage' method, but got: %v", responses[0]["method"])
	}
	if responses[1]["method"] != "window/logMessage" {
		t.Fatalf("Expected 'window/logMessage' method, but got: %v", responses[1]["method"])
	}
}
