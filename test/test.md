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
