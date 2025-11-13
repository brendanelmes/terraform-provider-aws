// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package applicationsignals

import (
	"context"
	"errors"
	"reflect"
	"time"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/applicationsignals"
	awstypes "github.com/aws/aws-sdk-go-v2/service/applicationsignals/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_applicationsignals_service_level_objective", name="Service Level Objective")
func newResourceServiceLevelObjective(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceServiceLevelObjective{}

	// TIP: ==== CONFIGURABLE TIMEOUTS ====
	// Users can configure timeout lengths but you need to use the times they
	// provide. Access the timeout they configure (or the defaults) using,
	// e.g., r.CreateTimeout(ctx, plan.Timeouts) (see below). The times here are
	// the defaults if they don't configure timeouts.
	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameServiceLevelObjective = "Service Level Objective"
)

type resourceServiceLevelObjective struct {
	framework.ResourceWithModel[resourceServiceLevelObjectiveModel]
	framework.WithTimeouts
}

// TIP: ==== SCHEMA ====
// In the schema, add each of the attributes in snake case (e.g.,
// delete_automated_backups).
//
// Formatting rules:
// * Alphabetize attributes to make them easier to find.
// * Do not add a blank line between attributes.
//
// Attribute basics:
//   - If a user can provide a value ("configure a value") for an
//     attribute (e.g., instances = 5), we call the attribute an
//     "argument."
//   - You change the way users interact with attributes using:
//   - Required
//   - Optional
//   - Computed
//   - There are only four valid combinations:
//
// 1. Required only - the user must provide a value
// Required: true,
//
//  2. Optional only - the user can configure or omit a value; do not
//     use Default or DefaultFunc
//
// Optional: true,
//
//  3. Computed only - the provider can provide a value but the user
//     cannot, i.e., read-only
//
// Computed: true,
//
//  4. Optional AND Computed - the provider or user can provide a value;
//     use this combination if you are using Default
//
// Optional: true,
// Computed: true,
//
// You will typically find arguments in the input struct
// (e.g., CreateDBInstanceInput) for the create operation. Sometimes
// they are only in the input struct (e.g., ModifyDBInstanceInput) for
// the modify operation.
//
// For more about schema options, visit
// https://developer.hashicorp.com/terraform/plugin/framework/handling-data/schemas?page=schemas
func (r *resourceServiceLevelObjective) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrCreatedTime: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
			},
			"evaluation_type": schema.StringAttribute{
				Computed: true,
			},
			"last_updated_time": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			"metric_source_type": schema.StringAttribute{
				Computed: true,
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
			},
		},
		Blocks: map[string]schema.Block{
			"goal": schema.SingleNestedBlock{
				CustomType: fwtypes.NewObjectTypeOf[goalModel](ctx),
				Attributes: map[string]schema.Attribute{
					"attainment_goal":   schema.Float64Attribute{Required: true},
					"warning_threshold": schema.Float64Attribute{Required: true},
				},
				Validators: []validator.Object{
					objectvalidator.IsRequired(),
				},
				Blocks: map[string]schema.Block{
					"interval": schema.SingleNestedBlock{
						CustomType: fwtypes.NewObjectTypeOf[intervalModel](ctx),
						Validators: []validator.Object{
							objectvalidator.IsRequired(),
						},
						Blocks: map[string]schema.Block{
							"calendar_interval": schema.SingleNestedBlock{
								CustomType: fwtypes.NewObjectTypeOf[calendarIntervalModel](ctx),
								Validators: []validator.Object{
									objectvalidator.ExactlyOneOf(
										path.Expressions{
											path.MatchRelative().AtParent().AtName("rolling_interval"),
										}...),
								},
								Attributes: map[string]schema.Attribute{
									"duration":      schema.Int32Attribute{Optional: true},
									"duration_unit": schema.StringAttribute{Optional: true},
									"start_time":    schema.StringAttribute{Optional: true},
								},
							},
							"rolling_interval": schema.SingleNestedBlock{
								CustomType: fwtypes.NewObjectTypeOf[rollingIntervalModel](ctx),
								Validators: []validator.Object{
									objectvalidator.ExactlyOneOf(
										path.Expressions{
											path.MatchRelative().AtParent().AtName("calendar_interval"),
										}...),
								},
								Attributes: map[string]schema.Attribute{
									"duration":      schema.Int32Attribute{Optional: true},
									"duration_unit": schema.StringAttribute{Optional: true},
								},
							},
						},
					},
				},
			},
			"burn_rate_configurations": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[burnRateConfigurationModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"look_back_window_minutes": schema.Int32Attribute{Optional: true},
					},
				},
			},
			"request_based_sli": schema.SingleNestedBlock{
				CustomType: fwtypes.NewObjectTypeOf[requestBasedSliModel](ctx),
				Validators: []validator.Object{
					objectvalidator.ExactlyOneOf(
						path.Expressions{
							path.MatchRelative().AtParent().AtName("sli"),
						}...),
				},
				Attributes: map[string]schema.Attribute{
					"metric_threshold":    schema.Float64Attribute{Optional: true},
					"comparison_operator": schema.StringAttribute{Optional: true},
				},
				Blocks: map[string]schema.Block{
					"request_based_sli_metric": schema.SingleNestedBlock{
						CustomType: fwtypes.NewObjectTypeOf[requestBasedSliMetricModel](ctx),
						Attributes: map[string]schema.Attribute{
							"key_attributes": schema.MapAttribute{CustomType: fwtypes.MapOfStringType, ElementType: types.StringType, Optional: true},
							"metric_type":    schema.StringAttribute{Optional: true},
							"operation_name": schema.StringAttribute{Optional: true},
						},
						Blocks: map[string]schema.Block{
							"total_request_count_metric": metricDataQueriesBlock(ctx),
							"dependency_config": schema.SingleNestedBlock{
								CustomType: fwtypes.NewObjectTypeOf[dependencyConfigModel](ctx),
							},
							// TODO - Implement MetricDataQuery
						},
					},
				},
			},
			"sli": schema.SingleNestedBlock{
				CustomType: fwtypes.NewObjectTypeOf[sliModel](ctx),
				Validators: []validator.Object{
					objectvalidator.ExactlyOneOf(
						path.Expressions{
							path.MatchRelative().AtParent().AtName("request_based_sli"),
						}...),
				},
				Attributes: map[string]schema.Attribute{
					"metric_threshold":    schema.Float64Attribute{Optional: true},
					"comparison_operator": schema.StringAttribute{Optional: true},
				},
				Blocks: map[string]schema.Block{
					"sli_metric": schema.SingleNestedBlock{
						CustomType: fwtypes.NewObjectTypeOf[sliMetricModel](ctx),
						Attributes: map[string]schema.Attribute{
							"key_attributes": schema.MapAttribute{CustomType: fwtypes.MapOfStringType, ElementType: types.StringType, Optional: true},
							"metric_type":    schema.StringAttribute{Optional: true},
							"metric_name":    schema.StringAttribute{Optional: true},
							"operation_name": schema.StringAttribute{Optional: true},
							"period_seconds": schema.Int32Attribute{Optional: true},
							"statistic":      schema.StringAttribute{Optional: true},
						},
						Blocks: map[string]schema.Block{
							"metric_data_queries": metricDataQueriesBlock(ctx),
							"dependency_config": schema.SingleNestedBlock{
								CustomType: fwtypes.NewObjectTypeOf[dependencyConfigModel](ctx),
							},
						},
					},
				},
			},
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func metricDataQueriesBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[metricDataQueryModel](ctx),
		NestedObject: schema.NestedBlockObject{
			CustomType: fwtypes.NewObjectTypeOf[metricDataQueryModel](ctx),
			Attributes: map[string]schema.Attribute{
				"id":          schema.StringAttribute{Optional: true},
				"account_id":  schema.StringAttribute{Computed: true},
				"expression":  schema.StringAttribute{Optional: true},
				"label":       schema.StringAttribute{Optional: true},
				"period":      schema.Int32Attribute{Optional: true},
				"return_data": schema.BoolAttribute{Optional: true},
			},
			Blocks: map[string]schema.Block{
				"metric_stat": schema.SingleNestedBlock{
					CustomType: fwtypes.NewObjectTypeOf[metricStatModel](ctx),
					Attributes: map[string]schema.Attribute{
						"period": schema.Int32Attribute{Optional: true},
						"stat":   schema.StringAttribute{Optional: true},
						"unit":   schema.StringAttribute{Optional: true},
					},
					Blocks: map[string]schema.Block{
						"metric": schema.SingleNestedBlock{
							CustomType: fwtypes.NewObjectTypeOf[metricModel](ctx),
							Attributes: map[string]schema.Attribute{
								"metric_name": schema.StringAttribute{Optional: true},
								"namespace":   schema.StringAttribute{Optional: true},
							},
							Blocks: map[string]schema.Block{
								"dimensions": schema.ListNestedBlock{
									CustomType: fwtypes.NewListNestedObjectTypeOf[dimensionModel](ctx),
									NestedObject: schema.NestedBlockObject{
										CustomType: fwtypes.NewObjectTypeOf[dimensionModel](ctx),
										Attributes: map[string]schema.Attribute{
											"name":  schema.StringAttribute{Computed: true},
											"value": schema.StringAttribute{Computed: true},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (r *resourceServiceLevelObjective) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().ApplicationSignalsClient(ctx)

	var plan resourceServiceLevelObjectiveModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	var input applicationsignals.CreateServiceLevelObjectiveInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := conn.CreateServiceLevelObjective(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Name.String())
		return
	}
	if out == nil || out.Slo == nil {
		smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, plan.Name.String())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out.Slo, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *resourceServiceLevelObjective) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().ApplicationSignalsClient(ctx)

	var state resourceServiceLevelObjectiveModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findServiceLevelObjectiveByID(ctx, conn, state.Name.ValueString())
	if tfresource.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.Name.String())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *resourceServiceLevelObjective) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().ApplicationSignalsClient(ctx)

	var plan, state resourceServiceLevelObjectiveModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	diff, d := flex.Diff(ctx, plan, state)
	smerr.AddEnrich(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		var input applicationsignals.UpdateServiceLevelObjectiveInput
		smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input, flex.WithFieldNamePrefix("Test")))
		if resp.Diagnostics.HasError() {
			return
		}

		out, err := conn.UpdateServiceLevelObjective(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Name.String())
			return
		}
		if out == nil || out.Slo == nil {
			smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, plan.Name.String())
			return
		}

		smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out.Slo, &plan))
		if resp.Diagnostics.HasError() {
			return
		}
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *resourceServiceLevelObjective) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().ApplicationSignalsClient(ctx)

	var state resourceServiceLevelObjectiveModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	input := applicationsignals.DeleteServiceLevelObjectiveInput{
		Id: state.Name.ValueStringPointer(),
	}

	_, err := conn.DeleteServiceLevelObjective(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}

		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.Name.String())
		return
	}
}

