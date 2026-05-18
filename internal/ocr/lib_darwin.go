//go:build darwin

package ocr

import _ "embed"

//go:embed assets/libonnxruntime.dylib
var onnxLibBytes []byte

const onnxLibFileName = "libonnxruntime.dylib"
