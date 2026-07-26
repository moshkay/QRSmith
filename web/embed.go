// Package web embeds the static frontend assets served by the QRForge server.
package web

import "embed"

//go:embed static/*
var Static embed.FS
