package openapi

// Contact ...
type Contact struct {
	Name  string `yaml:"name,omitempty"`
	URL   string `yaml:"url,omitempty"`
	Email string `yaml:"email,omitempty"`
}

// License ...
type License struct {
	Name string `yaml:"name,omitempty"`
	URL  string `yaml:"url,omitempty"`
}

// Info ...
type Info struct {
	Title          string   `yaml:"title,omitempty"`
	Description    string   `yaml:"description,omitempty"`
	Version        string   `yaml:"version,omitempty"`
	TermsOfService string   `yaml:"termsOfService,omitempty"`
	Contact        *Contact `yaml:"contact,omitempty"`
	License        *License `yaml:"license,omitempty"`
}

// Encoding ...
type Encoding struct {
	AllowReserved bool               `yaml:"allowReserved,omitempty"`
	ContentType   string             `yaml:"contentType,omitempty"`
	Explode       bool               `yaml:"explode,omitempty"`
	Headers       map[string]*Header `yaml:"header,omitempty"`
	Style         string             `yaml:"string,omitempty"`
}

// MediaType ...
type MediaType struct {
	Ref      string               `yaml:"$ref,omitempty"`
	Examples map[string]*Example  `yaml:"examples,omitempty"`
	Encoding map[string]*Encoding `yaml:"encoding,omitempty"`
	Schema   *Schema              `yaml:"schema,omitempty"`
}

// Response ...
type Response struct {
	Ref         string                `yaml:"$ref,omitempty"`
	Content     map[string]*MediaType `yaml:"content,omitempty"`
	Description string                `yaml:"description,omitempty"`
}

// Operation ...
type Operation struct {
	Tags         []string               `yaml:"tags,omitempty"`
	Summary      string                 `yaml:"summary,omitempty"`
	Description  string                 `yaml:"description,omitempty"`
	ExternalDocs *ExternalDocumentation `yaml:"externalDocs,omitempty"`
	OperationID  string                 `yaml:"operationId,omitempty"`
	Parameters   []*Parameter           `yaml:"parameters,omitempty"`
	RequestBody  *RequestBody           `yaml:"requestBody,omitempty"`
	Responses    map[string]*Response   `yaml:"responses,omitempty"`
	Callbacks    map[string]*Callback   `yaml:"callbacks,omitempty"`
	Deprecated   bool                   `yaml:"deprecated,omitempty"`
	Security     *SecurityRequirement   `yaml:"security,omitempty"`
	Servers      []*Server              `yaml:"servers,omitempty"`
}

// PathItem ...
type PathItem struct {
	Ref         string       `yaml:"$ref,omitempty"`
	Summary     string       `yaml:"summary,omitempty"`
	Description string       `yaml:"description,omitempty"`
	Get         *Operation   `yaml:"get,omitempty"`
	Put         *Operation   `yaml:"put,omitempty"`
	Post        *Operation   `yaml:"post,omitempty"`
	Delete      *Operation   `yaml:"delete,omitempty"`
	Options     *Operation   `yaml:"options,omitempty"`
	Head        *Operation   `yaml:"head,omitempty"`
	Patch       *Operation   `yaml:"patch,omitempty"`
	Trace       *Operation   `yaml:"trace,omitempty"`
	Servers     []*Server    `yaml:"servers,omitempty"`
	Parameters  []*Parameter `yaml:"parameters,omitempty"`
}

// ServerVariableObject ...
type ServerVariableObject struct {
	Default     string   `yaml:"default,omitempty"`
	Description string   `yaml:"description,omitempty"`
	Enum        []string `yaml:"enum,omitempty"`
}

// Server ...
type Server struct {
	URL         string                           `yaml:"url,omitempty"`
	Description string                           `yaml:"description,omitempty"`
	Variables   map[string]*ServerVariableObject `yaml:"variables,omitempty"`
}

// Tag ...
type Tag struct {
	Name         string                 `yaml:"name,omitempty"`
	Description  string                 `yaml:"description,omitempty"`
	ExternalDocs *ExternalDocumentation `yaml:"externalDocs,omitempty"`
}

// ExternalDocumentation ...
type ExternalDocumentation struct {
	Description string `yaml:"description,omitempty"`
	URL         string `yaml:"url,omitempty"`
}

// SecurityRequirement ...
type SecurityRequirement = map[string][]string

// OpenAPI ...
type OpenAPI struct {
	Version      string                 `yaml:"openapi,omitempty"`
	Info         *Info                  `yaml:"info,omitempty"`
	Paths        map[string]*PathItem   `yaml:"paths,omitempty"`
	Servers      []*Server              `yaml:"servers,omitempty"`
	Components   *Components            `yaml:"components,omitempty"`
	Security     []SecurityRequirement  `yaml:"security,omitempty"`
	Tags         []*Tag                 `yaml:"tags,omitempty"`
	ExternalDocs *ExternalDocumentation `yaml:"externalDocs,omitempty"`
}

