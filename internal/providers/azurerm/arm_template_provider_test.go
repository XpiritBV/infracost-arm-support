package azurerm

import (
	"path/filepath"
	"testing"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/usage"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestArmTemplateProvider(t *testing.T) {
	ctx := config.NewProjectContext(config.EmptyRunContext(), &config.Project{}, log.Fields{})
	ctx.ProjectConfig.Path = filepath.Join("testdata", "azuredeploy.group.json")
	opts := &ArmTemplateProviderOpts{
		IncludePastResources: true,
		Location:             "westeurope",
		ResourceGroup:        "rg-infracost-test",
		Scope:                ResourceGroup,
		Mode:                 Incremental,
	}

	provider, err := NewArmTemplateProvider(ctx, opts)
	if err != nil {
		t.Fatalf(errors.Wrap(err, "Failed constructing ARM template provider").Error())
	}

	usage := usage.NewBlankUsageFile().ToUsageDataMap()
	project, err := provider.LoadResources(usage)
	if err != nil {
		t.Fatalf("Error loading resources: " + err.Error())
	}

	// Ensure all resources in the whatif are returned from the provider
	assert.Equal(t, 3, len(project[0].PartialResources))
	assert.Equal(t, 0, len(project[0].PartialPastResources))
}
