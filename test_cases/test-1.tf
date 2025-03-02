terraform {
  backend "s3" {
    dynamodb_table = "deployment-terraform-lock"
    region         = "eu-west-2"
    bucket         = "deployment-terraform-123456789"
    key            = "terraform.tfstate"
    encrypt        = true
  }

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 5.11.0"
    }
  }

  required_version = ">= 1.2.0"
}

provider "aws" {
  region = "eu-central-1"
}

provider "aws" {
  alias  = "infra-account"
  region = "eu-central-1"
  assume_role {
    role_arn = "arn:aws:iam::0987654321:role/assumable_role"
  }
}
locals {
  app_name    = "bruno-beans"
  app_version = 1
  app_float   = 1.45

  cba_base_domain  = var.cba_base_domain
  tasks            = jsondecode("[{\"name\":\"datetime\",\"image\":\"datetime-image-path\",\"env\":[{\"name\":\"ALWAYS_LOG_WARNINGS_STDERR\",\"value\":\"1\"},{\"name\":\"SERVER_ROLE_MAY_RUN_EXPERIMENTS\",\"value\":\"0\"}]}]")
  rollout_strategy = jsondecode("{\"Steps\":[{\"Name\":\"staging\",\"Traffic\":{\"Old\":100,\"New\":0}},{\"Name\":\"vanguard\",\"Traffic\":{\"Old\":90,\"New\":10}},{\"Name\":\"full on\",\"Traffic\":{\"Old\":0,\"New\":100}}]}")
  iam_policy_names = ["rds_policy_1", "rds_policy_2"]
}

locals {
  tags = {
    Terraform = "true"
    App       = local.app_name
  }
}

data "aws_region" "current" {}
data "aws_caller_identity" "current" {}
module "bruno-beans-7132aaa" {
  source = "git::https://github.com/module"

  app_name         = local.app_name
  app_version      = local.cell_version
  cba_base_domain  = local.cba_base_domain
  tasks            = local.tasks
  iam_policy_names = local.iam_policy_names
  cba_environment  = local.environment
  tag = {
    deployment_sha1 = "7132aaa4-6db3-4cae-8d36-8e903fd06698"
  }
}

resource "aws_vpc_endpoint" "this" {
  for_each = local.endpoints

  vpc_id            = var.vpc_id
  service_name      = try(each.value.service_endpoint, data.aws_vpc_endpoint_service.this[each.key].service_name)
  service_region    = try(each.value.service_region, null)
  vpc_endpoint_type = try(each.value.service_type, "Interface")
  auto_accept       = try(each.value.auto_accept, null)

  security_group_ids  = try(each.value.service_type, "Interface") == "Interface" ? length(distinct(concat(local.security_group_ids, lookup(each.value, "security_group_ids", [])))) > 0 ? distinct(concat(local.security_group_ids, lookup(each.value, "security_group_ids", []))) : null : null
  subnet_ids          = try(each.value.service_type, "Interface") == "Interface" ? distinct(concat(var.subnet_ids, lookup(each.value, "subnet_ids", []))) : null
  route_table_ids     = try(each.value.service_type, "Interface") == "Gateway" ? lookup(each.value, "route_table_ids", null) : null
  policy              = try(each.value.policy, null)
  private_dns_enabled = try(each.value.service_type, "Interface") == "Interface" ? try(each.value.private_dns_enabled, null) : null
  ip_address_type     = try(each.value.ip_address_type, null)

  dynamic "dns_options" {
    for_each = try([each.value.dns_options], [])

    content {
      dns_record_ip_type                             = try(dns_options.value.dns_options.dns_record_ip_type, null)
      private_dns_only_for_inbound_resolver_endpoint = try(dns_options.value.private_dns_only_for_inbound_resolver_endpoint, null)
    }
  }

  tags = merge(
    var.tags,
    { "Name" = replace(each.key, ".", "-") },
    try(each.value.tags, {}),
  )

  timeouts {
    create = try(var.timeouts.create, "10m")
    update = try(var.timeouts.update, "10m")
    delete = try(var.timeouts.delete, "10m")
  }
}