// Discriminator ...
type Discriminator struct {
}

// XML ...
type XML struct {
}

// Schema ...
type Schema struct {
	Ref              string                 `yaml:"$ref,omitempty"`
	Properties       map[string]*Schema     `yaml:"properties,omitempty"`
	Nullable         bool                   `yaml:"nullable,omitempty"`
	Discriminator    *Discriminator         `yaml:"discriminator,omitempty"`
	ReadOnly         bool                   `yaml:"readOnly,omitempty"`
	WriteOnly        bool                   `yaml:"writeOnly,omitempty"`
	XML              XML                    `yaml:"xml,omitempty"`
	ExternalDocs     *ExternalDocumentation `yaml:"externalDocs,omitempty"`
	Example          string                 `yaml:"example,omitempty"`
	Deprecated       bool                   `yaml:"deprecated,omitempty"`
	Type             string                 `yaml:"type,omitempty"`
	Format           string                 `yaml:"format,omitempty"`
	Title            string                 `yaml:"title,omitempty"`
	MultipleOf       int                    `yaml:"multipleOf,omitempty"`
	Maximum          int                    `yaml:"maximum,omitempty"`
	ExclusiveMaximum bool                   `yaml:"exclusiveMaximum,omitempty"`
	Minimum          int                    `yaml:"minimum,omitempty"`
	ExclusiveMinimum bool                   `yaml:"exclusiveMinimum,omitempty"`
	MaxLength        uint                   `yaml:"maxLength,omitempty"`
	MinLength        uint                   `yaml:"minLength,omitempty"`
	Pattern          string                 `yaml:"pattern,omitempty"`
	MaxItems         uint                   `yaml:"maxItems,omitempty"`
	MinItems         uint                   `yaml:"minItems,omitempty"`
	UniqueItems      bool                   `yaml:"uniqueItmes,omitempty"`
	MaxProperties    uint                   `yaml:"maxProperties,omitempty"`
	MinProperties    uint                   `yaml:"minProperties,omitempty"`
	Required         []string               `yaml:"required,omitempty"`
	Enum             []string               `yaml:"enum,omitempty"`
	Items            *Schema                `yaml:"items,omitempty"`
	AllOf            []*Schema              `yaml:"allOf,omitempty"`
	OneOf            []*Schema              `yaml:"oneOf,omitempty"`
	AnyOf            []*Schema              `yaml:"anyOf,omitempty"`
	Not              []*Schema              `yaml:"not,omitempty"`
}

// Parameter ...
type Parameter struct {
	Ref             string              `yaml:"$ref,omitempty"`
	Name            string              `yaml:"name,omitempty"`
	In              string              `yaml:"in,omitempty"`
	Description     string              `yaml:"description,omitempty"`
	Required        bool                `yaml:"required,omitempty"`
	Deprecated      bool                `yaml:"deprecated,omitempty"`
	AllowEmptyValue bool                `yaml:"allowEmptyValue,omitempty"`
	Style           string              `yaml:"style,omitempty"`
	Explode         bool                `yaml:"explode,omitempty"`
	AllowReserved   bool                `yaml:"allowReserved,omitempty"`
	Schema          *Schema             `yaml:"schema,omitempty"`
	Example         string              `yaml:"example,omitempty"`
	Examples        map[string]*Example `yaml:"examples,omitempty"`
}

// Example ...
type Example struct {
	Ref string `yaml:"$ref,omitempty"`
}

// RequestBody ...
type RequestBody struct {
	Ref         string                `yaml:"$ref,omitempty"`
	Description string                `yaml:"description,omitempty"`
	Content     map[string]*MediaType `yaml:"content,omitempty"`
	Required    bool                  `yaml:"required,omitempty"`
}

// Header ...
type Header struct {
	Ref string `yaml:"$ref,omitempty"`
}

// SecurityScheme ...
type SecurityScheme struct {
	Ref string `yaml:"$ref,omitempty"`
}

// Link ...
type Link struct {
	Ref string `yaml:"$ref,omitempty"`
}

// Callback ...
type Callback struct {
	Ref string `yaml:"$ref,omitempty"`
}

// Components ...
type Components struct {
	Schemas         map[string]*Schema         `yaml:"schemas,omitempty"`
	Responses       map[string]*Response       `yaml:"responses,omitempty"`
	Parameters      map[string]*Parameter      `yaml:"parameters,omitempty"`
	Examples        map[string]*Example        `yaml:"examples,omitempty"`
	RequestBodies   map[string]*RequestBody    `yaml:"requestBodies,omitempty"`
	Headers         map[string]*Header         `yaml:"headers,omitempty"`
	SecuritySchemes map[string]*SecurityScheme `yaml:"securitySchemes,omitempty"`
	Links           map[string]*Link           `yaml:"links,omitempty"`
	Callbacks       map[string]*Callback       `yaml:"callback,omitempty"`
}
