package main

// ConfigurationJSONSchema is the entire json string from configuration.json.schema
const ConfigurationJSONSchema = `{
	"$schema": "http://json-schema.org/schema",
	"$id": ".portfoliodb.yml",
	"definitions": {
		"validate_check": {
			"type": [
				"string",
				"boolean"
			],
			"enum": [
				"warn",
				"error",
				"fatal",
				"info",
				"off",
				false,
				true
			]
		}
	},
	"properties": {
		"build steps": {
			"type": "object",
			"properties": {
				"extract colors": {
					"type": "object",
					"properties": {
						"extract": {
							"enum": [
								"primary",
								"secondary",
								"tertiary"
							]
						},
						"default file name": {
							"type": "array",
							"items": {
								"type": "string"
							}
						}
					}
				},
				"make gifs": {
					"type": "object",
					"properties": {
						"file name template": {
							"type": "string"
						}
					}
				},
				"make thumbnails": {
					"type": "object",
					"properties": {
						"widths": {
							"type": "array",
							"items": {
								"type": "integer"
							}
						},
						"input file": {
							"type": "string"
						},
						"file name template": {
							"type": "string"
						}
					}
				}
			}
		},
		"features": {
			"type": "object",
			"properties": {
				"made with": {
					"type": "boolean"
				},
				"media hoisting": {
					"type": "boolean"
				}
			}
		},
		"validate": {
			"type": "object",
			"properties": {
				"checks": {
					"type": "object",
					"properties": {
						"schema compliance": {
							"$ref": "#/definitions/validate_check",
							"default": "fatal"
						},
						"work folder uniqueness": {
							"$ref": "#/definitions/validate_check",
							"default": "fatal"
						},
						"work folder safeness": {
							"$ref": "#/definitions/validate_check",
							"default": "error"
						},
						"yaml header": {
							"$ref": "#/definitions/validate_check",
							"default": "error"
						},
						"title presence": {
							"$ref": "#/definitions/validate_check",
							"default": "error"
						},
						"title uniqueness": {
							"$ref": "#/definitions/validate_check",
							"default": "error"
						},
						"tags presence": {
							"$ref": "#/definitions/validate_check",
							"default": "warn"
						},
						"tags knowledge": {
							"$ref": "#/definitions/validate_check",
							"default": "error"
						},
						"working media": {
							"$ref": "#/definitions/validate_check",
							"default": "warn"
						},
						"working urls": {
							"$ref": "#/definitions/validate_check",
							"default": false
						}
					}
				}
			}
		},
		"markdown": {
			"type": "object",
			"properties": {
				"abbreviations": {
					"type": "boolean"
				},
				"definition lists": {
					"type": "boolean"
				},
				"admonitions": {
					"type": "boolean"
				},
				"markdown in HTML": {
					"type": "boolean"
				},
				"new-line-to-line-break": {
					"type": "boolean"
				},
				"smarty pants": {
					"type": "boolean"
				},
				"anchored headings": {
					"type": "boolean"
				},
				"custom syntaxes": {
					"type": "array",
					"items": {
						"type": "object",
						"properties": {
							"from": {
								"type": "string"
							},
							"to": {
								"type": "string"
							}
						}
					}
				}
			}
		}
	}
}
`