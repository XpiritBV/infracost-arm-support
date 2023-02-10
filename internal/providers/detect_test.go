package providers

import (
	"os"
	"testing"

	"github.com/infracost/infracost/internal/config"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

type DetectTest struct {
	Path     string
	Config   string
	Expected string
}

type configSpec struct {
	Version  string            `yaml:"version"`
	Projects []*config.Project `yaml:"projects" ignored:"true"`
}

func TestAzureRmProviderDetection(t *testing.T) {
	tests := []DetectTest{
		{
			Path:     "../../examples/azurerm/web_app/what_if.json",
			Expected: "azurerm_whatif_json",
			Config:   "../../examples/azurerm/web_app/infracost-config.yml",
		},
		{
			Path:     "../../examples/azurerm/web_app/arm/azuredeploy.json",
			Expected: "azurerm_template_json",
			Config:   "../../examples/azurerm/web_app/infracost-config.yml",
		},
		{
			Path:     "../../examples/azurerm/web_app/bicep/main.bicep",
			Expected: "azurerm_bicep_template",
			Config:   "../../examples/azurerm/web_app/infracost-config.yml",
		},
	}

	for _, test := range tests {
		func(path string, expected string) {
			var projectCfg configSpec

			configData, err := os.ReadFile(test.Config)
			if err != nil {
				t.Fatal(err.Error())
			}

			err = yaml.Unmarshal(configData, &projectCfg)
			if err != nil {
				t.Fatal(err.Error())
			}

			ctx := config.NewProjectContext(config.EmptyRunContext(), projectCfg.Projects[0], log.Fields{})
			ctx.ProjectConfig.Path = path

			res, err := Detect(ctx, true)
			if err != nil {
				t.Fatal("Detect threw an error: " + err.Error())
			}

			assert.Equal(t, expected, res.Type())
		}(test.Path, test.Expected)
	}
}
