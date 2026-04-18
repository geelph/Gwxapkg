package formatter

import (
	"fmt"
	"strings"
)

// Formatter 是一个文件格式化器接口
type Formatter interface {
	Format([]byte) ([]byte, error)
}

// FileFormatter 支持基于文件路径的扩展格式化
type FileFormatter interface {
	Formatter
	FormatFile(input []byte, filePath string) ([]byte, *DeobfuscationResult, error)
}

// 注册所有格式化器
var formatters = map[string]Formatter{}

// RegisterFormatter 注册文件扩展名对应的格式化器
func RegisterFormatter(ext string, formatter Formatter) {
	formatters[strings.ToLower(ext)] = formatter
}

// GetFormatter 返回文件扩展名对应的格式化器
func GetFormatter(ext string) (Formatter, error) {
	formatter, exists := formatters[strings.ToLower(ext)]
	if !exists {
		return nil, fmt.Errorf("不支持的文件类型: %s", ext)
	}
	return formatter, nil
}
