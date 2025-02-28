package mdbcommon

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/datasize"
	utils "github.com/yandex-cloud/terraform-provider-yandex/pkg/wrappers"
	"google.golang.org/genproto/googleapis/type/timeofday"
)

var (
	baseOptions = basetypes.ObjectAsOptions{UnhandledNullAsEmpty: false, UnhandledUnknownAsEmpty: false}
)

type BackupWindow struct {
	Hours   types.Int64 `tfsdk:"hours"`
	Minutes types.Int64 `tfsdk:"minutes"`
}

var BackupWindowType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"hours":   types.Int64Type,
		"minutes": types.Int64Type,
	},
}

type Resource struct {
	ResourcePresetId types.String `tfsdk:"resource_preset_id"`
	DiskSize         types.Int64  `tfsdk:"disk_size"`
	DiskTypeId       types.String `tfsdk:"disk_type_id"`
}

type resourceModel[T any] interface {
	SetResourcePresetId(string)
	SetDiskSize(int64)
	SetDiskTypeId(string)

	GetResourcePresetId() string
	GetDiskSize() int64
	GetDiskTypeId() string
	*T
}

var ResourceType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"resource_preset_id": types.StringType,
		"disk_type_id":       types.StringType,
		"disk_size":          types.Int64Type,
	},
}

func FlattenBackupWindow(ctx context.Context, bw *timeofday.TimeOfDay) (types.Object, diag.Diagnostics) {
	if bw == nil {
		return types.ObjectNull(BackupWindowType.AttributeTypes()), nil
	}

	result := BackupWindow{
		Hours:   types.Int64Value(int64(bw.GetHours())),
		Minutes: types.Int64Value(int64(bw.GetMinutes())),
	}

	return types.ObjectValueFrom(ctx, BackupWindowType.AttributeTypes(), result)
}

func ExpandBackupWindow(ctx context.Context, bw types.Object) (*timeofday.TimeOfDay, diag.Diagnostics) {
	if !utils.IsPresent(bw) {
		return nil, nil
	}
	backupWindow := &BackupWindow{}
	diags := bw.As(ctx, backupWindow, baseOptions)
	if diags.HasError() {
		return nil, diags
	}
	rs := &timeofday.TimeOfDay{
		Hours:   int32(backupWindow.Hours.ValueInt64()),
		Minutes: int32(backupWindow.Minutes.ValueInt64()),
	}
	return rs, diags
}

func ExpandResources[V any, T resourceModel[V]](ctx context.Context, o types.Object) (T, diag.Diagnostics) {
	if !utils.IsPresent(o) {
		return nil, nil
	}
	d := &Resource{}
	diags := o.As(ctx, d, baseOptions)
	if diags.HasError() {
		return nil, diags
	}
	rs := T(new(V))
	rs.SetResourcePresetId(d.ResourcePresetId.ValueString())
	rs.SetDiskSize(datasize.ToBytes(d.DiskSize.ValueInt64()))
	if utils.IsPresent(d.DiskTypeId) {
		rs.SetDiskTypeId(d.DiskTypeId.ValueString())
	}

	return rs, diags
}

func FlattenResources[V any, T resourceModel[V]](ctx context.Context, r T) (types.Object, diag.Diagnostics) {
	if r == nil {
		return types.ObjectNull(ResourceType.AttributeTypes()), nil
	}

	a := Resource{
		ResourcePresetId: types.StringValue(r.GetResourcePresetId()),
		DiskSize:         types.Int64Value(datasize.ToGigabytes(r.GetDiskSize())),
		DiskTypeId:       types.StringValue(r.GetDiskTypeId()),
	}
	return types.ObjectValueFrom(ctx, ResourceType.AttributeTypes(), a)
}
