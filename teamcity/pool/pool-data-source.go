package pool

import (
    "context"

    "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
    "github.com/hashicorp/terraform-plugin-framework/datasource"
    "github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-teamcity/client"
)

type PoolModel struct {
    Name        types.String    `tfsdk:"name"`
    Id          types.String    `tfsdk:"id"`
    Size        types.Int       `tfsdk:"size"`
}

var (
    _ datasource.DataSource                 = &poolDataSource{}
    _ datasource.DataSourceWithConfigure    = &poolDataSource{}
)

func NewPoolDataSource() datasource.DataSource {
    return &poolDataSource{}
}

type poolDataSource struct {
    client *client.Client
}


// DataSource functions implementation
// returns the full name of the data source
func (d *poolDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
    resp.TypeName = req.ProviderTypeName + "_pool"
}

// returns the schema of the data source
func (d *poolDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
    resp.Schema = schema.Schema {
        Description: "A Agent Pool in TeamCity is a group of agents that can be associated to projects. More info [here](https://www.jetbrains.com/help/teamcity/configuring-agent-pools.html)",
        Attributes: map[string]schema.Attribute{
            "name" : schema.StringAttribute {
                Computed: true,
            },
            "id": schema.Int64Attribute {
                Computed: true,
            },
            "size": schema.Int64Attribute {
                Computed: true,
                MarkdownDescription: "Agents capacity for the given pool"
            },
        },
    }
}

// returns the state of the data source
func (d *poolDataSource) Read(_ context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {

}

// DataSourceWithConfigure functions implementation
func (d *poolDatasource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasoruce.ConfigureReponse) {
    if req.ProviderData == nil {
        return
    }

    client, ok := req.ProviderData.(*client.Client); if !ok {
        resp.Diagnostics.AddError(
            "Unexpected Data Source Configure Type",
            fmt.Sprintf("Expected *client.Client, got: %T.", req.ProviderData),
        )
        return
    }

    d.client = client
}
