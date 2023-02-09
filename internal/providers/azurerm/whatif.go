package azurerm

import (
	"encoding/json"

	"github.com/tidwall/gjson"
)

type ChangeType string
type PropertyChangeType string

const (
	Create      ChangeType = "Create"
	Delete      ChangeType = "Delete"
	Deploy                 = "Deploy"
	Ignore                 = "Ignore"
	Modify                 = "Modify"
	NoChange               = "NoChange"
	Unsupported            = "Unsupported"
)

const (
	PropCreate   PropertyChangeType = "Create"
	PropDelete   PropertyChangeType = "Delete"
	PropArray                       = "Array"
	PropModify                      = "Modify"
	PropNoEffect                    = "NoEffect"
)

// Struct for deserializing the JSON response of a whatif call
// Modeled after the schema of deployments/whatIf in the AzureRM REST API
// see: https://learn.microsoft.com/en-us/rest/api/resources/deployments/what-if-at-subscription-scope
type WhatIf struct {
	Status  string          `json:"status"`
	Error   ErrorResponse   `json:"error,omitempty"`
	Changes []*WhatifChange `json:"changes,omitempty"`
}

type WhatifProperties struct {
	CorrelationId string `json:"correlationId"`
}

type WhatifChange struct {
	ResourceId        string                  `json:"resourceId"`
	UnsupportedReason string                  `json:"unsupportedReason,omitempty"`
	ChangeType        ChangeType              `json:"changeType"`
	Delta             []*WhatIfPropertyChange `json:"delta,omitempty"`

	BeforeRaw json.RawMessage `json:"before,omitempty"`
	AfterRaw  json.RawMessage `json:"after,omitempty"`

	before gjson.Result
	after  gjson.Result
}

func (w *WhatifChange) After() (*gjson.Result, error) {
	if w.after.Get("id").Exists() {
		return &w.after, nil
	}
	w.after = gjson.ParseBytes(w.AfterRaw)
	return &w.after, nil
}

func (w *WhatifChange) Before() (*gjson.Result, error) {
	if w.before.Get("id").Exists() {
		return &w.before, nil
	}
	w.before = gjson.ParseBytes(w.BeforeRaw)
	return &w.before, nil
}

type WhatIfPropertyChange struct {
	PropertyChangeType PropertyChangeType
	AfterRaw           json.RawMessage         `json:"after,omitempty"`
	BeforeRaw          json.RawMessage         `json:"before,omitempty"`
	Children           []*WhatIfPropertyChange `json:"children,omitempty"`

	Path   string `json:"path,omitempty"`
	before gjson.Result
	after  gjson.Result
}

func (w *WhatIfPropertyChange) After() (*gjson.Result, error) {
	if w.after.Get("path").Exists() {
		return &w.after, nil
	}
	w.after = gjson.ParseBytes(w.AfterRaw)
	return &w.after, nil
}

func (w *WhatIfPropertyChange) Before() (*gjson.Result, error) {
	if w.before.Get("path").Exists() {
		return &w.before, nil
	}
	w.before = gjson.ParseBytes(w.BeforeRaw)
	return &w.before, nil
}

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Target  string `json:"target"`
}

type ErrorAdditionalInfo struct {
	Info string `json:"info"`
	Type string `json:"type"`
}
