package test

import (
	"testing"
	"time"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/util"
	"github.com/muidea/magicOrm/orm"
	"github.com/muidea/magicOrm/provider"
)

// TestUpdateRelationDiffReference 验证引用关系 Update 按差异增量更新：只增删链接，不删除关联实体
func TestUpdateRelationDiffReference(t *testing.T) {
	orm.Initialize()
	defer orm.Uninitialized()

	localProvider := provider.NewLocalProvider("updateRelationDiff", nil)
	o1, err := orm.NewOrm(localProvider, config, "abc")
	defer o1.Release()
	if err != nil {
		t.Fatalf("new Orm failed: %v", err)
	}

	entityList := []any{&Simple{}, &Reference{}, &Compose{}}
	modelList, modelErr := registerLocalModel(localProvider, entityList)
	if modelErr != nil {
		t.Fatalf("register model failed: %v", modelErr)
	}
	if err = dropModel(o1, modelList); err != nil {
		t.Fatalf("drop model failed: %v", err)
	}
	if err = createModel(o1, modelList); err != nil {
		t.Fatalf("create model failed: %v", err)
	}

	ts, _ := time.ParseInLocation(util.CSTLayout, "2018-01-02 15:04:05", time.Local)
	s1 := Simple{I8: 12, I16: 23, I32: 34, I64: 45, Name: "test", Value: 12.345, F64: 23.456, TimeStamp: ts, Flag: true}
	s1Model, _ := localProvider.GetEntityModel(s1, true)
	s1Model, err = o1.Insert(s1Model)
	if err != nil {
		t.Fatalf("insert simple failed: %v", err)
	}
	s1 = s1Model.Interface(false).(Simple)

	r1 := makeReference(ts, "ref1", 1, 1)
	r1Model, _ := localProvider.GetEntityModel(r1, true)
	r1Model, err = o1.Insert(r1Model)
	if err != nil {
		t.Fatalf("insert r1 failed: %v", err)
	}
	r1 = r1Model.Interface(false).(Reference)

	r2 := makeReference(ts, "ref2", 2, 2)
	r2Model, _ := localProvider.GetEntityModel(r2, true)
	r2Model, err = o1.Insert(r2Model)
	if err != nil {
		t.Fatalf("insert r2 failed: %v", err)
	}
	r2 = r2Model.Interface(false).(Reference)

	// Compose 引用 r1：ReferencePtr=r1, ReferencePtrArray=[r1]
	c1 := &Compose{
		Name:              "compose1",
		Simple:            s1,
		SimplePtr:         &s1,
		SimpleArray:       []Simple{s1},
		SimplePtrArray:    []*Simple{&s1},
		Reference:         r1,
		ReferencePtr:      &r1,
		ReferenceArray:    []Reference{r1},
		ReferencePtrArray: []*Reference{&r1},
	}
	c1Model, _ := localProvider.GetEntityModel(c1, true)
	c1Model, err = o1.Insert(c1Model)
	if err != nil {
		t.Fatalf("insert compose failed: %v", err)
	}
	c1 = c1Model.Interface(true).(*Compose)

	// Update：改为引用 r2；ReferencePtr=r2, ReferencePtrArray=[r2]
	// 差异更新应只删 (compose->r1) 的链接、插 (compose->r2) 的链接，不删除 r1 实体
	c1.ReferencePtr = &r2
	c1.ReferencePtrArray = []*Reference{&r2}
	c1Model, _ = localProvider.GetEntityModel(c1, true)
	_, err = o1.Update(c1Model)
	if err != nil {
		t.Fatalf("update compose failed: %v", err)
	}

	// 验证 r1、r2 仍存在（引用关系不应删除被引用的实体）
	qR1 := &Reference{ID: r1.ID}
	qR1Model, _ := localProvider.GetEntityModel(qR1, true)
	qR1Model, err = o1.Query(qR1Model)
	if err != nil {
		t.Fatalf("query r1 after update failed: %v", err)
	}
	if qR1Model.Interface(false).(Reference).Name != "ref1" {
		t.Errorf("r1 entity should still exist after update")
	}

	qR2 := &Reference{ID: r2.ID}
	qR2Model, _ := localProvider.GetEntityModel(qR2, true)
	qR2Model, err = o1.Query(qR2Model)
	if err != nil {
		t.Fatalf("query r2 after update failed: %v", err)
	}
	if qR2Model.Interface(false).(Reference).Name != "ref2" {
		t.Errorf("r2 entity should still exist after update")
	}

	// 验证 Compose 查出来关系为 r2：ReferencePtrArray 必为 [r2]（引用切片差异更新）
	qC := &Compose{ID: c1.ID}
	qCModel, _ := localProvider.GetEntityModel(qC, true)
	qCModel, err = o1.Query(qCModel)
	if err != nil {
		t.Fatalf("query compose after update failed: %v", err)
	}
	got := qCModel.Interface(true).(*Compose)
	if len(got.ReferencePtrArray) != 1 || got.ReferencePtrArray[0].ID != r2.ID {
		t.Errorf("compose ReferencePtrArray should be [r2] after update, got %v", got.ReferencePtrArray)
	}
	// 单值引用 ReferencePtr 若被正确加载则应为 r2
	if got.ReferencePtr != nil && got.ReferencePtr.ID != r2.ID {
		t.Errorf("compose ReferencePtr should be r2 when loaded, got id %d", got.ReferencePtr.ID)
	}
}

