package structs

import (
	"fmt"
	"fx.prodigy9.co/contrib/reflection"
	"log"
	"reflect"
	"strings"
)

type ParsedStruct struct {
	source interface{}
	Fields []ParsedStructField
}

type ParsedStructField struct {
	Name   string
	Type   string
	Value  interface{}
	IsZero bool
	Tags   []ParsedStructTags
}

type ResourceField struct {
	Name       string
	ID         string
	OwnerID    string
	IsRequired bool
	DbTable    string
}

func (ps *ParsedStruct) FindFieldByTag(tagType string, tagName string, tagValue *string) *ParsedStructField {
	for _, field := range ps.Fields {
		tag := field.GetTag(tagType, tagName)
		if tag != nil && (tagValue == nil || tag.Value == *tagValue) {
			return &field
		}
	}
	return nil
}

func (ps *ParsedStruct) FindFieldsByTag(tagType string, tagName string) []ParsedStructField {
	fields := []ParsedStructField{}

	for _, field := range ps.Fields {
		tag := field.GetTag(tagType, tagName)
		if tag != nil {
			fields = append(fields, field)
		}
	}

	return fields
}

// GetResourceFields will pull all fields with tag "fx:resource" which represent database entities
func (ps *ParsedStruct) GetResourceFields() []ResourceField {
	fields := []ResourceField{}
	fieldMap := map[string]ResourceField{}

	for _, field := range ps.Fields {
		tag := field.GetTag("fx", "resource")
		ownerTag := field.GetTag("fx", "owner")
		if tag == nil && ownerTag == nil {
			continue
		}

		resource := ResourceField{
			Name:       field.Name,
			IsRequired: field.GetTag("fx", "required") != nil,
		}
		resource.ID = getFieldFromTagValue[string](tag, ps.source)
		if resource.ID != "" || tag == nil {
			resource.OwnerID = getFieldFromTagValue[string](ownerTag, ps.source)
		}
		resource.DbTable = reflection.CallMethod[string](field.Value, "GetTableName")

		fields = append(fields, resource)
		fieldMap[field.Name] = resource
	}

	return fields
}

func getFieldFromTagValue[T any](tag *ParsedStructTag, data interface{}) (result T) {
	if tag != nil {
		var err error
		result, err = reflection.GetField[T](data, tag.Value)
		if err != nil {
			log.Printf("error getting field %s", tag.Value)
		}
	}
	return
}

// GetTag returns a tag param value
//
// Example:  `fx:"query=id"` GetTag("fx", "query") // "id"
func (s *ParsedStructField) GetTag(tagType string, tagName string) *ParsedStructTag {
	for _, tags := range s.Tags {
		if tags.Type == tagType {
			for _, tag := range tags.Tags {
				if tag.Name == tagName {
					return &tag
				}
			}
		}
	}
	return nil
}

type ParsedStructTags struct {
	Type string
	Tags []ParsedStructTag
}

type ParsedStructTag struct {
	Name  string
	Value string
}

// Parse returns info about a struct's fields
func Parse(source interface{}) (parsed *ParsedStruct) {
	parsed = &ParsedStruct{source: source}
	typeEl := reflect.TypeOf(source).Elem()
	valueEl := reflect.ValueOf(source).Elem()

	// fields
	for i := 0; i < valueEl.NumField(); i++ {
		typeField := typeEl.Field(i)
		newField := ParsedStructField{
			Name:   typeField.Name,
			Type:   typeField.Type.String(),
			Value:  valueEl.Field(i).Interface(),
			IsZero: valueEl.Field(i).IsZero(),
			Tags:   []ParsedStructTags{},
		}
		// raw tags [in:"query=id;body=id2", fx:"where"]
		tagsStr := strings.Split(fmt.Sprintf("%v", typeField.Tag), " ")
		for _, tagStr := range tagsStr {
			tagType := strings.Split(tagStr, ":")[0]
			newTagType := ParsedStructTags{
				Type: tagType,
				Tags: []ParsedStructTag{},
			}
			// tag params [query=id; body=id2]
			tagParams := strings.Split(typeField.Tag.Get(tagType), ";")
			for _, tagParam := range tagParams {
				newTag := ParsedStructTag{}
				tagParts := strings.Split(tagParam, "=")
				if len(tagParts) >= 1 {
					newTag.Name = tagParts[0]
				}
				if len(tagParts) > 1 {
					newTag.Value = tagParts[1]
				}
				if newTag.Name != "" {
					newTagType.Tags = append(newTagType.Tags, newTag)
				}
			}
			newField.Tags = append(newField.Tags, newTagType)
		}
		parsed.Fields = append(parsed.Fields, newField)
	}
	return
}
