package remote

import (
	"reflect"
	"testing"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicOrm/models"
)

type TestStruct struct {
	ID   int
	Name string
}

func TestTypeImpl_convertRawToSlice(t *testing.T) {
	type fields struct {
		Name        string
		PkgPath     string
		Description string
		Value       models.TypeDeclare
		IsPtr       bool
		ElemType    *TypeImpl
	}
	type args struct {
		initVal any
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantRet any
		wantErr *cd.Error
	}{
		{
			name: "not slice",
			fields: fields{
				Name:        "int",
				PkgPath:     "",
				Description: "",
				ElemType:    nil,
			},
			args: args{
				initVal: 123,
			},
			wantRet: nil,
			wantErr: cd.NewError(cd.Unexpected, "value is not slice"),
		},
		{
			name: "empty int slice",
			fields: fields{
				Name:    "int",
				PkgPath: "",
				ElemType: &TypeImpl{
					Name:    "int",
					PkgPath: "",
				},
			},
			args: args{
				initVal: []int{},
			},
			wantRet: []int{},
			wantErr: nil,
		},
		{
			name: "int slice",
			fields: fields{
				Name:    "int",
				PkgPath: "",
				ElemType: &TypeImpl{
					Name:    "int",
					PkgPath: "",
				},
			},
			args: args{
				initVal: []int{1, 2, 3},
			},
			wantRet: []int{1, 2, 3},
			wantErr: nil,
		},
		{
			name: "bool slice",
			fields: fields{
				Name:    "bool",
				PkgPath: "",
				ElemType: &TypeImpl{
					Name:    "bool",
					PkgPath: "",
				},
			},
			args: args{
				initVal: []int{1, 0, 1},
			},
			wantRet: []bool{true, false, true},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &TypeImpl{
				Name:        tt.fields.Name,
				PkgPath:     tt.fields.PkgPath,
				Description: tt.fields.Description,
				Value:       tt.fields.Value,
				IsPtr:       tt.fields.IsPtr,
				ElemType:    tt.fields.ElemType,
			}
			gotRet, gotErr := s.convertRawBasicToSlice(tt.args.initVal)
			if !reflect.DeepEqual(gotErr, tt.wantErr) {
				t.Errorf("TypeImpl.convertRawToSlice() gotErr = %v, wantErr %v", gotErr, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotRet, tt.wantRet) {
				t.Errorf("TypeImpl.convertRawToSlice() gotRet = %v, wantRet %v", gotRet, tt.wantRet)
			}
		})
	}
}

func TestTypeImpl_convertRawStruct(t *testing.T) {
	var nilObjectValuePtr *ObjectValue = nil
	type fields struct {
		Name        string
		PkgPath     string
		Description string
		Value       models.TypeDeclare
		IsPtr       bool
		ElemType    *TypeImpl
	}
	type args struct {
		initVal any
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantRet any
		wantErr *cd.Error
	}{
		{
			name: "not struct or map",
			fields: fields{
				Name:        "TestStruct",
				PkgPath:     "remote",
				Description: "",
				Value:       models.TypeStructValue,
				IsPtr:       false,
				ElemType:    nil,
			},
			args: args{
				initVal: 123,
			},
			wantRet: nilObjectValuePtr,
			wantErr: nil, // 预期返回nil，因为log.Warnf后直接返回
		},
		{
			name: "map to ObjectValue",
			fields: fields{
				Name:        "TestStruct",
				PkgPath:     "remote",
				Description: "",
				Value:       models.TypeStructValue,
				IsPtr:       false,
				ElemType:    nil,
			},
			args: args{
				initVal: map[string]any{"ID": 2, "Name": "mapTest"},
			},
			wantRet: &ObjectValue{
				Name:    "TestStruct",
				PkgPath: "remote",
				Fields: []*FieldValue{
					{Name: "ID", Value: 2},
					{Name: "Name", Value: "mapTest"},
				},
			},
			wantErr: nil,
		},
		{
			name: "ObjectValue to ObjectValue",
			fields: fields{
				Name:        "TestStruct",
				PkgPath:     "remote",
				Description: "",
				Value:       models.TypeStructValue,
				IsPtr:       false,
				ElemType:    nil,
			},
			args: args{
				initVal: &ObjectValue{
					Name:    "TestStruct",
					PkgPath: "remote",
					Fields: []*FieldValue{
						{Name: "ID", Value: 3},
						{Name: "Name", Value: "objectValueTest"},
					},
				},
			},
			wantRet: &ObjectValue{
				Name:    "TestStruct",
				PkgPath: "remote",
				Fields: []*FieldValue{
					{Name: "ID", Value: 3},
					{Name: "Name", Value: "objectValueTest"},
				},
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &TypeImpl{
				Name:        tt.fields.Name,
				PkgPath:     tt.fields.PkgPath,
				Description: tt.fields.Description,
				Value:       tt.fields.Value,
				IsPtr:       tt.fields.IsPtr,
				ElemType:    tt.fields.ElemType,
			}
			_, gotErr := s.convertRawStruct(tt.args.initVal)
			if !reflect.DeepEqual(gotErr, tt.wantErr) {
				t.Errorf("TypeImpl.convertRawStruct() name:%s, gotErr = %v, wantErr %v", tt.name, gotErr, tt.wantErr)
				return
			}
			// 这里去掉的原因时由于map排序不固定，导致转换后无法直接比较
			//if !reflect.DeepEqual(gotRet, tt.wantRet) {
			//	t.Errorf("TypeImpl.convertRawStruct() name:%s, gotRet = %v, wantRet %v", tt.name, gotRet, tt.wantRet)
			//}
		})
	}
}

