package local

import (
	"reflect"
	"testing"
	"time"

	"github.com/muidea/magicOrm/model"
	"github.com/stretchr/testify/assert"
)

type testStruct struct {
	Field1 int
	Field2 string
}

func TestNewType(t *testing.T) {
	tests := []struct {
		name    string
		input   reflect.Type
		wantErr bool
	}{
		{
			name:    "int type",
			input:   reflect.TypeOf(0),
			wantErr: false,
		},
		{
			name:    "string type",
			input:   reflect.TypeOf(""),
			wantErr: false,
		},
		{
			name:    "struct type",
			input:   reflect.TypeOf(testStruct{}),
			wantErr: false,
		},
		{
			name:    "pointer type",
			input:   reflect.TypeOf(&testStruct{}),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewType(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.Nil(t, err)
			assert.NotNil(t, got)
		})
	}
}

func TestTypeImpl_GetName(t *testing.T) {
	tests := []struct {
		name  string
		input reflect.Type
		want  string
	}{
		{
			name:  "int type",
			input: reflect.TypeOf(0),
			want:  "int",
		},
		{
			name:  "struct type",
			input: reflect.TypeOf(testStruct{}),
			want:  "testStruct",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			typ, _ := NewType(tt.input)
			assert.Equal(t, tt.want, typ.GetName())
		})
	}
}

func TestTypeImpl_GetPkgPath(t *testing.T) {
	typ, _ := NewType(reflect.TypeOf(testStruct{}))
	assert.Contains(t, typ.GetPkgPath(), "testStruct")
}

func TestTypeImpl_GetValue(t *testing.T) {
	tests := []struct {
		name  string
		input reflect.Type
		want  model.TypeDeclare
	}{
		{
			name:  "int type",
			input: reflect.TypeOf(0),
			want:  model.TypeIntegerValue,
		},
		{
			name:  "string type",
			input: reflect.TypeOf(""),
			want:  model.TypeStringValue,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			typ, _ := NewType(tt.input)
			assert.Equal(t, tt.want, typ.GetValue())
		})
	}
}

func TestTypeImpl_Interface(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name    string
		typ     reflect.Type
		initVal any
		wantErr bool
	}{
		{
			name:    "int value",
			typ:     reflect.TypeOf(0),
			initVal: 123,
			wantErr: false,
		},
		{
			name:    "string value",
			typ:     reflect.TypeOf(""),
			initVal: "test",
			wantErr: false,
		},
		{
			name:    "time value",
			typ:     reflect.TypeOf(time.Time{}),
			initVal: now,
			wantErr: false,
		},
		{
			name:    "invalid value",
			typ:     reflect.TypeOf(0),
			initVal: "invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			typ, _ := NewType(tt.typ)
			_, err := typ.Interface(tt.initVal)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.Nil(t, err)
		})
	}
}

func TestTypeImpl_Elem(t *testing.T) {
	typ, _ := NewType(reflect.TypeOf([]int{}))
	elem := typ.Elem()
	assert.Equal(t, model.TypeIntegerValue, elem.GetValue())
}

func TestTypeImpl_IsBasic(t *testing.T) {
	tests := []struct {
		name  string
		input reflect.Type
		want  bool
	}{
		{
			name:  "basic type",
			input: reflect.TypeOf(0),
			want:  true,
		},
		{
			name:  "struct type",
			input: reflect.TypeOf(testStruct{}),
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			typ, _ := NewType(tt.input)
			assert.Equal(t, tt.want, typ.IsBasic())
		})
	}
}

func TestTypeImpl_IsStruct(t *testing.T) {
	tests := []struct {
		name  string
		input reflect.Type
		want  bool
	}{
		{
			name:  "basic type",
			input: reflect.TypeOf(0),
			want:  false,
		},
		{
			name:  "struct type",
			input: reflect.TypeOf(testStruct{}),
			want:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			typ, _ := NewType(tt.input)
			assert.Equal(t, tt.want, typ.IsStruct())
		})
	}
}

func TestTypeImpl_IsSlice(t *testing.T) {
	tests := []struct {
		name  string
		input reflect.Type
		want  bool
	}{
		{
			name:  "slice type",
			input: reflect.TypeOf([]int{}),
			want:  true,
		},
		{
			name:  "non-slice type",
			input: reflect.TypeOf(0),
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			typ, _ := NewType(tt.input)
			assert.Equal(t, tt.want, typ.IsSlice())
		})
	}
}

func TestTypeImpl_Interface_CanSet(t *testing.T) {
	iVal := 100
	intTypePtr, intTypeErr := NewType(reflect.TypeOf(iVal))
	assert.Nil(t, intTypeErr)
	assert.Equal(t, model.TypeIntegerValue, intTypePtr.GetValue())
	valPtr, valErr := intTypePtr.Interface(nil)
	assert.Nil(t, valErr)
	assert.Equal(t, true, valPtr.IsValid())
	assert.Equal(t, true, valPtr.IsZero())
	valPtr.Set(iVal)
	assert.Equal(t, false, valPtr.IsZero())
	assert.Equal(t, iVal, valPtr.Get())
}