// TIP: ==== TERRAFORM IMPORTING ====
// If Read can get all the information it needs from the Identifier
// (i.e., path.Root("id")), you can use the PassthroughID importer. Otherwise,
// you'll need a custom import function.
//
// See more:
// https://developer.hashicorp.com/terraform/plugin/framework/resources/import
func (r *resourceServiceLevelObjective) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrID), req, resp)
}

func findServiceLevelObjectiveByID(ctx context.Context, conn *applicationsignals.Client, name string) (*awstypes.ServiceLevelObjective, error) {
	input := applicationsignals.GetServiceLevelObjectiveInput{
		Id: aws.String(name),
	}

	out, err := conn.GetServiceLevelObjective(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, smarterr.NewError(&retry.NotFoundError{
				LastError:   err,
				LastRequest: &input,
			})
		}

		return nil, smarterr.NewError(err)
	}

	if out == nil || out.Slo == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError(&input))
	}

	return out.Slo, nil
}

var _ flex.Flattener = &intervalModel{}

func (m *intervalModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics

	m.CalendarInterval = fwtypes.NewObjectValueOfNull[calendarIntervalModel](ctx)
	m.RollingInterval = fwtypes.NewObjectValueOfNull[rollingIntervalModel](ctx)

	switch t := v.(type) {

	case awstypes.IntervalMemberCalendarInterval:
		var model calendarIntervalModel
		diags.Append(flex.Flatten(ctx, t.Value, &model)...)
		if !diags.HasError() {
			m.CalendarInterval = fwtypes.NewObjectValueOfMust(ctx, &model)
		}

	case awstypes.IntervalMemberRollingInterval:
		var model rollingIntervalModel
		diags.Append(flex.Flatten(ctx, t.Value, &model)...)
		if !diags.HasError() {
			m.RollingInterval = fwtypes.NewObjectValueOfMust(ctx, &model)
		}
	}

	return diags
}

