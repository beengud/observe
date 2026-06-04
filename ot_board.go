package main

import (
	"encoding/json"
	"fmt"
)

func init() {
	RegisterObjectType(ObjectTypeBoard, &objectBoard{})
}

type objectBoard struct {
	Id          string
	Name        string
	WorkspaceId string
	Description string
	FolderId    string
	Visibility  string
	UpdatedDate string
	Layout      string
	Stages      string
}

var _ ObjectInstance = &objectBoard{}

func (o *objectBoard) GetInfo() *ObjectInfo {
	return &ObjectInfo{
		Id:           o.Id,
		Name:         o.Name,
		Presentation: []string{o.Id, o.Name},
		Object:       o,
	}
}

func (o *objectBoard) GetValues() []PropertyInstance {
	props := ObjectTypeBoard.GetProperties()
	r := make([]PropertyInstance, len(props))
	for i, p := range props {
		r[i] = &propertyInstance{p, o}
	}
	return r
}

func (o *objectBoard) PrintToYaml(op Output, otyp ObjectType, obj ObjectInstance) error {
	return printToYamlFromObjectInstance(op, otyp, obj)
}

type objectTypeBoard struct{}

var ObjectTypeBoard ObjectType = &objectTypeBoard{}

var propertyDescBoard = []PropertyDesc{
	{"id", PropertyTypeString, false, true,
		func(o any) any { return o.(*objectBoard).Id },
		func(o any, v any) {
			if v != nil {
				o.(*objectBoard).Id = v.(string)
			}
		}},
	{"name", PropertyTypeString, false, false,
		func(o any) any { return o.(*objectBoard).Name },
		func(o any, v any) {
			if v != nil {
				o.(*objectBoard).Name = v.(string)
			}
		}},
	{"workspaceId", PropertyTypeString, false, false,
		func(o any) any { return o.(*objectBoard).WorkspaceId },
		func(o any, v any) {
			if v != nil {
				o.(*objectBoard).WorkspaceId = v.(string)
			}
		}},
	{"description", PropertyTypeString, false, false,
		func(o any) any { return o.(*objectBoard).Description },
		func(o any, v any) {
			if v != nil {
				o.(*objectBoard).Description = v.(string)
			}
		}},
	{"folderId", PropertyTypeString, false, false,
		func(o any) any { return o.(*objectBoard).FolderId },
		func(o any, v any) {
			if v != nil {
				o.(*objectBoard).FolderId = v.(string)
			}
		}},
	{"visibility", PropertyTypeString, false, false,
		func(o any) any { return o.(*objectBoard).Visibility },
		func(o any, v any) {
			if v != nil {
				o.(*objectBoard).Visibility = v.(string)
			}
		}},
	{"updatedDate", PropertyTypeString, true, false,
		func(o any) any { return o.(*objectBoard).UpdatedDate },
		func(o any, v any) {
			if v != nil {
				o.(*objectBoard).UpdatedDate = v.(string)
			}
		}},
	{"layout", PropertyTypeString, true, false,
		func(o any) any { return o.(*objectBoard).Layout },
		func(o any, v any) {
			if v != nil {
				o.(*objectBoard).Layout = v.(string)
			}
		}},
	{"stages", PropertyTypeString, true, false,
		func(o any) any { return o.(*objectBoard).Stages },
		func(o any, v any) {
			if v != nil {
				o.(*objectBoard).Stages = v.(string)
			}
		}},
}

func (*objectTypeBoard) TypeName() string { return "board" }
func (*objectTypeBoard) Help() string {
	return "A board (dashboard) organizes visualizations of data within a workspace."
}
func (*objectTypeBoard) CanList() bool                   { return true }
func (*objectTypeBoard) CanGet() bool                    { return true }
func (*objectTypeBoard) CanCreate() bool                 { return false }
func (*objectTypeBoard) CanUpdate() bool                 { return false }
func (*objectTypeBoard) CanDelete() bool                 { return true }
func (*objectTypeBoard) GetPresentationLabels() []string { return []string{"id", "name"} }
func (*objectTypeBoard) GetProperties() []PropertyDesc   { return propertyDescBoard }

var gqlListBoard = compileGqlQuery(
`query Board_List($workspaceId: [ObjectId!]!) {
		dashboardSearch(terms: { workspaceId: $workspaceId }) {
			dashboards {
				score
				dashboard { id name workspaceId updatedDate }
			}
		}
	}`,
	"data", "dashboardSearch", "dashboards",
)

// unpackBoardItems extracts []*ObjectInfo from an array of dashboardSearch result items.
// Each item is expected to have a "dashboard" sub-object.
func unpackBoardItems(items array) []*ObjectInfo {
	var ret []*ObjectInfo
	for _, item := range items {
		m, ok := item.(object)
		if !ok {
			continue
		}
		dash, ok := m["dashboard"]
		if !ok {
			continue
		}
		dashObj, ok := dash.(object)
		if !ok {
			continue
		}
		o := &objectBoard{}
		if v, ok := dashObj["id"]; ok && v != nil {
			o.Id = v.(string)
		}
		if v, ok := dashObj["name"]; ok && v != nil {
			o.Name = v.(string)
		}
		if v, ok := dashObj["workspaceId"]; ok && v != nil {
			o.WorkspaceId = v.(string)
		}
		if v, ok := dashObj["updatedDate"]; ok && v != nil {
			o.UpdatedDate = v.(string)
		}
		ret = append(ret, o.GetInfo())
	}
	return ret
}

