variable "log_group" {
  description = "The log group the lambda function subscribes to."
}

variable "log_stream" {
  description = <<EOF
The log stream that is created by this example configuration. Used in tests as the only
stream to which log events are published.
EOF
}

variable "syslog_server_port" {
  description = "The port used by the syslog server."
  default     = 2048
}

locals {
  description        = "The host (IP) of the syslog server."
  syslog_server_host = "${aws_instance.syslog_server.public_ip}"
}

resource "aws_cloudwatch_log_group" "log_group" {
  name = "${var.log_group}"
}

resource "aws_cloudwatch_log_stream" "log_stream" {
  name           = "${var.log_stream}"
  log_group_name = "${aws_cloudwatch_log_group.log_group.name}"
}

module "cloudwatch_to_syslog_server" {
  source = "../.."

  name       = "cloudwatch-to-syslog-server-${var.log_group}"
  region     = "${var.region}"
  account_id = "${var.account_id}"
  log_group  = "${var.log_group}"

  // We disable TLS support because our test syslog server (ncat) is not equipped
  // for TLS.
  disable_tls = "1"

  syslog_server_host = "${local.syslog_server_host}"
  syslog_server_port = "${var.syslog_server_port}"
}

output lambda_arn {
  value = "${module.cloudwatch_to_syslog_server.lambda_arn}"
}

output log_group {
  value = "${var.log_group}"
}

output log_stream {
  value = "${var.log_stream}"
}
