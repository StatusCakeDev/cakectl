// Package view turns what the API returns into what the user sees.
//
// Each type is a read model: a flat, stable shape with JSON field names, that
// can be rendered as a table or block of fields for people
// ([printer.TextFormatter]) and encoded as JSON or YAML for scripts. A view
// never includes more than its command needs: lists omit secrets and personal
// data such as heartbeat URLs and contact details, which only an explicit
// `describe` shows.
//
// The description of a resource carries its writable fields as a [spec], in the
// same shape that `-f` reads, so that a description can be edited and applied
// back.
package view
