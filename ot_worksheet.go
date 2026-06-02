package main

import (
	"encoding/json"
	"fmt"
)

func init() {
	RegisterObjectType(ObjectTypeWorksheet, &objectWorksheet{})
}

type objectWorksheet struct {
	Id          string
	Name        string
	WorkspaceId string
	UpdatedDate string
	Stages      string
}

var _ ObjectInstance = &objectWorksheet{}

func (o *objectWorksheet) GetInfo() *ObjectInfo {
	return &ObjectInfo{
		Id:           o.Id,
		Name:         o.Name,
		Presentation: []string{o.Id, o.Name},
		Object:       o,
	}
}

func (o *objectWorksheet) GetValues() []PropertyInstance {
	props := ObjectTypeWorksheet.GetProperties()
	r := make([]PropertyInstance, len(props))
	for i, p := range props {
		r[i] = &propertyInstance{p, o}
	}
	return r
}

func (o *objectWorksheet) PrintToYaml(op Output, otyp ObjectType, obj ObjectInstance) error {
	return printToYamlFromObjectInstance(op, otyp, obj)
}

type objectTypeWorksheet struct{}

var ObjectTypeWorksheet ObjectType = &objectTypeWorksheet{}

var propertyDescWorksheet = []PropertyDesc{
	{"id", PropertyTypeString, false, true,
		func(o any) any { return o.(*objectWorksheet).Id },
		func(o any, v any) {
			if v != nil {
				o.(*objectWorksheet).Id = v.(string)
			}
		}},
	{"name", PropertyTypeString, false, false,
		func(o any) any { return o.(*objectWorksheet).Name },
		func(o any, v any) {
			if v != nil {
				o.(*objectWorksheet).Name = v.(string)
			}
		}},
	{"workspaceId", PropertyTypeString, false, false,
		func(o any) any { return o.(*objectWorksheet).WorkspaceId },
		func(o any, v any) {
			if v != nil {
				o.(*objectWorksheet).WorkspaceId = v.(string)
			}
		}},
	{"updatedDate", PropertyTypeString, true, false,
		func(o any) any { return o.(*objectWorksheet).UpdatedDate },
		func(o any, v any) {
			if v != nil {
				o.(*objectWorksheet).UpdatedDate = v.(string)
			}
		}},
	{"stages", PropertyTypeString, true, false,
		func(o any) any { return o.(*objectWorksheet).Stages },
		func(o any, v any) {
			if v != nil {
				o.(*objectWorksheet).Stages = v.(string)
			}
		}},
}

func (*objectTypeWorksheet) TypeName() string { return "worksheet" }
func (*objectTypeWorksheet) Help() string {
	return "A worksheet is an exploratory data analysis document within a workspace."
}
func (*objectTypeWorksheet) CanList() bool                   { return true }
func (*objectTypeWorksheet) CanGet() bool                    { return true }
func (*objectTypeWorksheet) CanCreate() bool                 { return true }
func (*objectTypeWorksheet) CanUpdate() bool                 { return false }
func (*objectTypeWorksheet) CanDelete() bool                 { return true }
func (*objectTypeWorksheet) GetPresentationLabels() []string { return []string{"id", "name"} }
func (*objectTypeWorksheet) GetProperties() []PropertyDesc   { return propertyDescWorksheet }

var gqlWorksheetSearch = compileGqlQuery(
`query Worksheet_Search($terms: DWSearchInput!, $maxCount: Int) {
		worksheetSearch(terms: $terms, maxCount: $maxCount) {
			results {
				worksheet { id name workspaceId updatedDate }
				score
			}
		}
	}`,
	"data", "worksheetSearch", "results",
)

func (ot *objectTypeWorksheet) List(cfg *Config, op Output, hc httpClient) ([]*ObjectInfo, error) {
	workspaceId := cfg.WorkspaceIdOrName
	termMap := object{"workspaceId": workspaceId}
	obj, err := gqlWorksheetSearch.query(cfg, op, hc, object{"terms": termMap})
	if err != nil || obj == nil {
		return nil, err
	}
	items, ok := obj.(array)
	if !ok {
		return nil, fmt.Errorf("worksheet list: unexpected response type")
	}
	var ret []*ObjectInfo
	for _, item := range items {
		m, ok := item.(object)
		if !ok {
			continue
		}
		ws, ok := m["worksheet"]
		if !ok {
			continue
		}
		wsObj, ok := ws.(object)
		if !ok {
			continue
		}
		o := &objectWorksheet{}
		if v, ok := wsObj["id"]; ok && v != nil {
			o.Id = v.(string)
		}
		if v, ok := wsObj["name"]; ok && v != nil {
			o.Name = v.(string)
		}
		if v, ok := wsObj["workspaceId"]; ok && v != nil {
			o.WorkspaceId = v.(string)
		}
		if v, ok := wsObj["updatedDate"]; ok && v != nil {
			o.UpdatedDate = v.(string)
		}
		ret = append(ret, o.GetInfo())
	}
	return ret, nil
}

