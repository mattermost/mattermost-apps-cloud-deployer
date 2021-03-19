variable "environment" {
  default = ""
  type    = string
}

variable "lambda_name" {
  default = ""
  type    = string
}

variable "lambda_file" {
  default = ""
  type    = string
}

variable "bundle_name" {
  default = ""
  type    = string
}

variable "handler" {
  default = ""
  type    = string
}

variable "runtime" {
  default = ""
  type    = string
}

variable "region" {
  default = "us-east-1"
  type    = string
}

variable "private_subnet_ids" {
  type    = list(string)
  default = [""]
}
