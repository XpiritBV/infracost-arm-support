package azurerm

import (
	"os"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/ui"
	"github.com/pkg/errors"
)

// TODO: AzureRM doesn't have a concept of a 'Project', needs its own config.ProjectContext object
type AzureRMWhatifProvider struct {
	ctx                  *config.ProjectContext
	Path                 string
	includePastResources bool
	content              []byte
}

func NewWhatifJsonProvider(ctx *config.ProjectContext, includePastResources bool) *AzureRMWhatifProvider {
	return &AzureRMWhatifProvider{
		ctx:                  ctx,
		Path:                 ctx.ProjectConfig.Path,
		includePastResources: includePastResources,
	}
}

func newWhatifJsonProviderWithContent(ctx *config.ProjectContext, content []byte, includePastResources bool) *AzureRMWhatifProvider {
	return &AzureRMWhatifProvider{
		ctx:                  ctx,
		Path:                 ctx.ProjectConfig.Path,
		includePastResources: includePastResources,
		content:              content,
	}
}

func (p *AzureRMWhatifProvider) Type() string {
	return "azurerm_whatif_json"
}

func (p *AzureRMWhatifProvider) DisplayType() string {
	return "Azure Resource Manager WhatIf JSON"
}

func (p *AzureRMWhatifProvider) AddMetadata(metadata *schema.ProjectMetadata) {
	// no op
}

func (p *AzureRMWhatifProvider) LoadResources(usage map[string]*schema.UsageData) ([]*schema.Project, error) {
	spinner := ui.NewSpinner("Extracting only cost-related params from WhatIf", ui.SpinnerOptions{
		EnableLogging: p.ctx.RunContext.Config.IsLogging(),
		NoColor:       p.ctx.RunContext.Config.NoColor,
		Indent:        "  ",
	})
	defer spinner.Fail()

	if p.content == nil || len(p.content) == 0 {
		j, err := os.ReadFile(p.Path)
		if err != nil {
			return []*schema.Project{}, errors.Wrap(err, "Error reading WhatIf result JSON file")
		}
		p.content = j
	}

	metadata := config.DetectProjectMetadata(p.ctx.ProjectConfig.Path)
	metadata.Type = p.Type()
	p.AddMetadata(metadata)

	name := p.ctx.ProjectConfig.Name
	if name == "" {
		name = metadata.GenerateProjectName(p.ctx.RunContext.VCSMetadata.Remote, p.ctx.RunContext.IsCloudEnabled())
	}

	// TODO: This should probably do a call to the whatif endpoint for the subscription
	// Then pass the response to the code below
	// For now, pass a whatif result JSON file directly

	project := schema.NewProject(name, metadata)
	parser := NewParser(p.ctx)

	// TODO: pastResources are ??, check what they are in Azure context
	whatIfResources, err := parser.parse(p.content, usage)
	if err != nil {
		return []*schema.Project{}, errors.Wrap(err, "Error parsing WhatIf data")
	}

	for _, res := range whatIfResources {
		if res.PartialPastResource != nil {
			project.PartialPastResources = append(project.PartialPastResources, res.PartialPastResource)
		}
		if res.PartialResource != nil {
			project.PartialResources = append(project.PartialResources, res.PartialResource)
		}
	}

	spinner.Success()

	return []*schema.Project{project}, nil
}
