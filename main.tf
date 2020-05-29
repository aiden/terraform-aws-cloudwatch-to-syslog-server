resource aws_lambda_function cloudwatch_to_syslog_server {
  filename         = "${path.module}/lambda.zip"
  function_name    = "${var.name}"
  role             = "${aws_iam_role.cloudwatch_to_syslog_server.arn}"
  handler          = "index.handler"
  source_code_hash = "${filebase64sha256("${path.module}/lambda.zip")}"
  runtime          = "nodejs10.x"
  timeout          = 120

  environment {
    variables = {
      SYSLOG_SERVER_HOST = "${var.syslog_server_host}"
      SYSLOG_SERVER_PORT = "${var.syslog_server_port}"
      DISABLE_TLS        = "${var.disable_tls}"
    }
  }
}

resource aws_iam_role cloudwatch_to_syslog_server {
  name = "${var.name}"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource aws_iam_role_policy_attachment lambda_basic_execution {
  role       = "${aws_iam_role.cloudwatch_to_syslog_server.name}"
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

resource aws_lambda_permission cloudwatch_logs {
  statement_id   = "${var.name}"
  action         = "lambda:InvokeFunction"
  function_name  = "${aws_lambda_function.cloudwatch_to_syslog_server.arn}"
  principal      = "logs.${var.region}.amazonaws.com"
  source_arn     = "arn:aws:logs:${var.region}:${var.account_id}:log-group:${var.log_group}:*"
  source_account = "${var.account_id}"
}

/* prod only */
resource aws_cloudwatch_log_subscription_filter papertrail {
  name            = "${var.name}"
  log_group_name  = "${var.log_group}"
  filter_pattern  = ""
  destination_arn = "${aws_lambda_function.cloudwatch_to_syslog_server.arn}"
  distribution    = "ByLogStream"
}
