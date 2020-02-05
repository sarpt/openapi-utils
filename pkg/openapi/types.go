package openapi

type Info struct {
	Title       string `yaml:"title"`
	Description string `yaml:"description"`
	Version     string `yaml:"version"`
}

type JSON struct {
	Schema Schema `yaml:"schema"`
}

type Content struct {
	JSON `yaml:"application/json"`
}

type Response struct {
	Description string `yaml:"description"`
	Content     `yaml:"content"`
	Ref         string `yaml:"$ref"`
}

type Get struct {
	Summary     string               `yaml:"summary"`
	Description string               `yaml:"description"`
	Responses   map[string]*Response `yaml:"responses"`
}

type Path struct {
	Get Get    `yaml:"get"`
	Ref string `yaml:"$ref"`
}

type Server struct {
	URL         string `yaml:"url"`
	Description string `yaml:"description"`
}

type OpenApi struct {
	Version    string           `yaml:"openapi"`
	Info       Info             `yaml:"info"`
	Paths      map[string]*Path `yaml:"paths"`
	Servers    []Server         `yaml:"servers"`
	Components Components       `yaml:"components"`
}

type Property struct {
	Type string `yaml:"type"`
}

type Schema struct {
	Properties map[string]Property `yaml:"properties"`
}

type Components struct {
	Schemas   map[string]Schema    `yaml:"schemas"`
	Responses map[string]*Response `yaml:"responses"`
}

type References struct {
	PathReferences     map[string]*Path
	ResponseReferences map[string]*Response
}
