package helper

import (
	"testing"
	"time"
)

// TestEntity 测试实体
// 用于在测试过程中模拟数据结构，包含各种常见的数据类型
type TestEntity struct {
	ID        int       `orm:"id key auto" view:"detail,lite"`
	Name      string    `orm:"name" view:"detail,lite"`
	Value     float32   `orm:"value" view:"detail,lite"`
	TimeStamp time.Time `orm:"ts dateTime" view:"detail,lite"`
	Flag      bool      `orm:"flag" view:"detail,lite"`
	Items     []int     `orm:"items" view:"detail,lite"`
}

// TestUpdateSliceEntity 测试UpdateSliceEntity函数
// 场景说明: 验证从远程切片对象值更新到本地切片实体的功能
// 1. 创建一个包含多个实体的切片
// 2. 获取其远程切片对象值
// 3. 创建一个空的目标切片实体并使用UpdateSliceEntity更新它
// 4. 验证目标切片实体是否包含了正确数量的元素和字段值
func TestUpdateSliceEntity(t *testing.T) {
	// 创建原始切片实体
	originalEntities := []*TestEntity{
		{
			ID:        1,
			Name:      "Entity 1",
			Value:     123.456,
			TimeStamp: time.Now(),
			Flag:      true,
			Items:     []int{1, 2, 3},
		},
		{
			ID:        2,
			Name:      "Entity 2",
			Value:     789.012,
			TimeStamp: time.Now(),
			Flag:      false,
			Items:     []int{4, 5, 6},
		},
	}

	// 获取远程切片对象值
	sliceObjectValue, sliceObjectErr := GetSliceObjectValue(originalEntities)
	if sliceObjectErr != nil {
		t.Errorf("获取切片对象值失败: %s", sliceObjectErr.Error())
		return
	}

	// 创建目标切片实体用于更新
	targetEntities := []*TestEntity{}

	// 测试UpdateSliceEntity
	updateErr := UpdateSliceEntity(sliceObjectValue, &targetEntities)
	if updateErr != nil {
		t.Errorf("UpdateSliceEntity执行失败: %s", updateErr.Error())
		return
	}

	// 验证更新结果
	if len(targetEntities) != len(originalEntities) {
		t.Errorf("期望目标切片包含%d个实体，但实际获得%d个", len(originalEntities), len(targetEntities))
		return
	}

	for i, entity := range targetEntities {
		if entity.ID != originalEntities[i].ID {
			t.Errorf("实体%d: 期望ID为%d，但实际获得%d", i, originalEntities[i].ID, entity.ID)
			return
		}
		if entity.Name != originalEntities[i].Name {
			t.Errorf("实体%d: 期望Name为'%s'，但实际获得'%s'", i, originalEntities[i].Name, entity.Name)
			return
		}
	}
}

// TestObjectSerializationAndDeserialization 测试对象序列化和反序列化
// 场景说明: 验证对象的序列化和反序列化功能
// 1. 创建一个测试实体并获取其对象表示
// 2. 将对象序列化为字节数组
// 3. 从字节数组反序列化回对象
// 4. 验证反序列化后的对象是否保留了原始对象的属性和字段
func TestObjectSerializationAndDeserialization(t *testing.T) {
	// 创建测试实体
	testEntity := &TestEntity{
		ID:        1,
		Name:      "Serialization Test",
		Value:     123.456,
		TimeStamp: time.Now(),
		Flag:      true,
		Items:     []int{1, 2, 3},
	}

	// 获取对象
	objectPtr, objectErr := GetObject(testEntity)
	if objectErr != nil {
		t.Errorf("获取对象失败: %s", objectErr.Error())
		return
	}

	// 序列化对象
	serializedData, serializeErr := EncodeObject(objectPtr)
	if serializeErr != nil {
		t.Errorf("序列化对象失败: %s", serializeErr.Error())
		return
	}

	// 确保序列化数据不为空
	if len(serializedData) == 0 {
		t.Errorf("序列化数据不应为空")
		return
	}

	// 反序列化对象
	decodedObject, decodeErr := DecodeObject(serializedData)
	if decodeErr != nil {
		t.Errorf("反序列化对象失败: %s", decodeErr.Error())
		return
	}

	// 验证反序列化结果
	if decodedObject.GetName() != objectPtr.GetName() {
		t.Errorf("期望对象名称为'%s'，但实际获得'%s'", objectPtr.GetName(), decodedObject.GetName())
		return
	}

	if decodedObject.GetPkgPath() != objectPtr.GetPkgPath() {
		t.Errorf("期望包路径为'%s'，但实际获得'%s'", objectPtr.GetPkgPath(), decodedObject.GetPkgPath())
		return
	}

	// 字段数量应该相同
	if len(decodedObject.GetFields()) != len(objectPtr.GetFields()) {
		t.Errorf("期望字段数量为%d，但实际获得%d", len(objectPtr.GetFields()), len(decodedObject.GetFields()))
		return
	}
}
