package providers

import (
	"testing"

	"github.com/infracost/infracost/internal/config"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

type DetectTest struct {
	Path     string
	Expected string
}

func TestAzureRmProviderDetection(t *testing.T) {
	tests := []DetectTest{
		// {
		// 	Path:     "../../examples/azurerm/web_app/what_if.json",
		// 	Expected: "azurerm_whatif_json",
		// },
		{
			Path:     "../../examples/azurerm/web_app/azuredeploy.json",
			Expected: "azurerm_template_json",
		},
	}

	for _, test := range tests {
		func(path string, expected string) {
			ctx := config.NewProjectContext(config.EmptyRunContext(), &config.Project{}, log.Fields{})
			ctx.ProjectConfig.Path = path

			res, err := Detect(ctx, true)
			if err != nil {
				t.Fatal("Detect threw an error: " + err.Error())
			}

			assert.Equal(t, expected, res.Type())
		}(test.Path, test.Expected)
	}
}
