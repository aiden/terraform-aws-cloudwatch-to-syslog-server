variable name {
  description = <<EOF
Common name given to the lambda function, the IAM role, the lambda permission statement,
and the log subscription filter.
EOF
}

variable region {
  description = "The AWS region where the AWS CloudWatch Logs are located."
}

variable account_id {
  description = "The ID of the AWS account where the AWS CloudWatch Logs are located."
}

variable log_group {
  description = "The name of the AWS CloudWatch log group to forward to the syslog server."
}

variable syslog_server_host {
  description = "The host for the syslog server (e.g., logs5.papertrailapp.com)."
}

variable syslog_server_port {
  description = "The port for the syslog server."
}

variable disable_tls {
  default     = "0"
  description = "Whether to use TLS or not when communicating with the syslog server."
}
