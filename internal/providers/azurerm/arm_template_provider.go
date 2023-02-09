package azurerm

import (
	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/schema"
	"github.com/pkg/errors"
)

type DeploymentScope string

const (
	ResourceGroup   DeploymentScope = "group"
	ManagementGroup DeploymentScope = "mg"
	Subscription                    = "sub"
	Tenant                          = "tenant"
)

type DeploymentMode string

const (
	Complete    DeploymentMode = "Complete"
	Incremental DeploymentMode = "Incremental"
)

type ArmTemplateProviderOpts struct {
	Binary               string
	IncludePastResources bool
	Scope                DeploymentScope
	Mode                 DeploymentMode
	ParameterFile        string
	// Azure location , required when Scope = Subscription|Tenant|ManagementGroup
	Location string
	// Name of resource group, required when Scope = ResourceGroup
	ResourceGroup string
	// Id of the management group, required when Scope = ManagementGroup
	ManagementGroupId string
}

// TODO: AzureRM doesn't have a concept of a 'Project', needs its own config.ProjectContext object
type ArmTemplateProvider struct {
	binary            string
	parameterFile     string
	inner             *AzureRMWhatifProvider
	incremental       bool
	location          string
	resourceGroup     string
	managementGroupId string
}

func NewArmTemplateProvider(ctx *config.ProjectContext, opts *ArmTemplateProviderOpts) (schema.Provider, error) {
	azBinary := opts.Binary
	if azBinary == "" {
		azBinary = defaultAzBinary
	}

	whatIf, err := getWhatIfFromArmTemplate(ctx.ProjectConfig.Path, opts)
	if err != nil {
		return nil, err
	}

	innerProvider := newWhatifJsonProviderWithContent(ctx, whatIf, opts.IncludePastResources)

	return &ArmTemplateProvider{
		inner:             innerProvider,
		parameterFile:     opts.ParameterFile,
		binary:            azBinary,
		location:          opts.Location,
		managementGroupId: opts.ManagementGroupId,
		resourceGroup:     opts.ResourceGroup,
	}, nil
}

func (p *ArmTemplateProvider) Type() string {
	return "azurerm_whatif_json"
}

func (p *ArmTemplateProvider) DisplayType() string {
	return "Azure Resource Manager WhatIf JSON"
}

func (p *ArmTemplateProvider) AddMetadata(metadata *schema.ProjectMetadata) {
	// no op
}

func (p *ArmTemplateProvider) LoadResources(usage map[string]*schema.UsageData) ([]*schema.Project, error) {
	return p.inner.LoadResources(usage)
}

func getGroupDeploymentArgs(templateFile string, opts *ArmTemplateProviderOpts) ([]string, error) {
	if opts.ResourceGroup == "" {
		return nil, errors.New("Invalid resource group for resource group scoped deployment")
	}

	args := []string{
		"deployment",
		"group",
		"what-if",
		"--resource-group",
		opts.ResourceGroup,
		"--no-pretty-print",
	}

	if opts.ParameterFile != "" {
		args = append(args, "--parameters", opts.ParameterFile)
	}

	return args, nil
}
