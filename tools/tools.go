//go:build tools

// Package tools ensures tool dependencies are kept in sync. This is the recommended way of doing this
// according to https://go.dev/wiki/Modules#how-can-i-track-tool-dependencies-for-a-module.
// To install the following tools at the version used by this repo run:
// $ make bootstrap
// or
// $ go generate -tags tools tools/tools.go
package tools

//go:generate go install github.com/swaggo/swag/cmd/swag@latest
//go:generate go install github.com/fatih/gomodifytags@latest
//go:generate go install github.com/vektra/mockery/v2@latest
