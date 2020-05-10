package openapi

// Contact ...
type Contact struct {
	Name  string `yaml:"name"`
	URL   string `yaml:"url"`
	Email string `yaml:"email"`
}

// License ...
type License struct {
	Name string `yaml:"name"`
	URL  string `yaml:"url"`
}

// Info ...
type Info struct {
	Title          string   `yaml:"title"`
	Description    string   `yaml:"description"`
	Version        string   `yaml:"version"`
	TermsOfService string   `yaml:"termsOfService"`
	Contact        *Contact `yaml:"contact"`
	License        *License `yaml:"license"`
}

// Encoding ...
type Encoding struct {
	AllowReserved bool               `yaml:"allowReserved"`
	ContentType   string             `yaml:"contentType"`
	Explode       bool               `yaml:"explode"`
	Headers       map[string]*Header `yaml:"header"`
	Style         string             `yaml:"string"`
}

// MediaType ...
type MediaType struct {
	Ref      string               `yaml:"$ref"`
	Examples map[string]*Example  `yaml:"examples"`
	Encoding map[string]*Encoding `yaml:"encoding"`
	Schema   *Schema              `yaml:"schema"`
}

// Response ...
type Response struct {
	Ref         string                `yaml:"$ref"`
	Content     map[string]*MediaType `yaml:"content"`
	Description string                `yaml:"description"`
}

// Operation ...
type Operation struct {
	Tags         []string               `yaml:"tags"`
	Summary      string                 `yaml:"summary"`
	Description  string                 `yaml:"description"`
	ExternalDocs *ExternalDocumentation `yaml:"externalDocs"`
	OperationID  string                 `yaml:"operationId"`
	Parameters   []*Parameter           `yaml:"parameters"`
	RequestBody  *RequestBody           `yaml:"requestBody"`
	Responses    map[string]*Response   `yaml:"responses"`
	Callbacks    map[string]*Callback   `yaml:"callbacks"`
	Deprecated   bool                   `yaml:"deprecated"`
	Security     *SecurityRequirement   `yaml:"security"`
	Servers      []*Server              `yaml:"servers"`
}

// PathItem ...
type PathItem struct {
	Ref         string       `yaml:"$ref"`
	Summary     string       `yaml:"summary"`
	Description string       `yaml:"description"`
	Get         *Operation   `yaml:"get"`
	Put         *Operation   `yaml:"put"`
	Post        *Operation   `yaml:"post"`
	Delete      *Operation   `yaml:"delete"`
	Options     *Operation   `yaml:"options"`
	Head        *Operation   `yaml:"head"`
	Patch       *Operation   `yaml:"patch"`
	Trace       *Operation   `yaml:"trace"`
	Servers     []*Server    `yaml:"servers"`
	Parameters  []*Parameter `yaml:"parameters"`
}

// ServerVariableObject ...
type ServerVariableObject struct {
	Default     string   `yaml:"default"`
	Description string   `yaml:"description"`
	Enum        []string `yaml:"enum"`
}

// Server ...
type Server struct {
	URL         string                           `yaml:"url"`
	Description string                           `yaml:"description"`
	Variables   map[string]*ServerVariableObject `yaml:"variables"`
}

// Tag ...
type Tag struct {
	Name         string                 `yaml:"name"`
	Description  string                 `yaml:"description"`
	ExternalDocs *ExternalDocumentation `yaml:"externalDocs"`
}

// ExternalDocumentation ...
type ExternalDocumentation struct {
	Description string `yaml:"description"`
	URL         string `yaml:"url"`
}

// SecurityRequirement ...
type SecurityRequirement = map[string][]string

// OpenAPI ...
type OpenAPI struct {
	Version      string                 `yaml:"openapi"`
	Info         *Info                  `yaml:"info"`
	Paths        map[string]*PathItem   `yaml:"paths"`
	Servers      []*Server              `yaml:"servers"`
	Components   *Components            `yaml:"components"`
	Security     []SecurityRequirement  `yaml:"security"`
	Tags         []*Tag                 `yaml:"tags"`
	ExternalDocs *ExternalDocumentation `yaml:"externalDocs"`
}

