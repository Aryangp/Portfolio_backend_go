package tools

import "context"

// ParameterType defines the JSON Schema data type in Gemini function declarations
type ParameterType string

const (
	TypeString  ParameterType = "STRING"
	TypeNumber  ParameterType = "NUMBER"
	TypeInteger ParameterType = "INTEGER"
	TypeBoolean ParameterType = "BOOLEAN"
	TypeArray   ParameterType = "ARRAY"
	TypeObject  ParameterType = "OBJECT"
)

// Schema defines the parameter structure for Gemini tool declarations
type Schema struct {
	Type        ParameterType      `json:"type"`
	Description string             `json:"description,omitempty"`
	Properties  map[string]*Schema `json:"properties,omitempty"`
	Required    []string           `json:"required,omitempty"`
	Items       *Schema            `json:"items,omitempty"`
	Enum        []string           `json:"enum,omitempty"`
}

// FunctionDeclaration defines the metadata Gemini needs to invoke a tool
type FunctionDeclaration struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Parameters  *Schema `json:"parameters,omitempty"`
}

// Tool is the interface that all dynamic tools (GitHub, LinkedIn, etc.) must implement
type Tool interface {
	Name() string
	Description() string
	Declaration() FunctionDeclaration
	Execute(ctx context.Context, args map[string]any) (any, error)
}
