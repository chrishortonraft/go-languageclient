package client

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
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
	root := makeUniqueRoot(workspaceMode)
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

func makeUniqueRoot(workspaceMode bool) string {
	if workspaceMode {
		dir := filepath.Join("/app/workspace")
		os.MkdirAll(dir, 0755)
		return dir
	}
	// Single file mode (default)
	userFilePath := filepath.Join("/app/workspace/main.py")
	os.WriteFile(userFilePath, []byte(""), 0644)
	return userFilePath
}
