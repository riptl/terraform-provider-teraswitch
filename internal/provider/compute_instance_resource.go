// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/teraswitch/terraform-provider-teraswitch/internal/tsw"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &ComputeInstanceResource{}
var _ resource.ResourceWithImportState = &ComputeInstanceResource{}

func NewComputeInstanceResource() resource.Resource {
	return &ComputeInstanceResource{}
}

type ComputeInstanceResource struct {
	client *tsw.Client
}

type ComputeInstanceModel struct {
	Id          types.Int64  `tfsdk:"id"`
	ProjectId   types.Int64  `tfsdk:"project_id"`
	DisplayName types.String `tfsdk:"display_name"`
	Region      types.String `tfsdk:"region"`
	TierId      types.String `tfsdk:"tier_id"`
	ImageId     types.String `tfsdk:"image_id"`
	Tags        types.List   `tfsdk:"tags"`
	IpAddresses types.List   `tfsdk:"ip_addresses"`
	SshKeyIds   types.List   `tfsdk:"ssh_key_ids"`
	BootSize    types.Int64  `tfsdk:"boot_size"`
}

func (c *ComputeInstanceResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_compute_instance"
}

func (c *ComputeInstanceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Creates and manages TeraSwitch Cloud Compute servers.",

		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "The ID of the server",
			},
			"project_id": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "The ID of the project the server belongs to",
			},
			"display_name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The display name of the server",
			},
			"region": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The region the server is located in",
			},
			"tier_id": schema.StringAttribute{
				Required: true,
			},
			"image_id": schema.StringAttribute{
				Required: true,
			},
			"tags": schema.ListAttribute{
				Optional:    true,
				ElementType: basetypes.StringType{},
			},
			"ip_addresses": schema.ListAttribute{
				Computed:            true,
				MarkdownDescription: "The IP addresses assigned to the server",
				ElementType:         basetypes.StringType{},
			},
			"ssh_key_ids": schema.ListAttribute{
				Required:    true,
				ElementType: basetypes.Int64Type{},
			},
			"boot_size": schema.Int64Attribute{
				Required:            true,
				MarkdownDescription: "The size of the boot volume in GB",
			},
		},
	}
}

func (c *ComputeInstanceResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*tsw.Client)

	if !ok {
		resp.Diagnostics.AddError("Client Error", "Unable to configure provider")
		return
	}

	c.client = client
}

func (c *ComputeInstanceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ComputeInstanceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	params := tsw.InstanceCreateRequest{
		DisplayName: data.DisplayName.ValueString(),
		RegionId:    data.Region.ValueString(),
		TierId:      data.TierId.ValueString(),
		ImageId:     data.ImageId.ValueString(),
		BootSize:    int(data.BootSize.ValueInt64()),
	}
	resp.Diagnostics.Append(data.SshKeyIds.ElementsAs(context.Background(), &params.SshKeyIds, false)...)
	resp.Diagnostics.Append(data.Tags.ElementsAs(context.Background(), &params.Tags, false)...)

	if resp.Diagnostics.HasError() {
		return
	}

	instance, err := c.client.CreateInstance(ctx, &params)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create instance, got error: %s", err))
		return
	}

	data.copyFromApi(instance, resp.Diagnostics)

	tflog.Trace(ctx, "sent instance creation request, polling ...")

	// TODO add configurable polling interval
	ticker := time.NewTicker(1 * time.Second)
waitForStartup:
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			instance, err := c.client.GetInstance(ctx, data.Id.ValueInt64())
			if err != nil {
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get instance, got error: %s", err))
				return
			}
			if instance.PowerState == tsw.PowerStateOn {
				tflog.Trace(ctx, "instance is reporting power state on")
				break waitForStartup
			}
		}
	}

	tflog.Trace(ctx, "created instance")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (c *ComputeInstanceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ComputeInstanceModel

	// Read Terraform state into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	instance, err := c.client.GetInstance(ctx, data.Id.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get instance, got error: %s", err))
		return
	}

	// Save updated data into Terraform state
	data.copyFromApi(instance, resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (c *ComputeInstanceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	//TODO implement me
	panic("implement me")
}

func (c *ComputeInstanceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	//TODO implement me
	panic("implement me")
}

func (c *ComputeInstanceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idInt, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID", "ID should be numeric")
		return
	}

	instance, err := c.client.GetInstance(ctx, idInt)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get compute instance, got error: %s", err))
		return
	}

	// Save updated data into Terraform state
	var data ComputeInstanceModel
	data.copyFromApi(instance, resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (m *ComputeInstanceModel) copyFromApi(instance *tsw.Instance, diags diag.Diagnostics) {
	var nextDiag diag.Diagnostics

	m.Id = types.Int64Value(instance.Id)

	m.ProjectId = types.Int64Value(instance.ProjectId)

	m.Region = types.StringValue(instance.RegionId)

	m.TierId = types.StringValue(instance.TierId)

	m.ImageId = types.StringValue(instance.ImageId)

	ipAddrs := make([]attr.Value, len(instance.IpAddresses))
	for i, ip := range instance.IpAddresses {
		ipAddrs[i] = basetypes.NewStringValue(ip)
	}
	m.IpAddresses, nextDiag = types.ListValue(basetypes.StringType{}, ipAddrs)
	diags.Append(nextDiag...)
}
