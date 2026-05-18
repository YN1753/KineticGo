package ocr

import (
	"bytes"
	_ "embed"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"strings"
	"sync"

	"github.com/nfnt/resize"
	ort "github.com/yalue/onnxruntime_go"
)

//go:embed assets/common_old.onnx
var onnxModelData []byte

type OcrEngine struct {
	session *ort.DynamicAdvancedSession
}

var (
	globalEngineOnce sync.Once
	globalEngine     *OcrEngine
	globalEngineErr  error
)

// GetGlobalEngine 返回全局单例 OCR 引擎,首次调用时自动提取内嵌动态库并初始化.
func GetGlobalEngine() (*OcrEngine, error) {
	globalEngineOnce.Do(func() {
		libPath, err := ExtractLibrary()
		if err != nil {
			globalEngineErr = fmt.Errorf("提取 ONNX 动态库失败: %w", err)
			return
		}
		globalEngine, globalEngineErr = NewOcrEngine(libPath)
	})
	return globalEngine, globalEngineErr
}

func NewOcrEngine(dllPath string) (*OcrEngine, error) {
	// 初始化 ONNX 运行环境
	ort.SetSharedLibraryPath(dllPath)
	err := ort.InitializeEnvironment()
	if err != nil {
		return nil, fmt.Errorf("初始化 ONNX 环境失败: %v", err)
	}

	// 直接使用内嵌的 onnxModelData 字节流创建 Session
	session, err := ort.NewDynamicAdvancedSessionWithONNXData(
		onnxModelData,
		[]string{"input1"},
		[]string{"387"},
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("创建 Session 失败: %v", err)
	}

	return &OcrEngine{session: session}, nil
}

// Close 负责释放 C++ 底层内存（极其重要）
func (e *OcrEngine) Close() {
	if e.session != nil {
		e.session.Destroy()
	}
	ort.DestroyEnvironment()
}

func RecognizeCaptcha(dllPath string, imgBytes []byte) (string, error) {

	engine, err := NewOcrEngine(dllPath)
	if err != nil {
		return "", fmt.Errorf("无法初始化 OCR 引擎: %w", err)
	}

	defer engine.Close()

	return engine.Recognize(imgBytes)
}

// 核心推理函数
func (e *OcrEngine) Recognize(imgBytes []byte) (string, error) {
	// 1. 图像预处理
	tensorData, width, err := preprocessImage(imgBytes)
	if err != nil {
		return "", err
	}

	inputShape := ort.NewShape(1, 1, 64, int64(width))
	inputTensor, err := ort.NewTensor(inputShape, tensorData)
	if err != nil {
		return "", err
	}
	defer inputTensor.Destroy()

	outputs := make([]ort.Value, 1)

	// 3. 执行推理，将空切片传给模型
	err = e.session.Run(
		[]ort.Value{inputTensor},
		outputs,
	)
	if err != nil {
		return "", fmt.Errorf("模型推理失败: %v", err)
	}

	// 4. 将底层自动分配好并填满数据的内存，转换为我们需要的 Float32 张量
	outputTensor, ok := outputs[0].(*ort.Tensor[float32])
	if !ok {
		return "", fmt.Errorf("无法将底层自动分配的输出转换为 Float32 张量")
	}

	defer outputTensor.Destroy()

	outData := outputTensor.GetData()
	outShape := outputTensor.GetShape()

	// 动态解析输出维度
	var seqLen, charsetSize int
	if len(outShape) >= 3 {
		if outShape[0] == 1 {
			seqLen = int(outShape[1])
			charsetSize = int(outShape[2])
		} else {
			seqLen = int(outShape[0])
			charsetSize = int(outShape[2])
		}
	} else {
		return "", fmt.Errorf("未知的输出形状: %v", outShape)
	}

	return decodeCTC(outData, seqLen, charsetSize), nil
}
func preprocessImage(imgBytes []byte) ([]float32, int, error) {
	img, _, err := image.Decode(bytes.NewReader(imgBytes))
	if err != nil {
		return nil, 0, err
	}

	bounds := img.Bounds()
	ratio := 64.0 / float64(bounds.Dy())
	newWidth := int(float64(bounds.Dx()) * ratio)
	resizedImg := resize.Resize(uint(newWidth), 64, img, resize.Bilinear)

	tensorData := make([]float32, 0, 64*newWidth)

	for y := 0; y < 64; y++ {
		for x := 0; x < newWidth; x++ {
			r, g, b, _ := resizedImg.At(x, y).RGBA()
			gray := (0.299*float64(r>>8) + 0.587*float64(g>>8) + 0.114*float64(b>>8))
			normalized := float32((gray / 127.5) - 1.0)
			tensorData = append(tensorData, normalized)
		}
	}
	return tensorData, newWidth, nil
}

func decodeCTC(output []float32, seqLen, charsetSize int) string {
	var b strings.Builder
	lastIndex := -1

	for i := 0; i < seqLen; i++ {
		maxProb := float32(-math.MaxFloat32)
		bestIndex := 0

		for j := 0; j < charsetSize; j++ {
			idx := i*charsetSize + j
			if output[idx] > maxProb {
				maxProb = output[idx]
				bestIndex = j
			}
		}
		if bestIndex != 0 && bestIndex != lastIndex {
			if bestIndex < len(CommonOldCharset) {
				b.WriteString(CommonOldCharset[bestIndex])
			}
		}
		lastIndex = bestIndex
	}

	return b.String()
}
