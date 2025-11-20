---
subcategory: "Application Signals"
layout: "aws"
page_title: "AWS: aws_applicationsignals_service_level_objective"
description: |-
  Manages an AWS Application Signals Service Level Objective.
---
<!---
Documentation guidelines:
- Begin resource descriptions with "Manages..."
- Use simple language and avoid jargon
- Focus on brevity and clarity
- Use present tense and active voice
- Don't begin argument/attribute descriptions with "An", "The", "Defines", "Indicates", or "Specifies"
- Boolean arguments should begin with "Whether to"
- Use "example" instead of "test" in examples
--->

# Resource: aws_applicationsignals_service_level_objective

Manages an AWS Application Signals Service Level Objective.

## Example Usage

### Basic Usage

```terraform
resource "aws_applicationsignals_service_level_objective" "example" {
  name = "elb-error-rate"
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
    sli_metric {
      comparison_operator = "LessThan"
      metric_threshold    = 2
      metric_data_queries {
        id = "m1"
        metric_stat {
          metric {
            namespace  = "AWS/ApplicationELB"
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

## Argument Reference

The following arguments are required:

* `name` - (Required) The name of this SLO.
* [`goal`](#goal)- (Required) Configuration block containing the attributes that determine the goal of this SLO.

The following arguments are optional:

* `description` - (Optional) Brief description of the optional argument.
* [`burn_rate_configurations`](#burn_rate_configurations) - (Optional) Configuration block containing the attributes that determine the burn rate of this SLO.

### burn_rate_configurations

This argument is processed in [attribute-as-blocks mode](https://www.terraform.io/docs/configuration/attr-as-blocks.html).

The following arguments are required:

* `look_back_window_minutes` - (Required) The number of minutes to use as the look back window for calculating the burn rate.

### goal

This argument is processed in [attribute-as-blocks mode](https://www.terraform.io/docs/configuration/attr-as-blocks.html).

### interval

This argument is processed in [attribute-as-blocks mode](https://www.terraform.io/docs/configuration/attr-as-blocks.html).

The following arguments are required:



### rolling_interval

This argument is processed in [attribute-as-blocks mode](https://www.terraform.io/docs/configuration/attr-as-blocks.html).

The following arguments are required:

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Service Level Objective.
* `created_time` - The date and time that this SLO was created.
* `last_updated_time` - The time that this SLO was most recently updated.
* `evaluation_type` - Displays whether this is a period-based SLO or a request-based SLO.
* `metric_source_type` - Displays the SLI metric source type for this SLO.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `5m`)
* `update` - (Default `5m`)
* `delete` - (Default `5m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Application Signals Service Level Objective using the `example_id_arg`. For example:

```terraform
import {
  to = aws_applicationsignals_service_level_objective.example
  id = "service_level_objective-12345678"
}
```

Using `terraform import`, import Application Signals Service Level Objective using the `name`. For example:

```console
% terraform import aws_applicationsignals_service_level_objective.example service_level_objective-12345678
```
