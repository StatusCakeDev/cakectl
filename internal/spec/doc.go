// Package spec describes the desired state of each StatusCake resource.
//
// A spec is what a person asks for: the fields they may set when creating or
// updating a resource. Every field is a [gofp.Option] so that a spec can say
// "leave this alone" as well as "set this to the zero value", which is what
// makes partial updates and flag overrides possible. A spec holds no knowledge
// of the StatusCake SDK, of files or of flags; those are the concern of the
// layers around it.
package spec
