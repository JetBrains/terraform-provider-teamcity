package teamcity

import (
    "context"

    "github.com/hashicorp/terraform-plugin-framework/resource"
    "github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

var (
    _ resource.Resource = &poolResource{}
)

func NewPoolResource() resource.Resource {
    return &poolResource{}
}

type poolResource struct {
    client *client.Client
}

// Resource functions implementation
// returns the full name of the resource
func (r *poolResource) Metadata(_ context.Context, req resouce.MetadataRequest, resp *resouce.MetadataResponse) {
    resp.TypeName = req.ProviderTypeName + "_pool"
}

// returns the schema of the resource
func (r *poolResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
    resp.Schema = schema.Schema{}
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
