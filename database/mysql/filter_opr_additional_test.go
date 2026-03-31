package mysql

import (
	"reflect"
	"testing"
)

func TestInOprHandlesEmptyCollection(t *testing.T) {
	result := &ResultStack{}
	sql := InOpr("id", []any{}, result)
	if sql != "1=0" {
		t.Fatalf("InOpr(empty) sql mismatch, got %q", sql)
	}
	if len(result.Args()) != 0 {
		t.Fatalf("InOpr(empty) should not push args, got %#v", result.Args())
	}
}

func TestNotInOprHandlesEmptyCollection(t *testing.T) {
	result := &ResultStack{}
	sql := NotInOpr("id", []any{}, result)
	if sql != "1=1" {
		t.Fatalf("NotInOpr(empty) sql mismatch, got %q", sql)
	}
	if len(result.Args()) != 0 {
		t.Fatalf("NotInOpr(empty) should not push args, got %#v", result.Args())
	}
}

func TestInOprHandlesCollectionArgs(t *testing.T) {
	result := &ResultStack{}
	sql := InOpr("id", []any{int64(1), int64(2)}, result)
	if sql != "`id` in (?,?)" {
		t.Fatalf("InOpr(collection) sql mismatch, got %q", sql)
	}
	if !reflect.DeepEqual(result.Args(), []any{int64(1), int64(2)}) {
		t.Fatalf("InOpr(collection) args mismatch, got %#v", result.Args())
	}
}