// TestUpdateRelationDiffReferencePartial 引用切片：部分替换（[r1,r2] -> [r2,r3]），仅删 r1 链接、插 r3 链接
func TestUpdateRelationDiffReferencePartial(t *testing.T) {
	orm.Initialize()
	defer orm.Uninitialized()

	localProvider := provider.NewLocalProvider("updateRelationDiffPartial", nil)
	o1, err := orm.NewOrm(localProvider, config, "abc")
	defer o1.Release()
	if err != nil {
		t.Fatalf("new Orm failed: %v", err)
	}

	entityList := []any{&Simple{}, &Reference{}, &Compose{}}
	modelList, modelErr := registerLocalModel(localProvider, entityList)
	if modelErr != nil {
		t.Fatalf("register model failed: %v", modelErr)
	}
	if err = dropModel(o1, modelList); err != nil {
		t.Fatalf("drop model failed: %v", err)
	}
	if err = createModel(o1, modelList); err != nil {
		t.Fatalf("create model failed: %v", err)
	}

	ts, _ := time.ParseInLocation(util.CSTLayout, "2018-01-02 15:04:05", time.Local)
	s1 := Simple{I8: 12, I16: 23, I32: 34, I64: 45, Name: "test", Value: 12.345, F64: 23.456, TimeStamp: ts, Flag: true}
	s1Model, _ := localProvider.GetEntityModel(s1, true)
	s1Model, _ = o1.Insert(s1Model)
	s1 = s1Model.Interface(false).(Simple)

	r1 := makeReference(ts, "r1", 1, 1)
	r1Model, _ := localProvider.GetEntityModel(r1, true)
	r1Model, _ = o1.Insert(r1Model)
	r1 = r1Model.Interface(false).(Reference)
	r2 := makeReference(ts, "r2", 2, 2)
	r2Model, _ := localProvider.GetEntityModel(r2, true)
	r2Model, _ = o1.Insert(r2Model)
	r2 = r2Model.Interface(false).(Reference)
	r3 := makeReference(ts, "r3", 3, 3)
	r3Model, _ := localProvider.GetEntityModel(r3, true)
	r3Model, _ = o1.Insert(r3Model)
	r3 = r3Model.Interface(false).(Reference)

	c1 := &Compose{
		Name:              "c1",
		Simple:            s1,
		SimplePtr:         &s1,
		SimpleArray:       []Simple{s1},
		SimplePtrArray:    []*Simple{&s1},
		Reference:         r1,
		ReferencePtr:      &r1,
		ReferenceArray:    []Reference{r1},
		ReferencePtrArray: []*Reference{&r1, &r2},
	}
	c1Model, _ := localProvider.GetEntityModel(c1, true)
	c1Model, err = o1.Insert(c1Model)
	if err != nil {
		t.Fatalf("insert compose failed: %v", err)
	}
	c1 = c1Model.Interface(true).(*Compose)

	// 部分替换：[r1,r2] -> [r2,r3]
	c1.ReferencePtrArray = []*Reference{&r2, &r3}
	c1Model, _ = localProvider.GetEntityModel(c1, true)
	_, err = o1.Update(c1Model)
	if err != nil {
		t.Fatalf("update compose failed: %v", err)
	}

	// r1 仍存在（仅解除链接，不删实体）
	qR1 := &Reference{ID: r1.ID}
	qR1Model, _ := localProvider.GetEntityModel(qR1, true)
	_, err = o1.Query(qR1Model)
	if err != nil {
		t.Fatalf("r1 should still exist: %v", err)
	}
	// 查询 Compose 应为 [r2, r3]
	qC := &Compose{ID: c1.ID}
	qCModel, _ := localProvider.GetEntityModel(qC, true)
	qCModel, err = o1.Query(qCModel)
	if err != nil {
		t.Fatalf("query compose failed: %v", err)
	}
	got := qCModel.Interface(true).(*Compose)
	if len(got.ReferencePtrArray) != 2 {
		t.Errorf("ReferencePtrArray len want 2, got %d", len(got.ReferencePtrArray))
	} else {
		ids := map[int]bool{got.ReferencePtrArray[0].ID: true, got.ReferencePtrArray[1].ID: true}
		if !ids[r2.ID] || !ids[r3.ID] {
			t.Errorf("ReferencePtrArray should be [r2,r3], got %v", got.ReferencePtrArray)
		}
	}
}