// Discriminator ...
type Discriminator struct {
}

// XML ...
type XML struct {
}

// Schema ...
type Schema struct {
	Ref              string                 `yaml:"$ref"`
	Properties       map[string]*Schema     `yaml:"properties"`
	Nullable         bool                   `yaml:"nullable"`
	Discriminator    *Discriminator         `yaml:"discriminator"`
	ReadOnly         bool                   `yaml:"readOnly"`
	WriteOnly        bool                   `yaml:"writeOnly"`
	XML              XML                    `yaml:"xml"`
	ExternalDocs     *ExternalDocumentation `yaml:"externalDocs"`
	Example          string                 `yaml:"example"`
	Deprecated       bool                   `yaml:"deprecated"`
	Type             string                 `yaml:"type"`
	Format           string                 `yaml:"format"`
	Title            string                 `yaml:"title"`
	MultipleOf       int                    `yaml:"multipleOf"`
	Maximum          int                    `yaml:"maximum"`
	ExclusiveMaximum bool                   `yaml:"exclusiveMaximum"`
	Minimum          int                    `yaml:"minimum"`
	ExclusiveMinimum bool                   `yaml:"exclusiveMinimum"`
	MaxLength        uint                   `yaml:"maxLength"`
	MinLength        uint                   `yaml:"minLength"`
	Pattern          string                 `yaml:"pattern"`
	MaxItems         uint                   `yaml:"maxItems"`
	MinItems         uint                   `yaml:"minItems"`
	UniqueItems      bool                   `yaml:"uniqueItmes"`
	MaxProperties    uint                   `yaml:"maxProperties"`
	MinProperties    uint                   `yaml:"minProperties"`
	Required         []string               `yaml:"required"`
	Enum             []string               `yaml:"enum"`
	Items            *Schema                `yaml:"items"`
	AllOf            []*Schema              `yaml:"allOf"`
	OneOf            []*Schema              `yaml:"oneOf"`
	AnyOf            []*Schema              `yaml:"anyOf"`
	Not              []*Schema              `yaml:"not"`
}

// Parameter ...
type Parameter struct {
	Ref             string              `yaml:"$ref"`
	Name            string              `yaml:"name"`
	In              string              `yaml:"in"`
	Description     string              `yaml:"description"`
	Required        bool                `yaml:"required"`
	Deprecated      bool                `yaml:"deprecated"`
	AllowEmptyValue bool                `yaml:"allowEmptyValue"`
	Style           string              `yaml:"style"`
	Explode         bool                `yaml:"explode"`
	AllowReserved   bool                `yaml:"allowReserved"`
	Schema          *Schema             `yaml:"schema"`
	Example         string              `yaml:"example"`
	Examples        map[string]*Example `yaml:"examples"`
}

// Example ...
type Example struct {
	Ref string `yaml:"$ref"`
}

// RequestBody ...
type RequestBody struct {
	Ref         string                `yaml:"$ref"`
	Description string                `yaml:"description"`
	Content     map[string]*MediaType `yaml:"content"`
	Required    bool                  `yaml:"required"`
}

// Header ...
type Header struct {
	Ref string `yaml:"$ref"`
}

// SecurityScheme ...
type SecurityScheme struct {
	Ref string `yaml:"$ref"`
}

// Link ...
type Link struct {
	Ref string `yaml:"$ref"`
}

// Callback ...
type Callback struct {
	Ref string `yaml:"$ref"`
}

// Components ...
type Components struct {
	Schemas         map[string]*Schema         `yaml:"schemas"`
	Responses       map[string]*Response       `yaml:"responses"`
	Parameters      map[string]*Parameter      `yaml:"parameters"`
	Examples        map[string]*Example        `yaml:"examples"`
	RequestBodies   map[string]*RequestBody    `yaml:"requestBodies"`
	Headers         map[string]*Header         `yaml:"headers"`
	SecuritySchemes map[string]*SecurityScheme `yaml:"securitySchemes"`
	Links           map[string]*Link           `yaml:"links"`
	Callbacks       map[string]*Callback       `yaml:"callback"`
}