var gqlGetWorksheet = compileGqlQuery(
`query Worksheet_Get($id: ObjectId!) {
		worksheet(id: $id) {
			id
			name
			workspaceId
			updatedDate
			stages { stageID pipeline }
		}
	}`,
	"data", "worksheet",
)

func (ot *objectTypeWorksheet) Get(cfg *Config, op Output, hc httpClient, id string) (ObjectInstance, error) {
	obj, err := gqlGetWorksheet.query(cfg, op, hc, object{"id": id})
	if err != nil {
		return nil, err
	}
	if obj == nil {
		return nil, nil
	}
	raw, ok := obj.(object)
	if !ok {
		return nil, fmt.Errorf("worksheet get: unexpected response type")
	}
	o := &objectWorksheet{}
	if v, ok := raw["id"]; ok && v != nil {
		o.Id = v.(string)
	}
	if v, ok := raw["name"]; ok && v != nil {
		o.Name = v.(string)
	}
	if v, ok := raw["workspaceId"]; ok && v != nil {
		o.WorkspaceId = v.(string)
	}
	if v, ok := raw["updatedDate"]; ok && v != nil {
		o.UpdatedDate = v.(string)
	}
	if v, ok := raw["stages"]; ok && v != nil {
		b, err := json.Marshal(v)
		if err != nil {
			return nil, fmt.Errorf("worksheet get: failed to marshal stages: %w", err)
		}
		o.Stages = string(b)
	}
	return o, nil
}

var gqlSaveWorksheet = compileGqlQuery(
`mutation Worksheet_Save($wks: WorksheetInput!) {
		saveWorksheet(wks: $wks) {
			id
			name
			workspaceId
		}
	}`,
	"data", "saveWorksheet",
)

// readOnlyWorksheetFields are fields returned by the Observe API that are not
// accepted as input by the saveWorksheet mutation.
var readOnlyWorksheetFields = []string{"updatedDate"}

func (ot *objectTypeWorksheet) Create(cfg *Config, op Output, hc httpClient, input object) (ObjectInstance, error) {
	for _, f := range readOnlyWorksheetFields {
		delete(input, f)
	}
	obj, err := gqlSaveWorksheet.query(cfg, op, hc, object{"wks": input})
	if err != nil {
		return nil, err
	}
	if obj == nil {
		return nil, fmt.Errorf("worksheet create: no result returned")
	}
	raw, ok := obj.(object)
	if !ok {
		return nil, fmt.Errorf("worksheet create: unexpected response type")
	}
	o := &objectWorksheet{}
	if v, ok := raw["id"]; ok && v != nil {
		o.Id = v.(string)
	}
	if v, ok := raw["name"]; ok && v != nil {
		o.Name = v.(string)
	}
	if v, ok := raw["workspaceId"]; ok && v != nil {
		o.WorkspaceId = v.(string)
	}
	return o, nil
}

func (ot *objectTypeWorksheet) Update(cfg *Config, op Output, hc httpClient, id string, input object) (ObjectInstance, error) {
	return nil, nil
}

var gqlDeleteWorksheet = compileGqlQuery(
`mutation Worksheet_Delete($id: ObjectId!) {
		deleteWorksheet(wks: $id) {
			success
			errorMessage
		}
	}`,
	"data", "deleteWorksheet",
)

func (ot *objectTypeWorksheet) Delete(cfg *Config, op Output, hc httpClient, id string) error {
	obj, err := gqlDeleteWorksheet.query(cfg, op, hc, object{"id": id})
	if err != nil {
		return err
	}
	if obj == nil {
		return nil
	}
	result, ok := obj.(object)
	if !ok {
		return fmt.Errorf("worksheet delete: unexpected response type")
	}
	if errMsg, ok := result["errorMessage"]; ok && errMsg != nil {
		if s, ok := errMsg.(string); ok && s != "" {
			return fmt.Errorf("worksheet delete: %s", s)
		}
	}
	return nil
}