// setupUpdateRelationDiffEnv 创建表并插入 s1、r1、r2，供后续用例使用；owner 用于隔离不同测试
func setupUpdateRelationDiffEnv(t *testing.T, owner string) (o1 orm.Orm, localProvider provider.Provider, ts time.Time, s1 Simple, r1, r2 Reference) {
	t.Helper()
	orm.Initialize()
	localProvider = provider.NewLocalProvider(owner, nil)
	var err *cd.Error
	o1, err = orm.NewOrm(localProvider, config, "abc")
	if err != nil {
		t.Fatalf("new Orm failed: %v", err)
	}
	entityList := []any{&Simple{}, &Reference{}, &Compose{}}
	modelList, modelErr := registerLocalModel(localProvider, entityList)
	if modelErr != nil {
		t.Fatalf("register model failed: %v", modelErr)
	}
	if err = dropModel(o1, modelList); err != nil {
		t.Fatalf("drop model failed: %v", err)
	}
	if err = createModel(o1, modelList); err != nil {
		t.Fatalf("create model failed: %v", err)
	}
	ts, _ = time.ParseInLocation(util.CSTLayout, "2018-01-02 15:04:05", time.Local)
	s1 = Simple{I8: 12, I16: 23, I32: 34, I64: 45, Name: "test", Value: 12.345, F64: 23.456, TimeStamp: ts, Flag: true}
	s1Model, _ := localProvider.GetEntityModel(s1, true)
	s1Model, err = o1.Insert(s1Model)
	if err != nil {
		t.Fatalf("insert simple failed: %v", err)
	}
	s1 = s1Model.Interface(false).(Simple)
	r1 = makeReference(ts, "ref1", 1, 1)
	r1Model, _ := localProvider.GetEntityModel(r1, true)
	r1Model, err = o1.Insert(r1Model)
	if err != nil {
		t.Fatalf("insert r1 failed: %v", err)
	}
	r1 = r1Model.Interface(false).(Reference)
	r2 = makeReference(ts, "ref2", 2, 2)
	r2Model, _ := localProvider.GetEntityModel(r2, true)
	r2Model, err = o1.Insert(r2Model)
	if err != nil {
		t.Fatalf("insert r2 failed: %v", err)
	}
	r2 = r2Model.Interface(false).(Reference)
	return
}

// makeReference 构造满足表约束的 Reference（含非 nil 的 slice 字段）
func makeReference(ts time.Time, name string, fVal float32, f64 float64) Reference {
	return Reference{
		Name:        name,
		FValue:      fVal,
		F64:         f64,
		TimeStamp:   ts,
		Flag:        true,
		IArray:      []int{},
		FArray:      []float32{},
		StrArray:    []string{},
		BArray:      []bool{},
		StrPtrArray: []string{},
	}
}

