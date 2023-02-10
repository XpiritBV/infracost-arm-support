package azurerm

import (
	"fmt"
	"path/filepath"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/ui"
	"github.com/pkg/errors"
)

type DeploymentScope string

const (
	ResourceGroup   DeploymentScope = "resourceGroup"
	ManagementGroup DeploymentScope = "managementGroup"
	Subscription                    = "subscription"
	Tenant                          = "tenant"
)

type DeploymentMode string

const (
	Complete    DeploymentMode = "Complete"
	Incremental DeploymentMode = "Incremental"
)

type ArmDeploymentOpts struct {
	Binary            string
	Scope             DeploymentScope
	Mode              DeploymentMode
	ParameterFile     string
	Location          string
	ResourceGroup     string
	ManagementGroupId string
}

func (p *ArmTemplateProvider) Type() string {
	if filepath.Ext(p.ctx.ProjectConfig.Path) == ".bicep" {
		return "azurerm_bicep_template"
	}
	return "azurerm_template_json"
}

func (p *ArmTemplateProvider) DisplayType() string {
	if filepath.Ext(p.ctx.ProjectConfig.Path) == ".bicep" {
		return "Azure Bicep Template"
	}
	return "Azure Resource Manager WhatIf JSON"
}

func (p *ArmTemplateProvider) AddMetadata(metadata *schema.ProjectMetadata) {
	// no op
}

func NewArmTemplateProviderOptsFromProject(ctx *config.ProjectContext) *ArmDeploymentOpts {
	deploymentMode := ctx.ProjectConfig.ArmResourceGroup
	if deploymentMode == "" {
		deploymentMode = "incremental"
	}

	azBinary := ctx.ProjectConfig.AzBinary
	if azBinary == "" {
		azBinary = defaultAzBinary
	}

	return &ArmDeploymentOpts{
		Binary:            azBinary,
		Scope:             DeploymentScope(ctx.ProjectConfig.ArmDeploymentScope),
		Mode:              DeploymentMode(deploymentMode),
		ParameterFile:     ctx.ProjectConfig.ArmParametersPath,
		Location:          ctx.ProjectConfig.ArmLocation,
		ResourceGroup:     ctx.ProjectConfig.ArmResourceGroup,
		ManagementGroupId: ctx.ProjectConfig.ArmManagementGroupId,
	}
}

// TODO: AzureRM doesn't have a concept of a 'Project', needs its own config.ProjectContext object
type ArmTemplateProvider struct {
	inner                *AzureRMWhatifProvider
	ctx                  *config.ProjectContext
	opts                 *ArmDeploymentOpts
	includePastResources bool
}

func NewArmTemplateProvider(ctx *config.ProjectContext, includePastResources bool) (schema.Provider, error) {
	opts := NewArmTemplateProviderOptsFromProject(ctx)

	return &ArmTemplateProvider{
		ctx:                  ctx,
		opts:                 opts,
		includePastResources: includePastResources,
	}, nil
}

func (p *ArmTemplateProvider) LoadResources(usage map[string]*schema.UsageData) ([]*schema.Project, error) {
	spinner := ui.NewSpinner("Converting ARM template to WhatIf file...", ui.SpinnerOptions{
		EnableLogging: p.ctx.RunContext.Config.IsLogging(),
		NoColor:       p.ctx.RunContext.Config.NoColor,
		Indent:        "  ",
	})
	defer spinner.Fail()

	fmt.Print("Converting ARM Template to WhatIf file")
	whatIf, err := p.getWhatIfFromArmTemplate()
	p.inner = newWhatifJsonProviderWithContent(p.ctx, whatIf, p.includePastResources)

	if err != nil {
		return nil, err
	}

	return p.inner.LoadResources(usage)
}

func (p *ArmTemplateProvider) getWhatIfFromArmTemplate() ([]byte, error) {
	var args []string
	var err error
	templateFile := p.ctx.ProjectConfig.Path

	switch p.opts.Scope {
	case ResourceGroup:
		args, err = getGroupDeploymentArgs(templateFile, p.opts)
	default:
		err = errors.New(fmt.Sprintf("Unsupported scope %s", p.opts.Scope))
	}
	if err != nil {
		return nil, err
	}

	templateDir := filepath.Dir(templateFile)
	cmdOpts := &CmdOptions{
		Binary: p.opts.Binary,
		Flags:  args,
		Dir:    templateDir,
	}

	output, err := Cmd(cmdOpts)
	if err != nil {
		return nil, err
	}

	return output, nil
}
