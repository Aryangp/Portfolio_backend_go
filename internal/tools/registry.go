package tools

import (
	"context"
	"fmt"
	"sync"
)

// Registry manages the collection of dynamic tools available to the chatbot
type Registry struct {
	mu    sync.RWMutex
	tools map[string]Tool
}

// NewRegistry creates a new, empty dynamic tool registry
func NewRegistry() *Registry {
	return &Registry{
		tools: make(map[string]Tool),
	}
}

// Register registers a tool into the registry
func (r *Registry) Register(tool Tool) {
	if tool == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tools[tool.Name()] = tool
}

// RegisterAll registers multiple tools into the registry
func (r *Registry) RegisterAll(tools ...Tool) {
	for _, tool := range tools {
		r.Register(tool)
	}
}

// Get retrieves a tool by name
func (r *Registry) Get(name string) (Tool, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	tool, ok := r.tools[name]
	return tool, ok
}

// HasTools returns true if at least one tool is registered
func (r *Registry) HasTools() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.tools) > 0
}

// ListNames returns the list of registered tool names
func (r *Registry) ListNames() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.tools))
	for name := range r.tools {
		names = append(names, name)
	}
	return names
}

// GetDeclarations returns all function declarations from registered tools formatted for Gemini
func (r *Registry) GetDeclarations() []FunctionDeclaration {
	r.mu.RLock()
	defer r.mu.RUnlock()
	declarations := make([]FunctionDeclaration, 0, len(r.tools))
	for _, tool := range r.tools {
		declarations = append(declarations, tool.Declaration())
	}
	return declarations
}

// Execute executes the requested tool by name with arguments
func (r *Registry) Execute(ctx context.Context, name string, args map[string]any) (any, error) {
	r.mu.RLock()
	tool, exists := r.tools[name]
	r.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("tool '%s' not found in registry", name)
	}

	return tool.Execute(ctx, args)
}