func (ot *objectTypeBoard) List(cfg *Config, op Output, hc httpClient) ([]*ObjectInfo, error) {
	// If any search flags are set, delegate to Search with those terms.
	// The global --workspace flag populates cfg.WorkspaceIdOrName for workspace filtering.
	if flagBoardScaffold != "" || flagBoardSearchFolder != "" {
		terms := BoardSearchTerms{
			Name:        flagBoardScaffold,
			WorkspaceId: cfg.WorkspaceIdOrName,
			FolderId:    flagBoardSearchFolder,
		}
		return ot.Search(cfg, op, hc, terms)
	}
	workspaceId := cfg.WorkspaceIdOrName
	obj, err := gqlListBoard.query(cfg, op, hc, object{"workspaceId": []string{workspaceId}})
	if err != nil || obj == nil {
		return nil, err
	}
	items, ok := obj.(array)
	if !ok {
		return nil, fmt.Errorf("board list: unexpected response type")
	}
	return unpackBoardItems(items), nil
}

// BoardSearchTerms holds optional search parameters for dashboardSearch.
type BoardSearchTerms struct {
	Name        string
	WorkspaceId string
	FolderId    string
}

var gqlSearchBoard = compileGqlQuery(
`query Board_Search($terms: DWSearchInput!, $maxCount: Int) {
		dashboardSearch(terms: $terms, maxCount: $maxCount) {
			results {
				dashboard { id name workspaceId updatedDate }
				score
			}
		}
	}`,
	"data", "dashboardSearch", "results",
)

// Search queries dashboardSearch using optional name/workspace/folder filters.
// Results are presented the same way as List.
func (ot *objectTypeBoard) Search(cfg *Config, op Output, hc httpClient, terms BoardSearchTerms) ([]*ObjectInfo, error) {
	termMap := object{}
	if terms.Name != "" {
		termMap["name"] = terms.Name
	}
	if terms.WorkspaceId != "" {
		termMap["workspaceId"] = terms.WorkspaceId
	}
	if terms.FolderId != "" {
		termMap["folderId"] = terms.FolderId
	}
	obj, err := gqlSearchBoard.query(cfg, op, hc, object{"terms": termMap})
	if err != nil || obj == nil {
		return nil, err
	}
	items, ok := obj.(array)
	if !ok {
		return nil, fmt.Errorf("board search: unexpected response type")
	}
	return unpackBoardItems(items), nil
}

var gqlGetBoard = compileGqlQuery(
`query Board_Get($id: ObjectId!) {
		dashboard(id: $id) {
			id
			name
			workspaceId
			description
			folderId
			visibility
			updatedDate
			layout
			stages { id stageID pipeline input { inputName datasetId stageId } }
		}
	}`,
	"data", "dashboard",
)

func (ot *objectTypeBoard) Get(cfg *Config, op Output, hc httpClient, id string) (ObjectInstance, error) {
	obj, err := gqlGetBoard.query(cfg, op, hc, object{"id": id})
	if err != nil {
		return nil, err
	}
	if obj == nil {
		return nil, nil
	}
	raw, ok := obj.(object)
	if !ok {
		return nil, fmt.Errorf("board get: unexpected response type")
	}
	o := &objectBoard{}
	if v, ok := raw["id"]; ok && v != nil {
		o.Id = v.(string)
	}
	if v, ok := raw["name"]; ok && v != nil {
		o.Name = v.(string)
	}
	if v, ok := raw["workspaceId"]; ok && v != nil {
		o.WorkspaceId = v.(string)
	}
	if v, ok := raw["description"]; ok && v != nil {
		o.Description = v.(string)
	}
	if v, ok := raw["folderId"]; ok && v != nil {
		o.FolderId = v.(string)
	}
	if v, ok := raw["visibility"]; ok && v != nil {
		o.Visibility = v.(string)
	}
	if v, ok := raw["updatedDate"]; ok && v != nil {
		o.UpdatedDate = v.(string)
	}
	if v, ok := raw["layout"]; ok && v != nil {
		b, err := json.Marshal(v)
		if err != nil {
			return nil, fmt.Errorf("board get: failed to marshal layout: %w", err)
		}
		o.Layout = string(b)
	}
	if v, ok := raw["stages"]; ok && v != nil {
		b, err := json.Marshal(v)
		if err != nil {
			return nil, fmt.Errorf("board get: failed to marshal stages: %w", err)
		}
		o.Stages = string(b)
	}
	return o, nil
}

var gqlDeleteBoard = compileGqlQuery(
`mutation Board_Delete($id: ObjectId!) {
		deleteDashboard(id: $id) {
			success
			errorMessage
		}
	}`,
	"data", "deleteDashboard",
)

func (ot *objectTypeBoard) Delete(cfg *Config, op Output, hc httpClient, id string) error {
	obj, err := gqlDeleteBoard.query(cfg, op, hc, object{"id": id})
	if err != nil {
		return err
	}
	if obj == nil {
		return nil
	}
	result, ok := obj.(object)
	if !ok {
		return fmt.Errorf("board delete: unexpected response type")
	}
	if errMsg, ok := result["errorMessage"]; ok && errMsg != nil {
		if s, ok := errMsg.(string); ok && s != "" {
			return fmt.Errorf("board delete: %s", s)
		}
	}
	return nil
}

func (ot *objectTypeBoard) Create(cfg *Config, op Output, hc httpClient, input object) (ObjectInstance, error) {
	return nil, nil
}

func (ot *objectTypeBoard) Update(cfg *Config, op Output, hc httpClient, id string, input object) (ObjectInstance, error) {
	return nil, nil
}
