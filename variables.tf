variable name {
  description = <<EOF
Common name given to the lambda function, the IAM role, the lambda permission statement,
and the log subscription filter.
EOF
}

variable region {
  description = "The AWS region where the logs are located."
}

variable account_id {
  description = "The account ID where the logs are located."
}

variable log_group {
  description = "The AWS CloudWatch log group to forward to the syslog server."
}

variable syslog_server_host {
  description = "The host for the syslog server."
}

variable syslog_server_port {
  description = "The port for the syslog server."
}

variable disable_tls {
  default     = "0"
  description = "Whether to use TLS or not when communicating with the syslog server."
}