func stringPtr(v types.String) *string {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	val := v.ValueString()
	return &val
}

func (m resourceServiceLevelObjectiveModel) ExpandTo(ctx context.Context, targetType reflect.Type) (result any, diags diag.Diagnostics) {
	switch targetType {
	case reflect.TypeFor[applicationsignals.UpdateServiceLevelObjectiveInput]():
		return m.expandToUpdateServiceLevelObjectiveInput(ctx)

	case reflect.TypeFor[applicationsignals.CreateServiceLevelObjectiveInput]():
		return m.expandToCreateServiceLevelObjectiveInput(ctx)
	}
	return nil, diags
}

func (m resourceServiceLevelObjectiveModel) expandToUpdateServiceLevelObjectiveInput(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	input := &applicationsignals.UpdateServiceLevelObjectiveInput{}

	input.Id = stringPtr(m.Name)
	input.Description = stringPtr(m.Description)

	if !m.Goal.IsNull() {
		goalData, d := m.Goal.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var goal awstypes.Goal
		diags.Append(flex.Expand(ctx, goalData, &goal)...)
		if diags.HasError() {
			return nil, diags
		}
		input.Goal = &goal
	}

	if !m.BurnRateConfigurations.IsNull() {
		burnData, d := m.BurnRateConfigurations.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var burns []awstypes.BurnRateConfiguration
		diags.Append(flex.Expand(ctx, burnData, &burns)...)
		if diags.HasError() {
			return nil, diags
		}
		input.BurnRateConfigurations = burns
	}

	if !m.Sli.IsNull() {
		sliData, d := m.Sli.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var sli awstypes.ServiceLevelIndicatorConfig
		diags.Append(flex.Expand(ctx, sliData, &sli)...)
		if diags.HasError() {
			return nil, diags
		}
		input.SliConfig = &sli
	}

	if !m.RequestBasedSli.IsNull() {
		reqSliData, d := m.RequestBasedSli.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var reqSli awstypes.RequestBasedServiceLevelIndicatorConfig
		diags.Append(flex.Expand(ctx, reqSliData, &reqSli)...)
		if diags.HasError() {
			return nil, diags
		}
		input.RequestBasedSliConfig = &reqSli
	}

	return input, diags
}

