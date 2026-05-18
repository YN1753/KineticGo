package ocr

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

var (
	extractOnce  sync.Once
	extractedLib string
	extractErr   error
)

func ExtractLibrary() (string, error) {
	extractOnce.Do(func() {
		if len(onnxLibBytes) == 0 {
			extractErr = fmt.Errorf("当前平台未内嵌 ONNX Runtime 动态库: %s/%s",
				runtime.GOOS, runtime.GOARCH)
			return
		}
		dir := filepath.Join(os.TempDir(), "kineticgo-ocr")
		if err := os.MkdirAll(dir, 0o755); err != nil {
			extractErr = err
			return
		}

		target := filepath.Join(dir, onnxLibFileName)
		if st, err := os.Stat(target); err == nil && st.Size() == int64(len(onnxLibBytes)) {
			extractedLib = target
			return
		}

		if err := os.WriteFile(target, onnxLibBytes, 0o755); err != nil {
			extractErr = err
			return
		}
		extractedLib = target
	})
	return extractedLib, extractErr
}
