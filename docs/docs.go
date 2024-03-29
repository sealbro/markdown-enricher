// Package docs GENERATED BY SWAG; DO NOT EDIT
// This file was generated by swaggo/swag
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {},
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/v1/markdown/enrich": {
            "get": {
                "description": "Get enriched markdown elements",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "state"
                ],
                "parameters": [
                    {
                        "type": "string",
                        "description": "md file url (https://raw.githubusercontent.com/avelino/awesome-go/main/README.md)",
                        "name": "md_file_url",
                        "in": "query",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "markdown enriched",
                        "schema": {
                            "$ref": "#/definitions/model.MarkdownEnriched"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "model.GitHubRepoInfo": {
            "type": "object",
            "properties": {
                "f": {
                    "type": "integer"
                },
                "lc": {
                    "type": "string"
                },
                "o": {
                    "type": "string"
                },
                "r": {
                    "type": "string"
                },
                "s": {
                    "type": "integer"
                }
            }
        },
        "model.LinkEnriched": {
            "type": "object",
            "properties": {
                "i": {
                    "description": "Url  string          ` + "`" + `json:\"url\"` + "`" + `",
                    "$ref": "#/definitions/model.GitHubRepoInfo"
                }
            }
        },
        "model.MarkdownEnriched": {
            "type": "object",
            "properties": {
                "links": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/model.LinkEnriched"
                    }
                }
            }
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "1.0",
	Host:             "localhost:8080",
	BasePath:         "/markdown-enricher/api",
	Schemes:          []string{"http"},
	Title:            "Markdown enricher",
	Description:      "",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
