package main

import "embed"

//go:embed all:web/dist
var WebDist embed.FS

//go:embed all:frontend/dist
var DesktopDist embed.FS
