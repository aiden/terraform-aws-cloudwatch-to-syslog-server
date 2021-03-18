output log_group {
  value = var.log_group

  description = <<EOF
The name of the log group that is subscribed to. Its log events are forwarded to the syslog server.
EOF
}

output lambda_arn {
  value       = aws_lambda_function.cloudwatch_to_syslog_server.arn
  description = "The ARN of the lambda function subscribed to the log group."
}
