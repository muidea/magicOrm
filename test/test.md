# 测试对象

## 简单对象

```
type Simple struct {
	ID        int       `orm:"id key auto"`
	I8        int8      `orm:"i8"`
	I16       int16     `orm:"i16"`
	I32       int32     `orm:"i32"`
	I64       uint64    `orm:"i64"`
	Name      string    `orm:"name"`
	Value     float32   `orm:"value"`
	F64       float64   `orm:"f64"`
	TimeStamp time.Time `orm:"ts"`
	Flag      bool      `orm:"flag"`
}

``` 

## 包含&引用

```
type Reference struct {
	ID          int        `orm:"id key auto"`
	Name        string     `orm:"name"`
	FValue      *float32   `orm:"value"`
	F64         float64    `orm:"f64"`
	TimeStamp   *time.Time `orm:"ts"`
	Flag        *bool      `orm:"flag"`
	IArray      []int      `orm:"iArray"`
	FArray      []float32  `orm:"fArray"`
	StrArray    []string   `orm:"strArray"`
	BArray      []bool     `orm:"bArray"`
	PtrArray    *[]string  `orm:"ptrArray"`
	StrPtrArray []*string  `orm:"strPtrArray"`
	PtrStrArray *[]*string `orm:"ptrStrArray"`
}

```

## 复合对象

```
type Compose struct {
	ID             int           `orm:"id key auto"`
	Name           string        `orm:"name"`
	Simple         Simple        `orm:"simple"`
	PtrSimple      *Simple       `orm:"ptrSimple"`
	SimpleArray    []Simple      `orm:"simpleArray"`
	SimplePtrArray []*Simple     `orm:"simplePtrArray"`
	PtrSimpleArray *[]Simple     `orm:"ptrSimpleArray"`
	Reference      Reference     `orm:"reference"`
	PtrReference   *Reference    `orm:"ptrReference"`
	RefArray       []Reference   `orm:"refArray"`
	RefPtrArray    []*Reference  `orm:"refPtrArray"`
	PtrRefArray    *[]*Reference `orm:"ptrRefArray"`
	PtrCompose     *Compose      `orm:"ptrCompose"`
}

```

# 测试用例

## ORM初始化
```go

localProvider:
orm.Initialize(50, "root", "rootkit", "localhost:3306", "testdb", true)
defer orm.Uninitialize()

o1, err := orm.NewOrm(localOwner,"abc")
defer o1.Release()
if err != nil {
    t.Errorf("new Orm failed, err:%s", err.Error())
    return
}

provider := orm.GetProvider(localOwner,"abc")

remoteProvider:
orm.Initialize(50, "root", "rootkit", "localhost:3306", "testdb", false)
defer orm.Uninitialize()

o1, err := orm.NewOrm(localOwner,"abc")
defer o1.Release()
if err != nil {
    t.Errorf("new Orm failed, err:%s", err.Error())
    return
}

provider := orm.GetProvider(localOwner,"abc")

```

## 模型定义

```go

localProvider:
simpleDef := &Simple{}
referenceDef := &Reference{}
composeDef := &Compose{}

remoteProvider:
simpleDef := remote.GetObject(&Simple{})
referenceDef := remote.GetObject(&Reference{})
composeDef := remote.GetObject(&Compose{})

```

## 模型初始化
```go

entityList := []interface{}{simpleDef, referenceDef, composeDef}
modelList, modelErr := registerModel(provider, entityList)
if modelErr != nil {
    err = modelErr
    t.Errorf("register model failed. err:%s", err.Error())
    return
}

err = dropModel(o1, modelList)
if err != nil {
    t.Errorf("drop model failed. err:%s", err.Error())
    return
}

err = createModel(o1, modelList)
if err != nil {
    t.Errorf("create model failed. err:%s", err.Error())
    return
}

```

## 简单对象CURD

```go
ts, _ := time.Parse(util.CSTLayout, "2018-01-02 15:04:05")
sVal := Simple{I8: 12, I16: 23, I32: 34, I64: 45, Name: "test code", Value: 12.345, F64: 23.456, TimeStamp: ts, Flag: true}
sValList := []*Simple{}
sModelList := []model.Model{}

// insert
for idx:=0; idx<100; idx++ {
    sVal.I32 = int32(idx)
    sValList = append(sValList, &sVal)
    
    sModel,sErr := provider.GetEntityModel(&sVal)
    if sErr != nil {
        err = sErr
        t.Errorf("GetEntityModel failed. err:%s", err.Error())
        return
    }
    
    sModelList = append(sModelList, sModel)
}

for idx:=0; idx<100; idx++ {
    vModel,vErr := o1.Insert(sModelList[idx])
    if vErr != nil {
        err = vErr
        t.Errorf("Insert failed. err:%s", err.Error())
        return
    }
    
    sModelList[idx] = vModel
    sValList[idx] = vModel.Interface(true).(*Simple)
}

// update
for idx:=0; idx<100; idx++ {
    sVal := sValList[idx]
    sVal.Name = "hi"
    sModel,sErr := provider.GetEntityModel(sVal)
    if sErr != nil {
    err = sErr
        t.Errorf("GetEntityModel failed. err:%s", err.Error())
        return
    }
    
    sModelList[idx] = sModel
}
for idx:=0; idx<100; idx++ {
    vModel,vErr := o1.Update(sModelList[idx])
    if vErr != nil {
        err = vErr
        t.Errorf("Update failed. err:%s", err.Error())
        return
    }
    
    sModelList[idx] = vModel
    sValList[idx] = vModel.Interface(true).(*Simple)
}

// query
qValList := []*Simple{}
qModelList := []model.Model{}
for idx:=0; idx<100; idx++ {
    qVal := &Simple{ID: sValList[idx].ID}
    qValList = append(qValList, qVal)
    
    qModel,qErr := provider.GetEntityModel(qVal)
    if qErr != nil {
        err = qErr
        t.Errorf("GetEntityModel failed. err:%s", err.Error())
        return
    }
    
    qModelList = append(qModelList, qModel)
}

for idx:=0; idx<100; idx++ {
    qModel,qErr := o1.Query(qModelList[idx])
    if qErr != nil {
        err = qErr
        t.Errorf("Query failed. err:%s", err.Error())
        return
    }
    
    qModelList[idx] = qModel
    qValList[idx] = qModel.Interface(true).(*Simple)
}

for idx:=0; idx<100; idx++ {
    sVal := sValList[idx]
    qVal := qValList[idx]
    if !sVal.IsSame(qVal) {
        err = fmt.Errorf("compare value failed")
        t.Errorf("IsSame failed. err:%s", err.Error())
        return
    }
}

bqValList := []*Simple{}
bqModel,bqErr := provider.GetEntityModel(&bqValList)
if bqErr != nil {
    t.Errorf("GetEntityModel failed, err:%s", bqErr.Error())
    return
}

filter := orm.GetFilter(localOwner,"abc")
filter.Equal("Name", "hi")
filter.ValueMask(&Simple{})
bqModelList, bqModelErr := o1.BatchQuery(bqModel, filter)
if bqModelErr != nil {
    t.Errorf("BatchQuery failed, err:%s", bqModelErr.Error())
    return
}
if len(bqModelList) != 100 {
    t.Errorf("batch query simple failed")
    return
}

// delete
for idx:=0; idx<100; idx++ {
    _,qErr := o1.Delete(bqModelList[idx])
    if qErr != nil {
        err = qErr
        return
    }
}

```

## 包含&引用对象CURD

## 非必填字段CURD