package provider

import (
	"log/slog"

	cd "github.com/muidea/magicCommon/def"
)

// logError 统一错误日志记录辅助函数
// methodName: 方法名称
// operation: 操作描述（如 "s.getTypeFunc"）
// err: 错误对象
func logError(methodName, operation string, err *cd.Error) {
	if err != nil {
		if operation != "" {
			slog.Error("message")
		} else {
			slog.Error("message")
		}
	}
}

// withErrorLog 包装函数调用，自动记录错误
// 适用于简单的函数调用后错误处理
func withErrorLog(methodName, operation string, fn func() (*cd.Error, error)) *cd.Error {
	if err, _ := fn(); err != nil {
		logError(methodName, operation, err)
		return err
	}
	return nil
}