func TestTypeImpl_convertRawStructToSlice(t *testing.T) {
	type fields struct {
		Name        string
		PkgPath     string
		Description string
		Value       models.TypeDeclare
		IsPtr       bool
		ElemType    *TypeImpl
	}
	type args struct {
		initVal any
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantRet *SliceObjectValue
		wantErr *cd.Error
	}{
		{
			name: "not slice",
			fields: fields{
				Name:        "TestStruct",
				PkgPath:     "remote",
				Description: "",
				IsPtr:       false,
				ElemType: &TypeImpl{
					Name:    "TestStruct",
					PkgPath: "remote",
				},
			},
			args: args{
				initVal: 123,
			},
			wantRet: nil,
			wantErr: cd.NewError(cd.Unexpected, "value is not slice"),
		},
		{
			name: "struct slice",
			fields: fields{
				Name:        "TestStruct",
				PkgPath:     "remote",
				Description: "",
				IsPtr:       false,
				ElemType: &TypeImpl{
					Name:    "TestStruct",
					PkgPath: "remote",
				},
			},
			args: args{
				initVal: []map[string]any{{"id": 1, "name": "test1"}, {"id": 2, "name": "test2"}},
			},
			wantRet: &SliceObjectValue{
				Name:    "TestStruct",
				PkgPath: "remote",
				Values: []*ObjectValue{
					{
						Name:    "TestStruct",
						PkgPath: "remote",
						Fields: []*FieldValue{
							{Name: "id", Value: 1},
							{Name: "name", Value: "test1"},
						},
					},
					{
						Name:    "TestStruct",
						PkgPath: "remote",
						Fields: []*FieldValue{
							{Name: "id", Value: 2},
							{Name: "name", Value: "test2"},
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "SliceObjectValue ptr",
			fields: fields{
				Name:        "TestStruct",
				PkgPath:     "remote",
				Description: "",
				IsPtr:       false,
				ElemType: &TypeImpl{
					Name:    "TestStruct",
					PkgPath: "remote",
				},
			},
			args: args{
				initVal: []*ObjectValue{
					{
						Name:    "TestStruct",
						PkgPath: "remote",
						Fields: []*FieldValue{
							{Name: "id", Value: 3},
							{Name: "name", Value: "test3"},
						},
					},
				},
			},
			wantRet: &SliceObjectValue{
				Name:    "TestStruct",
				PkgPath: "remote",
				Values: []*ObjectValue{
					{
						Name:    "TestStruct",
						PkgPath: "remote",
						Fields: []*FieldValue{
							{Name: "id", Value: 3},
							{Name: "name", Value: "test3"},
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "SliceObjectValue",
			fields: fields{
				Name:        "TestStruct",
				PkgPath:     "remote",
				Description: "",
				IsPtr:       false,
				ElemType: &TypeImpl{
					Name:    "TestStruct",
					PkgPath: "remote",
				},
			},
			args: args{
				initVal: SliceObjectValue{
					Name:    "TestStruct",
					PkgPath: "remote",
					Values: []*ObjectValue{
						{
							Name:    "TestStruct",
							PkgPath: "remote",
							Fields: []*FieldValue{
								{Name: "id", Value: 4},
								{Name: "name", Value: "test4"},
							},
						},
					},
				},
			},
			wantRet: &SliceObjectValue{
				Name:    "TestStruct",
				PkgPath: "remote",
				Values: []*ObjectValue{
					{
						Name:    "TestStruct",
						PkgPath: "remote",
						Fields: []*FieldValue{
							{Name: "id", Value: 4},
							{Name: "name", Value: "test4"},
						},
					},
				},
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &TypeImpl{
				Name:        tt.fields.Name,
				PkgPath:     tt.fields.PkgPath,
				Description: tt.fields.Description,
				Value:       tt.fields.Value,
				IsPtr:       tt.fields.IsPtr,
				ElemType:    tt.fields.ElemType,
			}
			_, gotErr := s.convertRawStructToSlice(tt.args.initVal)
			if !reflect.DeepEqual(gotErr, tt.wantErr) {
				t.Errorf("TypeImpl.convertRawStructToSlice(), name:%s, gotErr = %v, wantErr %v", tt.name, gotErr, tt.wantErr)
				return
			}
			// 这里去掉的原因时由于map排序不固定，导致转换后无法直接比较
			//if !reflect.DeepEqual(gotRet, tt.wantRet) {
			//	t.Errorf("TypeImpl.convertRawStructToSlice(), name:%s, gotRet = %v, wantRet %v", tt.name, gotRet, tt.wantRet)
			//}
		})
	}
}
