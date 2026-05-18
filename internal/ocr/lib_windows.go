//go:build windows

package ocr

import _ "embed"

//go:embed assets/onnxruntime.dll
var onnxLibBytes []byte

const onnxLibFileName = "onnxruntime.dll"
