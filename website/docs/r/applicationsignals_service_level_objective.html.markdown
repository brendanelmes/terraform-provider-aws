---
subcategory: "Application Signals"
layout: "aws"
page_title: "AWS: aws_applicationsignals_service_level_objective"
description: |-
  Manages an AWS Application Signals Service Level Objective.
---

# Resource: aws_applicationsignals_service_level_objective

Manages an AWS Application Signals Service Level Objective.

## Example Usage

### Basic Usage with a Period-Based SLO

```terraform
resource "aws_applicationsignals_service_level_objective" "example" {
  name        = "elb-error-rate"
  description = "Error rate of 99.98% for 90 days"
  goal {
    interval {
      rolling_interval {
        duration_unit = "DAY"
        duration      = 90
      }
    }
    attainment_goal   = 99.98
    warning_threshold = 30.0
  }
  sli {
    comparison_operator = "LessThan"
    metric_threshold    = 2
    sli_metric {
      metric_data_queries {
        id = "m1"
        metric_stat {
          metric {
            namespace   = "AWS/ApplicationELB"
            metric_name = "HTTPCode_Target_5XX_Count"
            dimensions {
              name  = "LoadBalancer"
              value = "app/my-load-balancer"
            }
          }
          period = 300
          stat   = "Sum"
        }
        return_data = true
      }
    }
  }
}
```

### Request-Based SLO Usage

```terraform
resource "aws_applicationsignals_service_level_objective" "example" {
  name        = "lambda-success-rate"
  description = "Success rate of 99.9% for a specific operation over a calendar month"
  goal {
    interval {
      rolling_interval {
        duration      = 1
        duration_unit = "DAY"
      }
    }
    attainment_goal   = 99.90
    warning_threshold = 50.0
  }
  request_based_sli {
    request_based_sli_metric {
      total_request_count_metric {
        metric_stat {
          metric {
            namespace  = "AWS/Lambda"
            metric_name = "Invocations"
            dimensions {
              name  = "Dimension1"
              value = "my-dimension-name"
            }
          }
          period = 60
          stat = "Sum"
        }
        id = "total_requests"
        return_data = true
      }
      monitored_request_count_metric {
        bad_count_metric {
          metric_stat {
            metric {
              namespace   = "AWS/Lambda"
              metric_name = "ErrorCount"
            }
            period = 60
            stat   = "Sum"
          }
          id = "bad_requests"
          return_data = true
        }
      }
    }
  }
}
```

