package hl7converter

var jsonSchema = `{
    "$schema": "https://json-schema.org/draft/2020-12/schema",
    "$id": "./config.schema.json",
    
    "title": "modifications",
    "description": "Represenatations of formats",
    "type": "object",
    
    "additionalProperties": {
        "type": "object",
        "properties": {
            "component_separator" : {
                "type": "string"
            },
            "field_separator": {
                "type": "string"
            },
            "line_separator": {
                "type": "string"
            },

            "types" : {
                "type": "object",
                "additionalProperties": {
                    "type": "array",
                    "description": "array of arrrays tags which define some msg type",
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

            "tags_info": {
                "positions" : {
                    "type": "object",
                    "additionalProperties": {
                        "type": "object"
                    }
                },    

                "tags": {
                    "type": "object",

                    "additionalProperties": {
                        "type": "object",
                        "description": "single tag",

                        "properties": {
                            "options": {
                                "description": "options for the tag",
                                "type": "array",
                                "items": {
                                    "type": "string",
                                    "examples" : "autofill"
                                },

                                "uniqueItems": true
                            },

                            "linked" : {
                                "description": "link by other tag",
                                "type": "string"
                            },

                            "fields_number" : {
                                "description": "count of fields in row with tag",
                                "type": "integer"
                            },

                            "template" : {
                                "description": "template for filling values by this tag",
                                "type": "string"
                            }
                        }
                    },

                    "required": ["linked", "fields_number", "template"]
                }
            }  
        },
    
        "required": ["component_separator", "field_separator", "line_separator", "tags_info"]
    }    
}`