// TestUpdateRelationR1ReferenceSingleAdd 场景 R1：引用单值新增（nil → Author1），关系表新增 1 行，实体不变
func TestUpdateRelationR1ReferenceSingleAdd(t *testing.T) {
	o1, localProvider, _, s1, r1, _ := setupUpdateRelationDiffEnv(t, "updateR1")
	defer o1.Release()
	defer orm.Uninitialized()

	// Insert Compose 无引用：ReferencePtr=nil, ReferencePtrArray=[]
	c1 := &Compose{
		Name:              "c1",
		Simple:            s1,
		SimplePtr:         &s1,
		SimpleArray:       []Simple{s1},
		SimplePtrArray:    []*Simple{&s1},
		Reference:         r1,
		ReferencePtr:      nil,
		ReferenceArray:    []Reference{r1},
		ReferencePtrArray: []*Reference{},
	}
	c1Model, _ := localProvider.GetEntityModel(c1, true)
	c1Model, err := o1.Insert(c1Model)
	if err != nil {
		t.Fatalf("insert compose failed: %v", err)
	}
	c1 = c1Model.Interface(true).(*Compose)

	// Update：引用单值新增 ReferencePtr=&r1, ReferencePtrArray=[r1]
	c1.ReferencePtr = &r1
	c1.ReferencePtrArray = []*Reference{&r1}
	c1Model, _ = localProvider.GetEntityModel(c1, true)
	_, err = o1.Update(c1Model)
	if err != nil {
		t.Fatalf("update compose failed: %v", err)
	}

	// 验证 r1 仍在库中；Compose 查出来关系为 r1
	qR1 := &Reference{ID: r1.ID}
	qR1Model, _ := localProvider.GetEntityModel(qR1, true)
	_, err = o1.Query(qR1Model)
	if err != nil {
		t.Fatalf("r1 should still exist: %v", err)
	}
	qC := &Compose{ID: c1.ID}
	qCModel, _ := localProvider.GetEntityModel(qC, true)
	qCModel, err = o1.Query(qCModel)
	if err != nil {
		t.Fatalf("query compose failed: %v", err)
	}
	got := qCModel.Interface(true).(*Compose)
	if len(got.ReferencePtrArray) != 1 || got.ReferencePtrArray[0].ID != r1.ID {
		t.Errorf("ReferencePtrArray should be [r1] after update, got %v", got.ReferencePtrArray)
	}
}

// TestUpdateRelationR3ReferenceSingleClear 场景 R3：引用单值清空（Author1 → nil），关系表删行，Author1 仍在库中
// 语义：slice 类型 nil=未赋值、[]=已赋值；故 ReferencePtr=nil 与 ReferencePtrArray=[] 均会触发 updateRelation 并清空对应链接。
func TestUpdateRelationR3ReferenceSingleClear(t *testing.T) {
	o1, localProvider, _, s1, r1, _ := setupUpdateRelationDiffEnv(t, "updateR3")
	defer o1.Release()
	defer orm.Uninitialized()

	c1 := &Compose{
		Name:              "c1",
		Simple:            s1,
		SimplePtr:         &s1,
		SimpleArray:       []Simple{s1},
		SimplePtrArray:    []*Simple{&s1},
		Reference:         r1,
		ReferencePtr:      &r1,
		ReferenceArray:    []Reference{r1},
		ReferencePtrArray: []*Reference{&r1},
	}
	c1Model, _ := localProvider.GetEntityModel(c1, true)
	c1Model, err := o1.Insert(c1Model)
	if err != nil {
		t.Fatalf("insert compose failed: %v", err)
	}
	c1 = c1Model.Interface(true).(*Compose)

	// Update：清空单值引用与切片引用（nil / [] 均视为已赋值，会执行关系表删行）
	c1.ReferencePtr = nil
	c1.ReferencePtrArray = []*Reference{}
	c1Model, _ = localProvider.GetEntityModel(c1, true)
	_, err = o1.Update(c1Model)
	if err != nil {
		t.Fatalf("update compose failed: %v", err)
	}

	// r1 仍存在（仅删链接不删实体）
	qR1 := &Reference{ID: r1.ID}
	qR1Model, _ := localProvider.GetEntityModel(qR1, true)
	_, err = o1.Query(qR1Model)
	if err != nil {
		t.Fatalf("r1 should still exist after clear: %v", err)
	}
	// 关系表已清空：Query 后 ReferencePtr 为 nil、ReferencePtrArray 长度为 0
	qC := &Compose{ID: c1.ID}
	qCModel, _ := localProvider.GetEntityModel(qC, true)
	_, err = o1.Query(qCModel)
	if err != nil {
		t.Fatalf("query compose failed: %v", err)
	}
	got := qCModel.Interface(true).(*Compose)
	if got.ReferencePtr != nil {
		t.Errorf("ReferencePtr should be nil after clear, got %v", got.ReferencePtr)
	}
	if len(got.ReferencePtrArray) != 0 {
		t.Errorf("ReferencePtrArray should be empty after clear, got len=%d", len(got.ReferencePtrArray))
	}
}

