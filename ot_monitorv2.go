package main

import (
	"encoding/json"
	"fmt"
)

func init() {
	RegisterObjectType(ObjectTypeMonitorV2, &objectMonitorV2{})
}

type objectMonitorV2 struct {
	Id          string
	Name        string
	Description string
	Disabled    string
	UpdatedDate string
	Definition  string
}

var _ ObjectInstance = &objectMonitorV2{}

func (o *objectMonitorV2) GetInfo() *ObjectInfo {
	return &ObjectInfo{
		Id:           o.Id,
		Name:         o.Name,
		Presentation: []string{o.Id, o.Name, o.Disabled, o.UpdatedDate},
		Object:       o,
	}
}

func (o *objectMonitorV2) GetValues() []PropertyInstance {
	props := ObjectTypeMonitorV2.GetProperties()
	r := make([]PropertyInstance, len(props))
	for i, p := range props {
		r[i] = &propertyInstance{p, o}
	}
	return r
}

func (o *objectMonitorV2) PrintToYaml(op Output, otyp ObjectType, obj ObjectInstance) error {
	return printToYamlFromObjectInstance(op, otyp, obj)
}

type objectTypeMonitorV2 struct{}

var ObjectTypeMonitorV2 ObjectType = &objectTypeMonitorV2{}

var propertyDescMonitorV2 = []PropertyDesc{
	{"id", PropertyTypeString, false, true,
		func(o any) any { return o.(*objectMonitorV2).Id },
		func(o any, v any) {
			if v != nil {
				o.(*objectMonitorV2).Id = v.(string)
			}
		}},
	{"name", PropertyTypeString, false, false,
		func(o any) any { return o.(*objectMonitorV2).Name },
		func(o any, v any) {
			if v != nil {
				o.(*objectMonitorV2).Name = v.(string)
			}
		}},
	{"description", PropertyTypeString, false, false,
		func(o any) any { return o.(*objectMonitorV2).Description },
		func(o any, v any) {
			if v != nil {
				o.(*objectMonitorV2).Description = v.(string)
			}
		}},
	{"disabled", PropertyTypeString, false, false,
		func(o any) any { return o.(*objectMonitorV2).Disabled },
		func(o any, v any) {
			if v != nil {
				o.(*objectMonitorV2).Disabled = fmt.Sprintf("%v", v)
			}
		}},
	{"updatedDate", PropertyTypeString, true, false,
		func(o any) any { return o.(*objectMonitorV2).UpdatedDate },
		func(o any, v any) {
			if v != nil {
				o.(*objectMonitorV2).UpdatedDate = v.(string)
			}
		}},
	{"definition", PropertyTypeString, true, false,
		func(o any) any { return o.(*objectMonitorV2).Definition },
		func(o any, v any) {
			if v != nil {
				o.(*objectMonitorV2).Definition = fmt.Sprintf("%v", v)
			}
		}},
}

func (*objectTypeMonitorV2) TypeName() string { return "monitor" }
func (*objectTypeMonitorV2) Help() string {
	return "A Monitor V2 watches a data pipeline and fires alarms when conditions are met."
}
func (*objectTypeMonitorV2) CanList() bool                   { return true }
func (*objectTypeMonitorV2) CanGet() bool                    { return true }
func (*objectTypeMonitorV2) CanCreate() bool                 { return false }
func (*objectTypeMonitorV2) CanUpdate() bool                 { return false }
func (*objectTypeMonitorV2) CanDelete() bool                 { return false }
func (*objectTypeMonitorV2) GetPresentationLabels() []string { return []string{"id", "name", "disabled", "updatedDate"} }
func (*objectTypeMonitorV2) GetProperties() []PropertyDesc   { return propertyDescMonitorV2 }

var gqlListMonitorV2 = compileGqlQuery(
	`query SearchMonitorV2($workspaceId: ObjectId!, $nameSubstring: String) {
		searchMonitorV2(workspaceId: $workspaceId, nameSubstring: $nameSubstring) {
			monitors { id name description disabled updatedDate }
		}
	}`,
	"data", "searchMonitorV2", "monitors",
)

func (ot *objectTypeMonitorV2) List(cfg *Config, op Output, hc httpClient) ([]*ObjectInfo, error) {
	workspaceId := cfg.WorkspaceIdOrName
	if workspaceId == "" {
		workspaceId = "42379913"
	}
	args := object{"workspaceId": workspaceId}
	obj, err := gqlListMonitorV2.query(cfg, op, hc, args)
	if err != nil || obj == nil {
		return nil, err
	}
	items, ok := obj.(array)
	if !ok {
		return nil, fmt.Errorf("monitor list: unexpected response type")
	}
	var ret []*ObjectInfo
	for _, item := range items {
		m, ok := item.(object)
		if !ok {
			continue
		}
		o := monitorV2FromObject(m)
		ret = append(ret, o.GetInfo())
	}
	return ret, nil
}

var gqlGetMonitorV2 = compileGqlQuery(
	`query GetMonitorV2($id: ObjectId!) {
		monitorV2(id: $id) {
			id name description disabled updatedDate
			definition {
				... on MonitorV2CountDefinition {
					compareFunction
					countAggFunction
					threshold
				}
			}
		}
	}`,
	"data", "monitorV2",
)

func (ot *objectTypeMonitorV2) Get(cfg *Config, op Output, hc httpClient, id string) (ObjectInstance, error) {
	obj, err := gqlGetMonitorV2.query(cfg, op, hc, object{"id": id})
	if err != nil {
		return nil, err
	}
	if obj == nil {
		return nil, nil
	}
	raw, ok := obj.(object)
	if !ok {
		return nil, fmt.Errorf("monitor get: unexpected response type")
	}
	o := monitorV2FromObject(raw)
	if v, ok := raw["definition"]; ok && v != nil {
		b, err := json.Marshal(v)
		if err != nil {
			return nil, fmt.Errorf("monitor get: failed to marshal definition: %w", err)
		}
		o.Definition = string(b)
	}
	return o, nil
}

func (ot *objectTypeMonitorV2) Create(cfg *Config, op Output, hc httpClient, input object) (ObjectInstance, error) {
	return nil, nil
}

func (ot *objectTypeMonitorV2) Update(cfg *Config, op Output, hc httpClient, id string, input object) (ObjectInstance, error) {
	return nil, nil
}

func (ot *objectTypeMonitorV2) Delete(cfg *Config, op Output, hc httpClient, id string) error {
	return nil
}

// monitorV2FromObject populates an objectMonitorV2 from a GraphQL response map.
func monitorV2FromObject(m object) *objectMonitorV2 {
	o := &objectMonitorV2{}
	if v, ok := m["id"]; ok && v != nil {
		o.Id = v.(string)
	}
	if v, ok := m["name"]; ok && v != nil {
		o.Name = v.(string)
	}
	if v, ok := m["description"]; ok && v != nil {
		o.Description = v.(string)
	}
	if v, ok := m["disabled"]; ok && v != nil {
		o.Disabled = fmt.Sprintf("%v", v)
	}
	if v, ok := m["updatedDate"]; ok && v != nil {
		o.UpdatedDate = v.(string)
	}
	return o
}
