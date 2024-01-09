// Package docs Code generated by swaggo/swag. DO NOT EDIT
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
        "/analyze/{analysisId}/status": {
            "get": {
                "description": "Checks the status of an ongoing analysis.",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "transactions"
                ],
                "summary": "Check analysis status",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Analysis ID",
                        "name": "analysisId",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/api.AnalysisStatus"
                        }
                    }
                }
            }
        },
        "/analyze/{fileID}": {
            "post": {
                "description": "Initiates the analysis of the uploaded transactions file.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "transactions"
                ],
                "summary": "Analyze transactions",
                "parameters": [
                    {
                        "type": "string",
                        "description": "File ID",
                        "name": "fileID",
                        "in": "path",
                        "required": true
                    },
                    {
                        "description": "Rules for detecting fraudulent transactions",
                        "name": "rules",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/api.AnalyzeRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "analysis_id",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    }
                }
            }
        },
        "/health": {
            "get": {
                "description": "A simple health check endpoint to verify if the API is up and running.",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "health"
                ],
                "summary": "Health check",
                "responses": {
                    "200": {
                        "description": "message",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    }
                }
            }
        },
        "/results/{analysisId}": {
            "get": {
                "description": "Retrieves the results of the completed analysis.",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "transactions"
                ],
                "summary": "Get analysis results",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Analysis ID",
                        "name": "analysisId",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/api.AnalysisResult"
                        }
                    }
                }
            }
        },
        "/upload": {
            "post": {
                "description": "This endpoint is used to upload a CSV file containing transaction data.",
                "consumes": [
                    "multipart/form-data"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "transactions"
                ],
                "summary": "Upload transactions",
                "parameters": [
                    {
                        "type": "file",
                        "description": "Transaction file",
                        "name": "file",
                        "in": "formData",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "file_id",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "api.AnalysisResult": {
            "type": "object",
            "properties": {
                "result": {
                    "type": "string"
                }
            }
        },
        "api.AnalysisStatus": {
            "type": "object",
            "properties": {
                "status": {
                    "type": "string"
                }
            }
        },
        "api.AnalyzeRequest": {
            "type": "object",
            "properties": {
                "rules": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                }
            }
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "",
	Host:             "",
	BasePath:         "",
	Schemes:          []string{},
	Title:            "",
	Description:      "",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
