package iotools

import "io"

// Discard provides io.Discard but casted
// to the full set of interfaces it supports.
// Useful if you want to bypass an io.Copy()
// call for the discard io.ReaderFrom{} impl.
var Discard = io.Discard.(interface {
	io.Writer
	io.StringWriter
	io.ReaderFrom
})