func (m resourceServiceLevelObjectiveModel) expandToCreateServiceLevelObjectiveInput(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	input := &applicationsignals.CreateServiceLevelObjectiveInput{}

	input.Name = stringPtr(m.Name)
	input.Description = stringPtr(m.Description)

	if !m.Goal.IsNull() {
		goalData, d := m.Goal.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var goal awstypes.Goal
		diags.Append(flex.Expand(ctx, goalData, &goal)...)
		if diags.HasError() {
			return nil, diags
		}
		input.Goal = &goal
	}

	if !m.BurnRateConfigurations.IsNull() {
		burnData, d := m.BurnRateConfigurations.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var burns []awstypes.BurnRateConfiguration
		diags.Append(flex.Expand(ctx, burnData, &burns)...)
		if diags.HasError() {
			return nil, diags
		}
		input.BurnRateConfigurations = burns
	}

	if !m.Sli.IsNull() {
		sliData, d := m.Sli.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var sli awstypes.ServiceLevelIndicatorConfig
		diags.Append(flex.Expand(ctx, sliData, &sli)...)
		if diags.HasError() {
			return nil, diags
		}
		input.SliConfig = &sli
	}

	if !m.RequestBasedSli.IsNull() {
		reqSliData, d := m.RequestBasedSli.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var reqSli awstypes.RequestBasedServiceLevelIndicatorConfig
		diags.Append(flex.Expand(ctx, reqSliData, &reqSli)...)
		if diags.HasError() {
			return nil, diags
		}
		input.RequestBasedSliConfig = &reqSli
	}

	return input, diags
}

