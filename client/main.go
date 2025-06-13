package client

import (
	"context"
	"fmt"
	"log"
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
	root, err := makeUniqueRoot(workspaceMode)
	if err != nil {
		log.Print("Error creating a unique root directory")
		return nil, err
	}
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

func makeUniqueRoot(workspaceMode bool) (string, error) {
	if workspaceMode {
		dir := filepath.Join("/app/workspace")
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return "", err
		}
		return dir, nil
	}
	// Single file mode (default)
	userFilePath := "/app/workspace"
	err := os.MkdirAll(userFilePath, 0755)
	if err != nil {
		return "", err
	}
	log.Print("Unique File Path made: ", userFilePath)
	return userFilePath, nil
}
