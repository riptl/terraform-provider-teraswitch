// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/teraswitch/terraform-provider-teraswitch/internal/tsw"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &SshKeyResource{}
var _ resource.ResourceWithImportState = &SshKeyResource{}

func NewSshKeyResource() resource.Resource {
	return &SshKeyResource{}
}

type SshKeyResource struct {
	client *tsw.Client
}

type SshKeyModel struct {
	Id          types.Int64  `tfsdk:"id"`
	ProjectId   types.Int64  `tfsdk:"project_id"`
	SshKey      types.String `tfsdk:"ssh_key"`
	DisplayName types.String `tfsdk:"display_name"`
}

func (s *SshKeyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ssh_key"
}

func (s *SshKeyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Creates and manages SSH keys for TeraSwitch servers.",

		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "The ID of the SSH key",
			},
			"project_id": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "The ID of the project the SSH key belongs to",
			},
			"ssh_key": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The OpenSSH format SSH public key",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"display_name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (s *SshKeyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*tsw.Client)

	if !ok {
		resp.Diagnostics.AddError("Client Error", "Unable to configure provider")
		return
	}

	s.client = client
}

func (s *SshKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data SshKeyModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	params := tsw.SshKeyCreateRequest{
		DisplayName: data.DisplayName.ValueString(),
		SshKey:      data.SshKey.ValueString(),
	}
	key, err := s.client.CreateSshKey(ctx, &params)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create SSH key, got error: %s", err))
		return
	}

	data.copyFromApi(key)

	tflog.Trace(ctx, "created SSH key")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (s *SshKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SshKeyModel

	// Read Terraform state into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	key, err := s.client.GetSshKey(ctx, data.Id.ValueInt64())
	if errors.Is(err, tsw.ErrNotFound) {
		req.State.RemoveResource(ctx)
	} else if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get SSH key, got error: %s", err))
		return
	}

	// Save updated data into Terraform state
	data.copyFromApi(key)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (s *SshKeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Provider Error", "Sorry, support for updating SSH keys is not yet implemented")
}

func (s *SshKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data SshKeyModel

	// Read Terraform state into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := s.client.DeleteSshKey(ctx, data.Id.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete SSH key, got error: %s", err))
		return
	}

	resp.State.RemoveResource(ctx)
}

func (s *SshKeyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idInt, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID", "ID should be numeric")
		return
	}

	key, err := s.client.GetSshKey(ctx, idInt)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get SSH key, got error: %s", err))
		return
	}

	// Save updated data into Terraform state
	var data SshKeyModel
	data.copyFromApi(key)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (m *SshKeyModel) copyFromApi(key *tsw.SshKey) {
	m.Id = types.Int64Value(key.Id)
	m.ProjectId = types.Int64Value(key.ProjectId)
	m.DisplayName = types.StringValue(key.DisplayName)
	m.SshKey = types.StringValue(key.SshKey)
}
