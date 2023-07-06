package teamcity

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"io"
	"net/http"
	"terraform-provider-teamcity/client"
	"time"
)

var (
	_ datasource.DataSource              = &buildConfDataSource{}
	_ datasource.DataSourceWithConfigure = &buildConfDataSource{}
)

func NewBuildConfDataSource() datasource.DataSource {
	return &buildConfDataSource{}
}

type buildConfDataSource struct {
	client *client.Client
}

type buildConfDataSourceModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

type buildConfJson struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (d *buildConfDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = req.ProviderData.(*client.Client)
}

func (d *buildConfDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_build_configuration"
}

func (d *buildConfDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required: true,
			},
			"name": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (d *buildConfDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var conf buildConfDataSourceModel
	diags := req.Config.Get(ctx, &conf)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state buildConfDataSourceModel

	rclient := retryablehttp.NewClient()
	rclient.RetryWaitMin = 5 * time.Second
	rclient.RetryWaitMax = 5 * time.Second
	rclient.RetryMax = 60
	rclient.CheckRetry = retryPolicy

	request, err := retryablehttp.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/buildTypes/id:%s", d.client.RestURL, conf.ID.ValueString()), nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create HTTP request",
			err.Error(),
		)
		return
	}

	if d.client.Token != "" {
		request.Header.Set("Authorization", "Bearer "+d.client.Token)
	} else {
		request.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(d.client.Username+":"+d.client.Password)))
	}
	request.Header.Set("Accept", "application/json")

	response, err := rclient.Do(request)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error performing HTTP request",
			err.Error(),
		)
		return
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading HTTP response",
			err.Error(),
		)
		return
	}

	data := buildConfJson{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error decoding JSON",
			err.Error()+"\n"+fmt.Sprintf("Body: %q", body),
		)
		return
	}

	state.ID = types.StringValue(data.ID)
	state.Name = types.StringValue(data.Name)

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func retryPolicy(ctx context.Context, resp *http.Response, err error) (bool, error) {
	if resp.StatusCode == 404 {
		return true, nil
	}

	return retryablehttp.DefaultRetryPolicy(ctx, resp, err)
}
