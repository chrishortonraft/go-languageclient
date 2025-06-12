package client

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

// Client is the interface for interacting with any language client.
type Client interface {
	HoverAction(ctx context.Context, params any) (any, error)
	CompletionAction(ctx context.Context, params any) (any, error)
}

// config holds the configuration specific to the client.
type config struct {
	language      string
	root          string
	workspaceMode bool
}

// NewClient creates a new LSP client based on the specified language.
func NewClient(language string, workspaceMode bool) (Client, error) {
	id := uuid.New()
	root := makeUniqueRoot(id, workspaceMode)
	cfg := config{
		language:      language,
		root:          root,
		workspaceMode: workspaceMode,
	}

	switch language {
	case "python":
		return NewPythonClient(cfg)
	default:
		return nil, fmt.Errorf("unsupported language: %s", language)
	}
}

func makeUniqueRoot(id uuid.UUID, workspaceMode bool) string {
	if workspaceMode {
		dir := filepath.Join(id.String(), "workspace")
		os.MkdirAll(dir, 0755)
		filePath := filepath.Join(dir, "main.py")
		os.WriteFile(filePath, []byte("print('Hello, World!')"), 0644)
		return dir
	}
	// Single file mode (default)
	userFilePath := filepath.Join(id.String(), "main.py")
	os.WriteFile(userFilePath, []byte(""), 0644)
	return userFilePath
}
