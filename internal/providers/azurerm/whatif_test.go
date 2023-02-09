package azurerm

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Tests to check if serializing the Whatif JSON works properly, for development
// Pretty much testing Go's internal JSON system and gjson, so not particularly useful in the end
// TODO: Remove this when done with contribution
func TestWhatifSerialization(t *testing.T) {
	testDataPath := "./testdata"

	testFiles := []string{"whatif-single.json"}
	expected := []WhatIf{
		{
			Status: "Succeeded",
			Changes: []*WhatifChange{
				{
					ResourceId: "/subscriptions/00000000-0000-0000-0000-000000000001/resourceGroups/my-resource-group2",
					ChangeType: Create,
					BeforeRaw:  []byte("{\"apiVersion\":\"2018-11-30\",\"id\":\"/subscriptions/00000000-0000-0000-0000-000000000001/resourceGroups/my-resource-group/providers/Microsoft.ManagedIdentity/userAssignedIdentities/myExistingIdentity\",\"type\":\"Microsoft.ManagedIdentity/userAssignedIdentities\",\"name\":\"myExistingIdentity\",\"location\":\"westus2\"}"),
					AfterRaw:   []byte("{\"apiVersion\":\"2018-11-30\",\"id\":\"/subscriptions/00000000-0000-0000-0000-000000000001/resourceGroups/my-resource-group/providers/Microsoft.ManagedIdentity/userAssignedIdentities/myExistingIdentity\",\"type\":\"Microsoft.ManagedIdentity/userAssignedIdentities\",\"name\":\"myExistingIdentity\",\"location\":\"westus2\",\"tags\": {\"myNewTag\":\"my tag value\"}}"),
					Delta: []*WhatIfPropertyChange{
						{
							Path:               "tags.myNewTag",
							PropertyChangeType: PropCreate,
							AfterRaw:           []byte("\"my tag value\""),
							Children: []*WhatIfPropertyChange{
								{
									Path:               "tags.myNewTag2",
									PropertyChangeType: PropCreate,
									AfterRaw:           []byte("\"my tag value2\""),
									Children: []*WhatIfPropertyChange{
										{
											Path:               "tags.myNewTag3",
											PropertyChangeType: PropCreate,
											AfterRaw:           []byte("\"my tag value3\""),
										},
									},
								},
								{
									Path:               "tags.myNewTag4",
									PropertyChangeType: PropCreate,
									AfterRaw:           []byte("\"my tag value4\""),
								},
							},
						},
					},
				},
			},
		},
	}

	for i, f := range testFiles {
		exp := expected[i]
		var whatIf WhatIf
		filePath := path.Join(testDataPath, f)

		file, err := ioutil.ReadFile(filePath)
		if err != nil {
			log.Fatalf("Failed to read test json file")
		}

		json.Unmarshal(file, &whatIf)

		assert.Equal(t, len(exp.Changes), len(whatIf.Changes))

		for j, c := range whatIf.Changes {
			wiBefore, err := c.Before()
			if err != nil {
				t.Fatal(err.Error())
			}

			wiAfter, err := c.After()
			if err != nil {
				t.Fatal(err.Error())
			}

			expBefore, err := exp.Changes[j].Before()
			if err != nil {
				t.Fatal(err.Error())
			}

			expAfter, err := exp.Changes[j].After()
			if err != nil {
				t.Fatal(err.Error())
			}

			assert.Equal(t, exp.Changes[j].Delta, c.Delta)

			assert.Equal(t, expBefore.Get("apiVersion").Str, wiBefore.Get("apiVersion").Str)
			assert.Equal(t, expBefore.Get("id").Str, wiBefore.Get("id").Str)
			assert.Equal(t, expBefore.Get("type").Str, wiBefore.Get("type").Str)

			assert.Equal(t, expAfter.Get("apiVersion").Str, wiAfter.Get("apiVersion").Str)
			assert.Equal(t, expAfter.Get("id").Str, wiAfter.Get("id").Str)
			assert.Equal(t, expAfter.Get("type").Str, wiAfter.Get("type").Str)
		}
	}
}
