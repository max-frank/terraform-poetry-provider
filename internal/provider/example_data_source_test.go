package provider

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccExampleDataSource(t *testing.T) {

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/title/test", r.URL.Path)
		w.WriteHeader(200)
		w.Write([]byte(`[
			{
				"author":    "test",
				"title":     "test",
				"lines":     ["foo", "bar"],
				"linecount": "2"
					
			}
		]`))
	}))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(testServer.URL),
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccExampleDataSourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.poetry.test", "poems.#", "1"),
					resource.TestCheckResourceAttr("data.poetry.test", "poems.0.author", "test"),
					resource.TestCheckResourceAttr("data.poetry.test", "poems.0.title", "test"),
					resource.TestCheckResourceAttr("data.poetry.test", "poems.0.line_count", "2"),
					resource.TestCheckResourceAttr("data.poetry.test", "poems.0.lines.#", "2"),
					resource.TestCheckResourceAttr("data.poetry.test", "poems.0.lines.0", "foo"),
					resource.TestCheckResourceAttr("data.poetry.test", "poems.0.lines.1", "bar"),
				),
			},
		},
	})
}

const testAccExampleDataSourceConfig = `
data "poetry" "test" {
	title = "test"
}
`
