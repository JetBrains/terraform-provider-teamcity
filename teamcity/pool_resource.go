package teamcity

import (
    "fmt"
    "context"

    "terraform-provider-teamcity/client"
    "github.com/hashicorp/terraform-plugin-framework/resource"
    "github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

var (
    _ resource.Resource              = &poolResource{}
    _ resource.ResourceWithConfigure = &poolResource{}
)

func NewPoolResource() resource.Resource {
    return &poolResource{}
}

type poolResource struct {
    client *client.Client
}

// Resource functions implementation
// returns the full name of the resource
func (r *poolResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
    resp.TypeName = req.ProviderTypeName + "_pool"
}

// returns the schema of the resource
func (r *poolResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
    resp.Schema = schema.Schema{
		Description: "An Agent Pool in TeamCity is a group of agents that can be associated to projects. More info [here](https://www.jetbrains.com/help/teamcity/configuring-agent-pools.html)",
        Attributes: map[string]schema.Attribute{
            "name": schema.StringAttribute{
                Required: true, 
            },
            "id": schema.Int64Attribute{
                Computed: true,
            },
            "size": schema.Int64Attribute{
                Required: false,
				MarkdownDescription: "Agents capacity for the given pool, don't add for unlimited",
            },
        },
    }
}

// creates a resource and sets the initial terraform state
func (r *poolResource) Create(_ context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
      
}

// reads a resource and sets latest terraform state
func (r *poolResource) Read(_ context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {

}

// updates a resource and sets the latest updated terraform state
func (r *poolResource) Update(_ context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {

}

// deletes a resource and removes its terraform state
func (r *poolResource) Delete(_ context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {

}

// configure client
func (r *poolResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
    if req.ProviderData == nil {
        return
    }

    client, ok := req.ProviderData.(*client.Client)
    if !ok {
        resp.Diagnostics.AddError(
            "Unexpected Resource Configure Type",
            fmt.Sprintf("Expected *client.Client, got: %T.", req.ProviderData),
        )
        return
    }

    r.client = client
}
