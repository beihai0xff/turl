//go:build tools

// Package tools ensures tool dependencies are kept in sync.  This is the
// recommended way of doing this according to
// https://github.com/golang/go/wiki/Modules#go mod download how-can-i-track-tool-dependencies-for-a-module
// To install the following tools at the version used by this repo run:
// $ make bootstrap
// or
// $ go generate -tags tools tools/tools.go
package tools

//go:generate go install github.com/swaggo/swag/cmd/swag@latest
//go:generate go install github.com/fatih/gomodifytags@latest
//go:generate go install go.uber.org/mock/mockgen@latest
