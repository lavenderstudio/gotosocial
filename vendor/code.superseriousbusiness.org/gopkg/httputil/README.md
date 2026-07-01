# httputil

A series of types, wrappers and utility methods to aid in the construction of HTTP server applications. These largely follow in the style of github.com/gin-gonic/gin, but much simplified... Both because some of it is a nice API to work with, and this package was designed to migrate GoToSocial away from using Gin now they're using LLMs to write their code.

The general aim of this package is to have as few dependencies as possible, (i.e. just gopkg and stdlib), and wrap the standard library net/http as little as possible. The biggest change is the Context{} type and its handlers, but it aims to be as flexible as possible for your own extensions if you want to use it, and much of the other utilities are designed to be usable with just net/http.

`./context_util.go` is where the code gets opinionated. it expects you to be using httputil.Context{} based handlers, and logs to the global `gopkg/log` logger instance.
