terraform {
  required_version = ">= 1.0.0"
  backend "s3" {
    region = "us-east-1"
  }
  required_providers {
    aws = "~> 4.41.0"
  }
}


provider "aws" {
  region = var.region
}


module "apps_deployment" {
  source                           = "../modules/apps-deployment"
  lambda_name                      = var.lambda_name
  lambda_file                      = var.lambda_file
  bundle_name                      = var.bundle_name
  handler                          = var.handler
  runtime                          = var.runtime
  environment                      = var.environment
  private_subnet_ids               = var.private_subnet_ids

  tags = {
    Owner       = "cloud-team"
    Terraform   = "true"
    Environment = var.environment
    Purpose     = "app-deployment"
  }
}