// TestUpdateRelationR4ReferenceSingleUnchanged 场景 R4：引用单值不变，Update 前后相同，结果仍正确
func TestUpdateRelationR4ReferenceSingleUnchanged(t *testing.T) {
	o1, localProvider, _, s1, r1, _ := setupUpdateRelationDiffEnv(t, "updateR4")
	defer o1.Release()
	defer orm.Uninitialized()

	c1 := &Compose{
		Name:              "c1",
		Simple:            s1,
		SimplePtr:         &s1,
		SimpleArray:       []Simple{s1},
		SimplePtrArray:    []*Simple{&s1},
		Reference:         r1,
		ReferencePtr:      &r1,
		ReferenceArray:    []Reference{r1},
		ReferencePtrArray: []*Reference{&r1},
	}
	c1Model, _ := localProvider.GetEntityModel(c1, true)
	c1Model, err := o1.Insert(c1Model)
	if err != nil {
		t.Fatalf("insert compose failed: %v", err)
	}
	c1 = c1Model.Interface(true).(*Compose)

	// Update：不变（仍为 ReferencePtr=&r1, ReferencePtrArray=[r1]）
	c1Model, _ = localProvider.GetEntityModel(c1, true)
	_, err = o1.Update(c1Model)
	if err != nil {
		t.Fatalf("update compose failed: %v", err)
	}

	qC := &Compose{ID: c1.ID}
	qCModel, _ := localProvider.GetEntityModel(qC, true)
	qCModel, err = o1.Query(qCModel)
	if err != nil {
		t.Fatalf("query compose failed: %v", err)
	}
	got := qCModel.Interface(true).(*Compose)
	if len(got.ReferencePtrArray) != 1 || got.ReferencePtrArray[0].ID != r1.ID {
		t.Errorf("ReferencePtrArray should still be [r1], got %v", got.ReferencePtrArray)
	}
}

// TestUpdateRelationR7ReferenceSliceClear 场景 R7：引用切片清空（[T1,T2] → []），关系表删 2 行，T1/T2 仍在库中
// 语义：[] 视为已赋值（size 0），会触发 updateRelation 并删除关系表中对应链接。
func TestUpdateRelationR7ReferenceSliceClear(t *testing.T) {
	o1, localProvider, _, s1, r1, r2 := setupUpdateRelationDiffEnv(t, "updateR7")
	defer o1.Release()
	defer orm.Uninitialized()

	c1 := &Compose{
		Name:              "c1",
		Simple:            s1,
		SimplePtr:         &s1,
		SimpleArray:       []Simple{s1},
		SimplePtrArray:    []*Simple{&s1},
		Reference:         r1,
		ReferencePtr:      &r1,
		ReferenceArray:    []Reference{r1},
		ReferencePtrArray: []*Reference{&r1, &r2},
	}
	c1Model, _ := localProvider.GetEntityModel(c1, true)
	c1Model, err := o1.Insert(c1Model)
	if err != nil {
		t.Fatalf("insert compose failed: %v", err)
	}
	c1 = c1Model.Interface(true).(*Compose)

	// Update：ReferencePtrArray=[]（已赋值、size 0），关系表删 2 行
	c1.ReferencePtrArray = []*Reference{}
	c1Model, _ = localProvider.GetEntityModel(c1, true)
	_, err = o1.Update(c1Model)
	if err != nil {
		t.Fatalf("update compose failed: %v", err)
	}

	// r1、r2 仍存在（引用关系不删实体）
	for _, ref := range []Reference{r1, r2} {
		qR := &Reference{ID: ref.ID}
		qRModel, _ := localProvider.GetEntityModel(qR, true)
		_, err = o1.Query(qRModel)
		if err != nil {
			t.Fatalf("reference %d should still exist: %v", ref.ID, err)
		}
	}
	// 关系表已清空：Query 后 ReferencePtrArray 长度为 0
	qC := &Compose{ID: c1.ID}
	qCModel, _ := localProvider.GetEntityModel(qC, true)
	_, err = o1.Query(qCModel)
	if err != nil {
		t.Fatalf("query compose failed: %v", err)
	}
	got := qCModel.Interface(true).(*Compose)
	if len(got.ReferencePtrArray) != 0 {
		t.Errorf("ReferencePtrArray should be empty after clear, got len=%d", len(got.ReferencePtrArray))
	}
}

