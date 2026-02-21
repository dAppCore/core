package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormField_Good_Types(t *testing.T) {
	fields := []FormField{
		{Label: "Name", Key: "name", Type: FieldText},
		{Label: "Password", Key: "pass", Type: FieldPassword},
		{Label: "Accept", Key: "ok", Type: FieldConfirm},
	}
	assert.Equal(t, 3, len(fields))
	assert.Equal(t, FieldText, fields[0].Type)
	assert.Equal(t, FieldPassword, fields[1].Type)
	assert.Equal(t, FieldConfirm, fields[2].Type)
}

func TestFieldType_Good_Constants(t *testing.T) {
	assert.Equal(t, FieldType("text"), FieldText)
	assert.Equal(t, FieldType("password"), FieldPassword)
	assert.Equal(t, FieldType("confirm"), FieldConfirm)
	assert.Equal(t, FieldType("select"), FieldSelect)
}

func TestTabItem_Good_Structure(t *testing.T) {
	tabs := []TabItem{
		{Title: "Overview", Content: "overview content"},
		{Title: "Details", Content: "detail content"},
	}
	assert.Equal(t, 2, len(tabs))
	assert.Equal(t, "Overview", tabs[0].Title)
}
