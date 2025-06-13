package main

import (
	"context"
	"encoding/json"
	"fmt"
	"golang-client/client"
	"log"
	"net/http"
)

// The GoProxy struct
type GoProxy struct {
	client client.Client
}

// InitializeProxy initializes the proxy by creating a client
func (g *GoProxy) InitializeProxy(language string, workspaceMode bool) error {
	var err error
	g.client, err = client.NewClient(language, workspaceMode)
	if err != nil {
		return err
	}
	return nil
}

// ProcessRequest processes HTTP requests, delegating them to the appropriate method in the client
func (g *GoProxy) ProcessRequest(ctx context.Context, method string, params any) (any, error) {
	switch method {
	case "hover":
		return g.client.HoverAction(ctx, params)
	case "completion":
		return g.client.CompletionAction(ctx, params)
	default:
		return nil, fmt.Errorf("unknown method: %s", method)
	}
}

// HTTP request handler
func handleRequest(proxy *GoProxy, w http.ResponseWriter, r *http.Request) {
	var req struct {
		Method string `json:"method"`
		Params any    `json:"params"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Failed to parse request: %v", err), http.StatusBadRequest)
		return
	}

	fmt.Println("Recieved Request:", req.Method)

	ctx := context.Background()
	result, err := proxy.ProcessRequest(ctx, req.Method, req.Params)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to process request: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]any{
		"result": result,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, fmt.Sprintf("Failed to send response: %v", err), http.StatusInternalServerError)
	}
}

func main() {
	proxy := &GoProxy{}
	err := proxy.InitializeProxy("python", false)
	if err != nil {
		log.Fatalf("Failed to initialize proxy: %v", err)
	}

	http.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
		handleRequest(proxy, w, r)
	})

	port := ":8080"
	fmt.Printf("Server running on port %s...\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
