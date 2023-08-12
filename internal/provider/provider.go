// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/teraswitch/terraform-provider-teraswitch/internal/tsw"
)

// Ensure TSWProvider satisfies various provider interfaces.
var _ provider.Provider = &TSWProvider{}

// TSWProvider defines the provider implementation.
type TSWProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// TSWProviderModel describes the provider data model.
type TSWProviderModel struct {
	Endpoint types.String `tfsdk:"endpoint"`
	ApiToken types.String `tfsdk:"api_token"`
	//ProjectId types.Int64  `tfsdk:"project_id"`
}

func (p *TSWProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "teraswitch"
	resp.Version = p.version
}

func (p *TSWProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				MarkdownDescription: "TeraSwitch API URL",
				Optional:            true,
			},
			"api_token": schema.StringAttribute{
				Required:            true,
				Sensitive:           true,
				MarkdownDescription: "TeraSwitch REST API token",
			},
			//"project_id": schema.Int64Attribute{
			//	Required:            true,
			//	MarkdownDescription: "TeraSwitch project ID",
			//},
		},
	}
}

func (p *TSWProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data TSWProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	endpoint := data.Endpoint.ValueString()
	if endpoint == "" {
		endpoint = "https://api.tsw.io"
	}

	if resp.Diagnostics.HasError() {
		return
	}

	httpClient := http.DefaultClient
	client := tsw.NewClient(httpClient, endpoint, data.ApiToken.ValueString())
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *TSWProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		//NewComputeInstanceResource,
		NewSshKeyResource,
	}
}

func (p *TSWProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &TSWProvider{
			version: version,
		}
	}
}