-----

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of this SLO. Must be unique for your AWS account and is immutable after creation.
* [`goal`](#goal) - (Required) Configuration block determining the goal of this SLO.

The following arguments are optional:

* `description` - (Optional) Brief description of the SLO.
* [`burn_rate_configurations`](#burn_rate_configurations) - (Optional) Configuration block containing attributes that determine the burn rate of this SLO.
* [`request_based_sli`](#request_based_sli) - (Optional) Configuration block for a request-based Service Level Indicator (SLI)
* [`sli`](#sli) - (Optional) Configuration block for a period-based Service Level Indicator (SLI).
* `timeouts` - (Optional) Configuration block for setting operation timeouts.

> You must specify exactly one `sli` or `request_based_sli`.

## Block Reference

### burn_rate_configurations

This argument is processed in [attribute-as-blocks mode](https://www.terraform.io/docs/configuration/attr-as-blocks.html).

The following arguments are required:

* `look_back_window_minutes` - (Required) The number of minutes to use as the look back window for calculating the burn rate.

### goal

The following arguments are supported:

* `attainment_goal` - (Required) The percentage of time in the interval that the service must satisfy the SLI to achieve the attainment goal.
* [`interval`](#interval) - (Required) Configuration block defining the time period over which the SLO is evaluated.
* `warning_threshold` - (Required) The percentage of the attainment goal that is allowed to elapse before the user receives a warning.

### interval

The `interval` block must contain exactly one of the following blocks:

* [`calendar_interval`](#calendar_interval) - Configuration block for a time interval that starts at a specific time and runs for a specified duration.
* [`rolling_interval`](#rolling_interval) - Configuration block for a time interval that rolls forward by a specified duration.

### calendar_interval

* `duration` - (Required) The duration of the calendar interval.
* `duration_unit` - (Required) The unit of time for the duration (`MINUTE`, `HOUR`, `DAY`, or `MONTH`).
* `start_time` - (Required) The start time of the first interval in **RFC3339** format (e.g., `2024-01-01T00:00:00Z`).

### rolling_interval

* `duration` - (Required) The duration of the rolling interval.
* `duration_unit` - (Required) The unit of time for the duration (`MINUTE`, `HOUR`, or `DAY`).

### sli

Use this block to define an SLO based on a single metric, typically for latency or error rate where a single metric is compared to a threshold.

* `comparison_operator` - (Optional) The arithmetic operation to use when comparing the SLI metric value to the `metric_threshold`.
* `metric_threshold` - (Optional) The value the SLI metric value is compared to.
* [`sli_metric`](#sli_metric) - (Optional) Configuration block defining the metric for this period-based SLI.

### sli_metric

* [`dependency_config`](#dependency_config) - (Optional) Configuration block for filtering metrics for a dependency.
* `key_attributes` - (Optional) A map of key-value pairs that are used to filter the application's metric. (Type: `map(string)`)
* [`metric_data_queries`](#metric_data_queries) - (Optional) Configuration block for a list of CloudWatch metric data queries.
* `metric_name` - (Optional) The name of the CloudWatch metric to use.
* `metric_type` - (Optional) The metric type for the SLI. Valid values include `Availability`, `Latency`, `Fault`, `RequestCount`.
* `operation_name` - (Optional) The name of the operation this SLO applies to.
* `period_seconds` - (Optional) The number of seconds to use as the period for the CloudWatch metric.
* `statistic` - (Optional) The statistic to use for the CloudWatch metric.

### request_based_sli

Use this block to define an SLO based on the ratio of good or bad requests to total requests.

* `comparison_operator` - (Optional) The arithmetic operation to use when comparing the success rate to the `metric_threshold`.
* `metric_threshold` - (Optional) The percentage success rate the comparison operator is compared to.
* [`request_based_sli_metric`](#request_based_sli_metric) - (Optional) Configuration block defining the metrics for this request-based SLI.

### request_based_sli_metric

* [`dependency_config`](#dependency_config) - (Optional) Configuration block for filtering metrics for a dependency.
* `key_attributes` - (Optional) A map of key-value pairs that are used to filter the application's metric.
* `metric_type` - (Optional) The metric type for the SLI. Currently only `RequestCount` is supported.
* [`monitored_request_count_metric`](#monitored_request_count_metric) - (Optional) Configuration block defining the good and bad request count metrics.
* `operation_name` - (Optional) The name of the operation this SLO applies to.
* [`total_request_count_metric`](#total_request_count_metric) - (Optional) Configuration block for the total request count metric, as a list of metric data queries.

### monitored_request_count_metric

This block defines the metrics for good and bad requests.

* [`good_count_metric`](#good_count_metric) - (Optional) Configuration block for the metric that counts good requests.
* [`bad_count_metric`](#bad_count_metric) - (Optional) Configuration block for the metric that counts bad requests.

### good_count_metric

A list of CloudWatch metric data queries. This is a **List Nested Block**.

* `account_id` - (Optional) The ID of the account to use for the metric data query.
* `expression` - (Optional) The math expression to use on the returned metric.
* `id` - (Optional) A unique ID for the metric data query.
* `label` - (Optional) The label for the metric.
* `period` - (Optional) The period, in seconds, over which the metric is aggregated.
* `return_data` - (Optional) Whether to return the metric data.
* [`metric_stat`](#metric_stat) - (Optional) Configuration block for a CloudWatch metric and statistic.

### bad_count_metric

A list of CloudWatch metric data queries. This is a **List Nested Block**.

* `account_id` - (Optional) The ID of the account to use for the metric data query.
* `expression` - (Optional) The math expression to use on the returned metric.
* `id` - (Optional) A unique ID for the metric data query.
* `label` - (Optional) The label for the metric.
* `period` - (Optional) The period, in seconds, over which the metric is aggregated.
* `return_data` - (Optional) Whether to return the metric data.
* [`metric_stat`](#metric_stat) - (Optional) Configuration block for a CloudWatch metric and statistic.

### total_request_count_metric

A list of CloudWatch metric data queries. This is a **List Nested Block**.

* `account_id` - (Optional) The ID of the account to use for the metric data query.
* `expression` - (Optional) The math expression to use on the returned metric.
* `id` - (Optional) A unique ID for the metric data query.
* `label` - (Optional) The label for the metric.
* `period` - (Optional) The period, in seconds, over which the metric is aggregated.
* `return_data` - (Optional) Whether to return the metric data.
* [`metric_stat`](#metric_stat) - (Optional) Configuration block for a CloudWatch metric and statistic.

### dependency_config

Configuration for filtering metrics related to a specific dependency.

* `dependency_key_attributes` - (Required) A map of key-value pairs that are used to filter the dependency's metric. (Type: `map(string)`)
* `dependency_operation_name` - (Required) The name of the operation for the dependency.

### metric_data_queries

A list of CloudWatch metric data queries. This is a **List Nested Block**.

* `account_id` - (Optional) The ID of the account to use for the metric data query.
* `expression` - (Optional) The math expression to use on the returned metric.
* `id` - (Optional) A unique ID for the metric data query.
* `label` - (Optional) The label for the metric.
* `period` - (Optional) The period, in seconds, over which the metric is aggregated.
* `return_data` - (Optional) Whether to return the metric data.
* [`metric_stat`](#metric_stat) - (Optional) Configuration block for a CloudWatch metric and statistic.

### metric_stat

* [`metric`](#metric) - (Optional) Configuration block for the metric.
* `period` - (Optional) The period over which the metric is aggregated.
* `stat` - (Optional) The statistic to apply to the metric.
* `unit` - (Optional) The unit for the metric.

### metric

* [`dimensions`](#dimensions) - (Optional) A list of dimensions for the CloudWatch metric.
* `metric_name` - (Optional) The name of the CloudWatch metric.
* `namespace` - (Optional) The namespace of the CloudWatch metric.

### dimensions

A list of metric dimensions. This is a **List Nested Block**.

* `name` - (Required) The name of the dimension.
* `value` - (Required) The value of the dimension.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Service Level Objective.
* `created_time` - The date and time that this SLO was created (RFC3339 format).
* `last_updated_time` - The time that this SLO was most recently updated (RFC3339 format).
* `evaluation_type` - Displays whether this is a `PERIOD_BASED` SLO or a `REQUEST_BASED` SLO.
* `metric_source_type` - Displays the source of the SLI metric for this SLO.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `5m`)
* `update` - (Default `5m`)
* `delete` - (Default `5m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Application Signals Service Level Objective using its `name`. For example:

```terraform
import {
  to = aws_applicationsignals_service_level_objective.example
  id = "my-slo-name"
}
```

Using `terraform import`, import Application Signals Service Level Objective using the `name`. For example:

```console
% terraform import aws_applicationsignals_service_level_objective.example my-slo-name
```
