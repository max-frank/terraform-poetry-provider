package provider

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ tfsdk.DataSourceType = poemDataSourceType{}
var _ tfsdk.DataSource = poemDataSource{}

type APIPoem struct {
	Author    string   `tfsdk:"author" json:"author"`
	Title     string   `tfsdk:"title"  json:"title"`
	Lines     []string `tfsdk:"lines"  json:"lines"`
	LineCount string   `tfsdk:"line_count"  json:"linecount"`
}

type Poem struct {
	Author    string   `tfsdk:"author" json:"author"`
	Title     string   `tfsdk:"title"  json:"title"`
	Lines     []string `tfsdk:"lines"  json:"lines"`
	LineCount int64    `tfsdk:"line_count"  json:"linecount"`
}

type poemDataSourceType struct{}

func (t poemDataSourceType) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Poem data source",

		Attributes: map[string]tfsdk.Attribute{
			"title": {
				MarkdownDescription: "The poem title to lookup",
				Type:                types.StringType,
				Required:            true,
				Optional:            false,
			},
			"poems": {
				MarkdownDescription: "The list of poems retrieved",
				Type: types.ListType{
					ElemType: types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"author": types.StringType,
							"title":  types.StringType,
							"lines": types.ListType{
								ElemType: types.StringType,
							},
							"line_count": types.Int64Type,
						},
					},
				},
				Computed: true,
			},
			"id": {
				MarkdownDescription: "The id",
				Type:                types.StringType,
				Computed:            true,
			},
		},
	}, nil
}

func (t poemDataSourceType) NewDataSource(ctx context.Context, in tfsdk.Provider) (tfsdk.DataSource, diag.Diagnostics) {
	provider, diags := convertProviderType(in)

	return poemDataSource{
		provider: provider,
	}, diags
}

type poemDataSourceData struct {
	Poems       []Poem       `tfsdk:"poems"`
	Id          types.String `tfsdk:"id"`
	SearchTitle string       `tfsdk:"title"`
}

type poemDataSource struct {
	provider provider
}

func (d poemDataSource) Read(ctx context.Context, req tfsdk.ReadDataSourceRequest, resp *tfsdk.ReadDataSourceResponse) {
	if resp.Diagnostics.HasError() {
		return
	}

	var data poemDataSourceData
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	requestURL := fmt.Sprintf("%s/title/%s", d.provider.apiEndpoint, data.SearchTitle)
	c := http.Client{Timeout: time.Duration(30) * time.Second}
	httpResp, err := c.Get(requestURL)
	if err != nil {
		fmt.Printf("Error %s", err)
		return
	}

	defer httpResp.Body.Close()
	body, err := ioutil.ReadAll(httpResp.Body)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read response: %s", err))
		return
	}

	apiPoems := new([]APIPoem)
	err = json.Unmarshal(body, apiPoems)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to decode response: %s", err))
		return
	}

	poems := []Poem{}
	// convert linecounts
	for _, p := range *apiPoems {
		lineCount, err := strconv.ParseInt(p.LineCount, 10, 64)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to decode response: %s", err))
			return
		}
		poems = append(poems, Poem{
			Author:    p.Author,
			Title:     p.Title,
			Lines:     p.Lines,
			LineCount: lineCount,
		})
	}

	data.Id = types.String{Value: fmt.Sprintf("%x", md5.Sum(body))}
	data.Poems = poems

	diags = resp.State.Set(ctx, data)
	resp.Diagnostics.Append(diags...)
}
