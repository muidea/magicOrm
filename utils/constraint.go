package utils

import (
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

// parseConstraints 将 "req,min=1:100,ro" 解析为结构化 Map
func ParseConstraints(tagStr string) models.Constraints {
	directives := constraintsImpl{}

	// 1. 按逗号拆分各个指令单元
	units := strings.SplitSeq(tagStr, ",")

	for unit := range units {
		unit = strings.TrimSpace(unit)
		if unit == "" {
			continue
		}

		dir := directiveImpl{}

		// 2. 检查是否存在参数 (是否包含 '=')
		if strings.Contains(unit, "=") {
			parts := strings.SplitN(unit, "=", 2)
			dir.key = models.Key(strings.TrimSpace(parts[0]))
			dir.hasArgs = true
			// 3. 按冒号拆分参数
			argParts := strings.Split(parts[1], ":")
			for _, arg := range argParts {
				dir.args = append(dir.args, strings.TrimSpace(arg))
			}
		} else {
			dir.key = models.Key(unit)
			dir.hasArgs = false
		}

		directives[dir.key] = dir
	}

	return &directives
}
