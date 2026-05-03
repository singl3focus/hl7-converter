package hl7converter

var jsonSchema = `{
    "$schema": "https://json-schema.org/draft/2020-12/schema",
    "$id": "./config.schema.json",
    "title": "modifications",
    "description": "Representations of formats",
    "type": "object",
    "additionalProperties": {
        "type": "object",
        "required": [
            "component_separator",
            "component_array_separator",
            "field_separator",
            "line_separator",
            "tags_info"
        ],
        "properties": {
            "component_separator": {
                "type": "string",
                "minLength": 1
            },
            "component_array_separator": {
                "type": "string",
                "minLength": 1
            },
            "field_separator": {
                "type": "string",
                "minLength": 1
            },
            "line_separator": {
                "type": "string",
                "minLength": 1
            },
            "types": {
                "type": "object",
                "additionalProperties": {
                    "type": "array",
                    "description": "array of arrays tags which define some msg type",
                    "items": {
                        "type": "array",
                        "description": "array tags which define some msg type",
                        "items": {
                            "type": "string"
                        },
                        "uniqueItems": true
                    }
                }
            },
            "aliases": {
                "type": "object",
                "additionalProperties": {
                    "type": "string"
                }
            },
            "tags_info": {
                "type": "object",
                "required": ["positions", "tags"],
                "properties": {
                    "positions": {
                        "type": "object",
                        "additionalProperties": {
                            "type": "string"
                        },
                        "minProperties": 1
                    },
                    "tags": {
                        "type": "object",
                        "minProperties": 1,
                        "additionalProperties": {
                            "type": "object",
                            "description": "single tag",
                            "required": ["linked", "fields_number", "template"],
                            "properties": {
                                "options": {
                                    "description": "options for the tag",
                                    "type": "array",
                                    "items": {
                                        "type": "string",
                                        "enum": ["autofill"]
                                    },
                                    "uniqueItems": true
                                },
                                "linked": {
                                    "description": "link by other tag",
                                    "type": "string"
                                },
                                "fields_number": {
                                    "description": "count of fields in row with tag",
                                    "type": "integer"
                                },
                                "template": {
                                    "description": "template for filling values by this tag",
                                    "type": "string"
                                }
                            },
                            "additionalProperties": false
                        }
                    }
                },
                "additionalProperties": false
            }
        },
        "additionalProperties": false
    }
}`