// TestUpdateRelationR8ReferenceSliceUnchanged 场景 R8：引用切片完全相同，Update 后结果仍为 [T1,T2]
func TestUpdateRelationR8ReferenceSliceUnchanged(t *testing.T) {
	o1, localProvider, _, s1, r1, r2 := setupUpdateRelationDiffEnv(t, "updateR8")
	defer o1.Release()
	defer orm.Uninitialized()

	c1 := &Compose{
		Name:              "c1",
		Simple:            s1,
		SimplePtr:         &s1,
		SimpleArray:       []Simple{s1},
		SimplePtrArray:    []*Simple{&s1},
		Reference:         r1,
		ReferencePtr:      &r1,
		ReferenceArray:    []Reference{r1},
		ReferencePtrArray: []*Reference{&r1, &r2},
	}
	c1Model, _ := localProvider.GetEntityModel(c1, true)
	c1Model, err := o1.Insert(c1Model)
	if err != nil {
		t.Fatalf("insert compose failed: %v", err)
	}
	c1 = c1Model.Interface(true).(*Compose)

	// Update：完全相同 [r1,r2] → [r1,r2]
	c1Model, _ = localProvider.GetEntityModel(c1, true)
	_, err = o1.Update(c1Model)
	if err != nil {
		t.Fatalf("update compose failed: %v", err)
	}

	qC := &Compose{ID: c1.ID}
	qCModel, _ := localProvider.GetEntityModel(qC, true)
	qCModel, err = o1.Query(qCModel)
	if err != nil {
		t.Fatalf("query compose failed: %v", err)
	}
	got := qCModel.Interface(true).(*Compose)
	if len(got.ReferencePtrArray) != 2 {
		t.Errorf("ReferencePtrArray len want 2, got %d", len(got.ReferencePtrArray))
	} else {
		ids := map[int]bool{got.ReferencePtrArray[0].ID: true, got.ReferencePtrArray[1].ID: true}
		if !ids[r1.ID] || !ids[r2.ID] {
			t.Errorf("ReferencePtrArray should be [r1,r2], got %v", got.ReferencePtrArray)
		}
	}
}

// TestUpdateRelationC1ContainReplace 场景 C1：包含关系以新换旧，旧关联实体被删除，新关联实体被创建
func TestUpdateRelationC1ContainReplace(t *testing.T) {
	o1, localProvider, ts, s1, r1, _ := setupUpdateRelationDiffEnv(t, "updateC1")
	defer o1.Release()
	defer orm.Uninitialized()

	// 包含关系用“未插入”的 Reference 值（无 ID），Insert 时会创建该实体
	ref1Contain := makeReference(ts, "contain1", 10, 10)
	ref2Contain := makeReference(ts, "contain2", 20, 20)
	c1 := &Compose{
		Name:              "c1",
		Simple:            s1,
		SimplePtr:         &s1,
		SimpleArray:       []Simple{s1},
		SimplePtrArray:    []*Simple{&s1},
		Reference:         ref1Contain,
		ReferencePtr:      &r1,
		ReferenceArray:    []Reference{ref1Contain},
		ReferencePtrArray: []*Reference{&r1},
	}
	c1Model, _ := localProvider.GetEntityModel(c1, true)
	c1Model, err := o1.Insert(c1Model)
	if err != nil {
		t.Fatalf("insert compose failed: %v", err)
	}
	c1 = c1Model.Interface(true).(*Compose)
	// 包含关系下 Insert 会为 Reference 创建独立实体，取插入后 ID 作为“旧”关联实体
	oldRefID := c1.Reference.ID

	// Update：包含关系改为 ref2Contain（以新换旧）；旧 Reference 实体应被删除，新实体被创建
	c1.Reference = ref2Contain
	c1Model, _ = localProvider.GetEntityModel(c1, true)
	_, err = o1.Update(c1Model)
	if err != nil {
		t.Fatalf("update compose failed: %v", err)
	}
	c1 = c1Model.Interface(true).(*Compose)

	// 旧关联实体（Insert 时创建的 Reference，id=oldRefID）应被删除，Query 应 NotFound
	qOld := &Reference{ID: oldRefID}
	qOldModel, _ := localProvider.GetEntityModel(qOld, true)
	_, err = o1.Query(qOldModel)
	if err == nil {
		t.Errorf("old contained Reference (id=%d) should be deleted after contain update", oldRefID)
	}
	if err != nil && err.Code != cd.NotFound {
		t.Fatalf("query old ref: %v", err)
	}

	// 新 Compose.Reference 应为 contain2
	qC := &Compose{ID: c1.ID}
	qCModel, _ := localProvider.GetEntityModel(qC, true)
	qCModel, err = o1.Query(qCModel)
	if err != nil {
		t.Fatalf("query compose failed: %v", err)
	}
	got := qCModel.Interface(true).(*Compose)
	if got.Reference.Name != "contain2" {
		t.Errorf("compose Reference after contain update should be contain2, got Name %s", got.Reference.Name)
	}
}
