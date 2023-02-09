package azurerm

import (
	"encoding/json"
	"fmt"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/providers/azurerm/resources"
	"github.com/infracost/infracost/internal/schema"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
)

type usageMap map[string]*schema.UsageData

type Parser struct {
	ctx *config.ProjectContext
}

func NewParser(ctx *config.ProjectContext) *Parser {
	return &Parser{ctx}
}

type ParsedWhatifChange struct {
	PartialResource     *schema.PartialResource
	PartialPastResource *schema.PartialResource
	Delta               []*WhatIfPropertyChange
}

// Same as providers/terraform/parser.go:createPartialResource
func (p *Parser) createPartialResource(d *schema.ResourceData, u *schema.UsageData) *schema.PartialResource {
	registryMap := resources.GetRegistryMap()

	if registryItem, ok := (*registryMap)[d.Type]; ok {
		if registryItem.NoPrice {
			return &schema.PartialResource{
				ResourceData: d,
				Resource: &schema.Resource{
					Name:         d.Address,
					ResourceType: d.Type,
					Tags:         d.Tags,
					IsSkipped:    true,
					NoPrice:      true,
					SkipMessage:  "Free resource",
				},
				CloudResourceIDs: registryItem.CloudResourceIDFunc(d),
			}
		}

		if registryItem.CoreRFunc != nil {
			coreRes := registryItem.CoreRFunc(d)
			if coreRes != nil {
				return &schema.PartialResource{ResourceData: d, CoreResource: coreRes, CloudResourceIDs: registryItem.CloudResourceIDFunc(d)}
			}
		} else {
			res := registryItem.RFunc(d, u)
			if res != nil {
				if u != nil {
					res.EstimationSummary = u.CalcEstimationSummary()
				}

				return &schema.PartialResource{ResourceData: d, Resource: res, CloudResourceIDs: registryItem.CloudResourceIDFunc(d)}
			}
		}
	}

	return &schema.PartialResource{
		ResourceData: d,
		Resource: &schema.Resource{
			Name:        d.Address,
			IsSkipped:   true,
			SkipMessage: "This resource is not currently supported",
		},
	}
}

func (p *Parser) parse(j []byte, usage usageMap) ([]*ParsedWhatifChange, error) {
	var changes []*ParsedWhatifChange
	var whatif WhatIf

	err := json.Unmarshal(j, &whatif)
	if err != nil {
		return nil, errors.New("Failed to unmarshal whatif operation result")
	}

	if whatif.Status != "Succeeded" {
		return nil, errors.New("WhatIf operation was not successful")
	}

	for _, change := range whatif.Changes {
		parsed, err := p.parseChange(change, usage)
		if err != nil {
			return nil, err
		}

		if parsed != nil {
			changes = append(changes, parsed)
		}
	}

	// Recursively create WhatIfPropertyChanges

	return changes, nil
}

// TODO: need baseresources like TF provider?
func (p *Parser) parseChange(change *WhatifChange, usage usageMap) (*ParsedWhatifChange, error) {
	var after *schema.PartialResource
	var before *schema.PartialResource

	beforeData, err := change.Before()
	if err != nil {
		return nil, err
	}

	afterData, err := change.After()
	if err != nil {
		return nil, err
	}

	if afterData.Get("id").Exists() {
		afterRd, err := p.parseResourceData(afterData)
		if err != nil {
			return nil, err
		}
		after = p.createPartialResource(afterRd, afterRd.UsageData)
	}

	if beforeData.Get("id").Exists() {
		beforeRd, err := p.parseResourceData(beforeData)
		if err != nil {
			return nil, err
		}
		before = p.createPartialResource(beforeRd, beforeRd.UsageData)
	}

	return &ParsedWhatifChange{
		PartialResource:     after,
		PartialPastResource: before,
		Delta:               change.Delta,
	}, nil
}

// TODO: This is not exhaustive yet, probably need to do something with 'Delta' and 'WhatIfPropertyChange'
func (p *Parser) parseResourceData(data *gjson.Result) (*schema.ResourceData, error) {
	var resData *schema.ResourceData

	armType := data.Get("type")
	resId := data.Get("id")
	if !armType.Exists() || !resId.Exists() {
		return nil, errors.New(fmt.Sprintf("Failed to parse resource data"))
	}

	tfType := resources.GetTFResourceFromAzureRMType(armType.Str, data)
	if len(tfType) == 0 {
		return nil, errors.New(fmt.Sprintf("Could not convert AzureRM type '%s' to TF type", armType.Str))
	}

	resData = schema.NewAzureRMResourceData(tfType, resId.Str, *data)

	return resData, nil
}
