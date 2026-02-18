package provider

import (
	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/log"
)

// logError 统一错误日志记录辅助函数
// methodName: 方法名称
// operation: 操作描述（如 "s.getTypeFunc"）
// err: 错误对象
func logError(methodName, operation string, err *cd.Error) {
	if err != nil {
		if operation != "" {
			log.Errorf("%s failed, %s error:%v", methodName, operation, err.Error())
		} else {
			log.Errorf("%s failed, error:%v", methodName, err.Error())
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
