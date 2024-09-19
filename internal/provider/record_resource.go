// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"terraform-provider-usgdns/internal/usgdns"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &recordResource{}
	_ resource.ResourceWithConfigure   = &recordResource{}
	_ resource.ResourceWithImportState = &recordResource{}
)

// NewRecordResource is a helper function to simplify the provider implementation.
func NewRecordResource() resource.Resource {
	return &recordResource{}
}

// recordResource is the resource implementation.
type recordResource struct {
	client *usgdns.Client
}

// Metadata returns the resource type name.
func (r *recordResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_record"
}

// Schema defines the schema for the resource.
func (r *recordResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manage a record.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Identifier of the record.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the record.",
			},
			"target": schema.StringAttribute{
				Required:    true,
				Description: "Target of the record.",
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (r *recordResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Add a nil check when handling ProviderData because Terraform
	// sets that data after it calls the ConfigureProvider RPC.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*usgdns.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *usgdns.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

// ImportState imports the resource and sets the Terraform state.
func (r *recordResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Create creates the resource and sets the initial Terraform state.
func (r *recordResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan recordResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	record, err := r.client.CreateRecord(plan.Name.ValueString(), plan.Target.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create the usg-dns record",
			err.Error(),
		)
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan.ID = types.StringValue(record.ID)
	plan.Name = types.StringValue(record.Name)
	plan.Target = types.StringValue(record.Target)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *recordResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state recordResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed record value from usg-dns
	record, err := r.client.GetRecord(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading usg-dns record",
			"Could not read usg-dns record ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	// Overwrite items with refreshed state
	state.Name = types.StringValue(record.Name)
	state.Target = types.StringValue(record.Target)

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *recordResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state recordResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var plan recordResourceModel
	diags = req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "plan:", map[string]any{"plan": state})

	// Update existing record
	record, err := r.client.UpdateRecord(state.ID.ValueString(), plan.Name.ValueString(), plan.Target.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating usg-dns record",
			"Could not update record, unexpected error: "+err.Error(),
		)
		return
	}

	// Update resource state with updated items and timestamp
	plan.ID = types.StringValue(record.ID)
	plan.Name = types.StringValue(record.Name)
	plan.Target = types.StringValue(record.Target)

	// Set refreshed state
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *recordResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state recordResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete existing record
	err := r.client.DeleteRecord(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting usg-dns record",
			"Could not delete record, unexpected error: "+err.Error(),
		)
		return
	}
}
