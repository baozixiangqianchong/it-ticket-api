package service

import (
	"testing"

	"it-ticket-api/internal/model"
)

func TestNormalizeFieldsRequiresLabel(t *testing.T) {
	if _, err := normalizeFields(nil); err == nil {
		t.Fatal("空字段应失败")
	}
	got, err := normalizeFields([]model.TemplateField{{Label: "型号", Placeholder: "HP", Required: true}})
	if err != nil {
		t.Fatal(err)
	}
	if got[0].Key != "field_1" || !got[0].Required {
		t.Fatalf("got %+v", got[0])
	}
}

func TestValidFieldKey(t *testing.T) {
	if !validFieldKey("model") || !validFieldKey("field_1") {
		t.Fatal("合法 key 应通过")
	}
	if validFieldKey("1abc") || validFieldKey("A") || validFieldKey("") {
		t.Fatal("非法 key 应拒绝")
	}
}
