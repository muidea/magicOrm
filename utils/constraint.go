package utils

import (
	"fmt"
	"strings"

	"github.com/muidea/magicOrm/models"
)

// directiveImpl 存储单个指令及其参数
type directiveImpl struct {
	key     models.Key // 指令名称，如 "min"
	hasArgs bool       // 是否带参数
	args    []string   // 参数列表，如 ["18", "60"]
}

func (m directiveImpl) Key() models.Key {
	return m.key
}

func (m directiveImpl) HasArgs() bool {
	return m.hasArgs
}

func (m directiveImpl) Args() []string {
	return m.args
}

type constraintsImpl map[models.Key]directiveImpl

func (m constraintsImpl) Has(key models.Key) bool {
	_, ok := m[key]
	return ok
}

func (m constraintsImpl) Get(key models.Key) (models.Directive, bool) {
	val, ok := m[key]
	return val, ok
}

func (m constraintsImpl) Directives() []models.Directive {
	ret := make([]models.Directive, 0, len(m))
	for _, v := range m {
		ret = append(ret, &v)
	}

	return ret
}

// parseConstraints 将 "req,min=1:100,ro" 解析为结构化 Map
func ParseConstraints(tagStr string) models.Constraints {
	directives := constraintsImpl{}

	// 1. 预处理：保护正则表达式中的逗号
	// 正则表达式可能包含逗号，如 {2,64}，我们需要保护这些逗号不被分割
	protectedStr := tagStr
	rePatterns := []string{}

	// 查找所有 re=... 模式
	start := 0
	for {
		// 查找 "re="
		reStart := strings.Index(protectedStr[start:], "re=")
		if reStart == -1 {
			break
		}
		reStart += start

		// 找到 re= 后面的内容
		valueStart := reStart + 3 // "re=" 的长度

		// 我们需要找到这个正则表达式的结束位置
		// 正则表达式可能包含逗号，所以不能简单地按逗号分割
		// 我们假设正则表达式是最后一个指令，或者后面跟着其他指令
		// 我们查找下一个指令的开始（下一个逗号后面跟着非空格字符）
		valueEnd := len(protectedStr)
		for i := valueStart; i < len(protectedStr); i++ {
			// 如果遇到逗号，且这个逗号不在花括号内（简单判断）
			if protectedStr[i] == ',' {
				// 检查前面的字符，如果是数字或字母，可能是正则表达式的一部分
				// 简单起见，我们假设如果逗号前面是数字，且后面也是数字，那么它是正则表达式的一部分
				if i > valueStart && i+1 < len(protectedStr) {
					prevChar := protectedStr[i-1]
					nextChar := protectedStr[i+1]
					if (prevChar >= '0' && prevChar <= '9') && (nextChar >= '0' && nextChar <= '9') {
						// 这可能是正则表达式中的逗号，如 {2,64}
						continue
					}
				}
				// 否则，这是指令分隔符
				valueEnd = i
				break
			}
		}

		// 提取正则表达式
		regexPattern := protectedStr[valueStart:valueEnd]
		rePatterns = append(rePatterns, regexPattern)

		// 用占位符替换正则表达式
		placeholder := fmt.Sprintf("__REGEX_PLACEHOLDER_%d__", len(rePatterns)-1)
		protectedStr = protectedStr[:valueStart] + placeholder + protectedStr[valueEnd:]

		start = valueStart + len(placeholder)
	}

	// 2. 按逗号拆分各个指令单元
	units := strings.SplitSeq(protectedStr, ",")

	for unit := range units {
		unit = strings.TrimSpace(unit)
		if unit == "" {
			continue
		}

		dir := directiveImpl{}

		// 3. 检查是否存在参数 (是否包含 '=')
		if strings.Contains(unit, "=") {
			parts := strings.SplitN(unit, "=", 2)
			dir.key = models.Key(strings.TrimSpace(parts[0]))
			dir.hasArgs = true

			// 4. 恢复占位符为原始正则表达式
			argValue := parts[1]
			for i, placeholder := range rePatterns {
				argValue = strings.Replace(argValue, fmt.Sprintf("__REGEX_PLACEHOLDER_%d__", i), placeholder, 1)
			}

			// 5. 特殊处理 re 约束，正则表达式可能包含冒号
			if dir.key == models.KeyRegexp {
				// re 约束的参数是完整的正则表达式，不按冒号分割
				dir.args = []string{argValue}
			} else {
				// 其他约束按冒号拆分参数
				argParts := strings.Split(argValue, ":")
				for _, arg := range argParts {
					dir.args = append(dir.args, strings.TrimSpace(arg))
				}
			}
		} else {
			dir.key = models.Key(unit)
			dir.hasArgs = false
		}

		directives[dir.key] = dir
	}

	return &directives
}
