// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-usgdns/internal/usgdns"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &recordsDataSource{}
	_ datasource.DataSourceWithConfigure = &recordsDataSource{}
)

// recordsDataSourceModel maps the data source schema data.
type recordsDataSourceModel struct {
	Records []recordResourceModel `tfsdk:"records"`
}

func NewRecordsDataSource() datasource.DataSource {
	return &recordsDataSource{}
}

type recordsDataSource struct {
	client *usgdns.Client
}

func (d *recordsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_records"
}

func (d *recordsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetch the list of records.",
		Attributes: map[string]schema.Attribute{
			"records": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "Identifier of the record.",
						},
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "Name of the record.",
						},
						"target": schema.StringAttribute{
							Computed:    true,
							Description: "Target of the record.",
						},
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *recordsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	d.client = client
}

func (d *recordsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state recordsDataSourceModel

	records, err := d.client.GetRecords()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to fetch the usg-dns records",
			err.Error(),
		)
		return
	}

	// Map response body to model
	for _, record := range records {
		recordState := recordResourceModel{
			ID:     types.StringValue(record.ID),
			Name:   types.StringValue(record.Name),
			Target: types.StringValue(record.Target),
		}
		state.Records = append(state.Records, recordState)
	}

	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