func (m sliModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics

	var config awstypes.ServiceLevelIndicatorConfig

	if !m.ComparisonOperator.IsNull() {
		config.ComparisonOperator = awstypes.ServiceLevelIndicatorComparisonOperator(m.ComparisonOperator.ValueString())
	}

	if !m.MetricThreshold.IsNull() {
		val := m.MetricThreshold.ValueFloat64()
		config.MetricThreshold = &val
	}

	if !m.SliMetric.IsNull() {
		sliMetricData, d := m.SliMetric.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var metric awstypes.ServiceLevelIndicatorMetricConfig
		diags.Append(flex.Expand(ctx, sliMetricData, &metric)...)
		if diags.HasError() {
			return nil, diags
		}

		config.SliMetricConfig = &metric
	}

	return &config, diags
}

//func (m requestBasedSliMetricModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
//
//}

func (m intervalModel) Expand(ctx context.Context) (result any, diags diag.Diagnostics) {
	switch {
	case !m.RollingInterval.IsNull():
		rollingData, d := m.RollingInterval.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.IntervalMemberRollingInterval
		diags.Append(flex.Expand(ctx, rollingData, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}

		return &r, diags

	case !m.CalendarInterval.IsNull():
		calendarData, d := m.CalendarInterval.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.IntervalMemberCalendarInterval
		diags.Append(flex.Expand(ctx, calendarData, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}

		return &r, diags
	}

	return nil, diags
}

type resourceServiceLevelObjectiveModel struct {
	framework.WithRegionModel
	ARN                    types.String                                                `tfsdk:"arn"`
	CreatedTime            timetypes.RFC3339                                           `tfsdk:"created_time"`
	BurnRateConfigurations fwtypes.ListNestedObjectValueOf[burnRateConfigurationModel] `tfsdk:"burn_rate_configurations"`
	LastUpdatedTime        timetypes.RFC3339                                           `tfsdk:"last_updated_time"`
	Name                   types.String                                                `tfsdk:"name"`
	Description            types.String                                                `tfsdk:"description"`
	MetricSourceType       types.String                                                `tfsdk:"metric_source_type"`
	EvaluationType         types.String                                                `tfsdk:"evaluation_type"`
	Goal                   fwtypes.ObjectValueOf[goalModel]                            `tfsdk:"goal"`
	Sli                    fwtypes.ObjectValueOf[sliModel]                             `tfsdk:"sli"`
	RequestBasedSli        fwtypes.ObjectValueOf[requestBasedSliModel]                 `tfsdk:"request_based_sli"`
	Timeouts               timeouts.Value                                              `tfsdk:"timeouts"`
}

type goalModel struct {
	AttainmentGoal   types.Float64                        `tfsdk:"attainment_goal"`
	WarningThreshold types.Float64                        `tfsdk:"warning_threshold"`
	Interval         fwtypes.ObjectValueOf[intervalModel] `tfsdk:"interval"`
}

type intervalModel struct {
	CalendarInterval fwtypes.ObjectValueOf[calendarIntervalModel] `tfsdk:"calendar_interval"`
	RollingInterval  fwtypes.ObjectValueOf[rollingIntervalModel]  `tfsdk:"rolling_interval"`
}

type calendarIntervalModel struct {
	Duration     types.Int32  `tfsdk:"duration"`
	DurationUnit types.String `tfsdk:"duration_unit"`
	StartTime    types.String `tfsdk:"start_time"`
}

type rollingIntervalModel struct {
	Duration     types.Int32  `tfsdk:"duration"`
	DurationUnit types.String `tfsdk:"duration_unit"`
}

type sliModel struct {
	ComparisonOperator types.String                          `tfsdk:"comparison_operator"`
	MetricThreshold    types.Float64                         `tfsdk:"metric_threshold"`
	SliMetric          fwtypes.ObjectValueOf[sliMetricModel] `tfsdk:"sli_metric"`
}

type requestBasedSliModel struct {
	RequestBasedSliMetric fwtypes.ObjectValueOf[requestBasedSliMetricModel] `tfsdk:"request_based_sli_metric"`
	ComparisonOperator    types.String                                      `tfsdk:"comparison_operator"`
	MetricThreshold       types.Float64                                     `tfsdk:"metric_threshold"`
}

type burnRateConfigurationModel struct {
	LookBackWindowMinutes types.Int32 `tfsdk:"look_back_window_minutes"`
}

type requestBasedSliMetricModel struct {
	TotalRequestCountMetric fwtypes.ListNestedObjectValueOf[metricDataQueryModel] `tfsdk:"total_request_count_metric"`
	DependencyConfig        fwtypes.ObjectValueOf[dependencyConfigModel]          `tfsdk:"dependency_config"`
	KeyAttributes           fwtypes.MapOfString                                   `tfsdk:"key_attributes"`
	MetricType              types.String                                          `tfsdk:"metric_type"`
	OperationName           types.String                                          `tfsdk:"operation_name"`
}

type sliMetricModel struct {
	MetricDataQueries fwtypes.ListNestedObjectValueOf[metricDataQueryModel] `tfsdk:"metric_data_queries"`
	DependencyConfig  fwtypes.ObjectValueOf[dependencyConfigModel]          `tfsdk:"dependency_config"`
	KeyAttributes     fwtypes.MapOfString                                   `tfsdk:"key_attributes"`
	MetricName        types.String                                          `tfsdk:"metric_name"`
	MetricType        types.String                                          `tfsdk:"metric_type"`
	OperationName     types.String                                          `tfsdk:"operation_name"`
	PeriodSeconds     types.Int32                                           `tfsdk:"period_seconds"`
	Statistic         types.String                                          `tfsdk:"statistic"`
}

type metricDataQueryModel struct {
	Id         types.String                           `tfsdk:"id"`
	AccountId  types.String                           `tfsdk:"account_id"`
	Expression types.String                           `tfsdk:"expression"`
	Label      types.String                           `tfsdk:"label"`
	MetricStat fwtypes.ObjectValueOf[metricStatModel] `tfsdk:"metric_stat"`
	Period     types.Int32                            `tfsdk:"period"`
	ReturnData types.Bool                             `tfsdk:"return_data"`
}

type metricStatModel struct {
	Metric fwtypes.ObjectValueOf[metricModel] `tfsdk:"metric"`
	Period types.Int32                        `tfsdk:"period"`
	Stat   types.String                       `tfsdk:"stat"`
	Unit   types.String                       `tfsdk:"unit"`
}

type metricModel struct {
	Dimensions fwtypes.ListNestedObjectValueOf[dimensionModel] `tfsdk:"dimensions"`
	MetricName types.String                                    `tfsdk:"metric_name"`
	Namespace  types.String                                    `tfsdk:"namespace"`
}

type dimensionModel struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}

type dependencyConfigModel struct {
	DependencyKeyAttributes types.String `tfsdk:"dependency_key_attributes"`
	DependencyOperationName types.String `tfsdk:"dependency_operation_name"`
}

// TIP: ==== SWEEPERS ====
// When acceptance testing resources, interrupted or failed tests may
// leave behind orphaned resources in an account. To facilitate cleaning
// up lingering resources, each resource implementation should include
// a corresponding "sweeper" function.
//
// The sweeper function lists all resources of a given type and sets the
// appropriate identifers required to delete the resource via the Delete
// method implemented above.
//
// Once the sweeper function is implemented, register it in sweep.go
// as follows:
//
//	awsv2.Register("aws_applicationsignals_service_level_objective", sweepServiceLevelObjectives)
//
// See more:
// https://hashicorp.github.io/terraform-provider-aws/running-and-writing-acceptance-tests/#acceptance-test-sweepers
//func sweepServiceLevelObjectives(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
//	input := applicationsignals.ListServiceLevelObjectivesInput{}
//	conn := client.ApplicationSignalsClient(ctx)
//	var sweepResources []sweep.Sweepable
//
//	pages := applicationsignals.NewListServiceLevelObjectivesPaginator(conn, &input)
//	for pages.HasMorePages() {
//		page, err := pages.NextPage(ctx)
//		if err != nil {
//			return nil, smarterr.NewError(err)
//		}
//
//		for _, v := range page.Slos {
//			sweepResources = append(sweepResources, sweepfw.NewSweepResource(newResourceServiceLevelObjective, client,
//				sweepfw.NewAttribute(names.AttrID, aws.ToString(v.ServiceLevelObjectiveId))),
//			)
//		}
//	}
//
//	return sweepResources, nil
//}
