// Package hl7converter transforms row/tag messages with JSON mappings.
//
// Positions include the tag at position 1. A template link such as <A-2.10>
// selects component 10 of the first data field. Aliases may select the tag.
// This notation does not implement HL7-specific MSH numbering.
//
// A Converter can serve concurrent Convert calls while its config is immutable.
// IdentifyMsg reads shared config without changing it. Results are mutable and
// need external synchronization for concurrent mutation. Parsing rejects oversized
// Scanner tokens rather than returning a successful partial message.
//
// JavaScript hooks are trusted code. Their VM is allocated on the first UseScript
// call and reused by that Result. CheckMappings verifies versioned fixtures;
// ExplainMapping describes static dependencies without a message payload.
package hl7converter
