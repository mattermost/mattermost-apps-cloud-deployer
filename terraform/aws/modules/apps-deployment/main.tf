
data "aws_region" "current" {}

data "terraform_remote_state" "generic" {
  backend = "s3"
  config = {
    bucket = "terraform-cloud-monitoring-state-bucket-${var.environment}"
    key    = "${data.aws_region.current.name}/mattermost-generic"
    region = "us-east-1"
  }
}

resource "aws_lambda_function" "lambda_function" {
  function_name = var.lambda_name
  role          = data.terraform_remote_state.generic.outputs.mattermost_apps_lambda_role.arn
  filename      = "../../../tmp/${var.bundle_name}/${var.lambda_file}"
  handler       = var.handler
  timeout       = 120
  runtime       = var.runtime

  vpc_config {
    subnet_ids         = flatten(var.private_subnet_ids)
    security_group_ids = [data.terraform_remote_state.generic.outputs.mattermost_apps_security_group.id]
  }

  tags = var.tags

}
