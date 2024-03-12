package teamcity

import (
    "fmt"
    "context"

    "terraform-provider-teamcity/client"
    "terraform-provider-teamcity/models"
    "github.com/hashicorp/terraform-plugin-framework/path"
    "github.com/hashicorp/terraform-plugin-framework/types"
    "github.com/hashicorp/terraform-plugin-framework/types/basetypes"
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
                Optional: true,
				MarkdownDescription: "Agents capacity for the given pool, don't add for unlimited",
            },
        },
    }
}

// creates a resource and sets the initial terraform state
func (r *poolResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    // get values from plan 
    var plan models.PoolDataModel
    diags := req.Plan.Get(ctx, &plan)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }

    // verify values
    if plan.Name.IsNull() {
        resp.Diagnostics.AddAttributeError(
            path.Root("name"),
            "Agent Pool name cannot be null",
            "The Resource cannot create an Agent Pool since there is an invalid configuration value for the Agent Pool name.",
        )
    }
    if resp.Diagnostics.HasError() {
        return
    }

    // Generate API request
    var pool models.PoolJson
    var size int64

    pool.Name       = plan.Name.ValueString()
    if !plan.Size.IsNull() {
        size        = plan.Size.ValueInt64()
        pool.Size   = &size
    }
   
    // Create new agent pool
    result, err := r.client.NewPool(pool)
    if err != nil {
       resp.Diagnostics.AddError(
            "Error creating pool",
            "Cannot create pool, unexpected error: " + err.Error(),
       )
       return
    }

    // Populate computed attributes
    plan.Id = types.Int64Value(int64(*(result.Id)))
    
    // Set state
    diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// reads a resource and sets latest terraform state
func (r *poolResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
    // get current state
    var state models.PoolDataModel
    diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

    // get refreshed pool
    pool, err := r.client.GetPool(state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("name"),
			"Agent Pool not found",
			"The Datasource cannot get an Agent Pool since there is no Agent Pool with the provided name.",
		)
        return
	}

    // overwrite with refreshed state
    if pool.Size == nil {
        state = models.PoolDataModel{
            Name: types.StringValue(string(pool.Name)),
            Size: basetypes.NewInt64Null(),
            Id:   types.Int64Value(int64(*(pool.Id))),
        }
    } else {
        state = models.PoolDataModel{
            Name: types.StringValue(string(pool.Name)),
            Size: types.Int64Value(int64(*(pool.Size))),
            Id:   types.Int64Value(int64(*(pool.Id))),
        }
    }

    // set state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
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
