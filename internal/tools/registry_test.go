package tools

import (
	"context"
	"errors"
	"testing"
)

type dummyTool struct {
	name        string
	description string
	decl        FunctionDeclaration
	result      any
	err         error
}

func (d *dummyTool) Name() string                   { return d.name }
func (d *dummyTool) Description() string            { return d.description }
func (d *dummyTool) Declaration() FunctionDeclaration { return d.decl }
func (d *dummyTool) Execute(ctx context.Context, args map[string]any) (any, error) {
	return d.result, d.err
}

func TestRegistry_RegisterAndExecute(t *testing.T) {
	reg := NewRegistry()

	if reg.HasTools() {
		t.Fatalf("expected empty registry to have no tools")
	}

	tool1 := &dummyTool{
		name:        "mock_tool_1",
		description: "Mock tool for testing",
		decl: FunctionDeclaration{
			Name:        "mock_tool_1",
			Description: "Mock tool for testing",
		},
		result: map[string]string{"status": "ok"},
	}

	tool2 := &dummyTool{
		name:        "mock_tool_2",
		description: "Second mock tool",
		decl: FunctionDeclaration{
			Name:        "mock_tool_2",
			Description: "Second mock tool",
		},
		err: errors.New("simulated error"),
	}

	reg.RegisterAll(tool1, tool2)

	if !reg.HasTools() {
		t.Fatalf("expected registry to have tools")
	}

	names := reg.ListNames()
	if len(names) != 2 {
		t.Errorf("expected 2 tools, got %d", len(names))
	}

	decls := reg.GetDeclarations()
	if len(decls) != 2 {
		t.Errorf("expected 2 declarations, got %d", len(decls))
	}

	// Test successful execution
	res, err := reg.Execute(context.Background(), "mock_tool_1", map[string]any{"key": "val"})
	if err != nil {
		t.Fatalf("unexpected error executing tool1: %v", err)
	}
	resMap, ok := res.(map[string]string)
	if !ok || resMap["status"] != "ok" {
		t.Errorf("expected status 'ok', got %v", res)
	}

	// Test tool returning error
	_, err = reg.Execute(context.Background(), "mock_tool_2", nil)
	if err == nil {
		t.Fatalf("expected error from tool2, got nil")
	}

	// Test non-existent tool
	_, err = reg.Execute(context.Background(), "non_existent_tool", nil)
	if err == nil {
		t.Fatalf("expected error for non-existent tool, got nil")
	}
}
