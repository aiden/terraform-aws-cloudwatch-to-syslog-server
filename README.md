# cloudwatch-to-syslog-server: A Terraform module to send CloudWatch logs to a syslog server

[![Maintained by aiden.ai](https://img.shields.io/badge/maintained%20by-aiden.ai-blue.svg)](https://aiden.ai) [![CircleCI](https://circleci.com/gh/aiden/terraform-aws-cloudwatch-to-syslog-server/tree/master.svg?style=svg)](https://circleci.com/gh/aiden/terraform-aws-cloudwatch-to-syslog-server/tree/master) [![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

_(This module is available both [on GitHub](https://github.com/aiden/terraform-aws-cloudwatch-to-syslog-server) and [on the Terraform Registry](https://registry.terraform.io/modules/aiden/cloudwatch-to-syslog-server).)_

[![cloudwatch-to-syslog-server](https://github.com/aiden/terraform-aws-cloudwatch-to-syslog-server/raw/master/docs/cloudwatch-to-syslog-server.svg?sanitize=true)](./docs/cloudwatch-to-syslog-server.svg)

cloudwatch-to-syslog-server is a Terraform module that defines an AWS Lambda function
to forward the CloudWatch logs of a given log group to a syslog server. Many third-party services offer to collect logs with a syslog server, for instance:

- [Papertrail](https://papertrailapp.com/)
- [logstash](https://www.elastic.co/guide/en/logstash/current/plugins-inputs-syslog.html)
  (from the [ELK stack](https://www.elastic.co/elk-stack))
- [Datadog](https://docs.datadoghq.com/logs/?tab=ussite#log-collection)

## Example

You can find an example Terraform configuration in the [example folder](https://github.com/aiden/terraform-aws-cloudwatch-to-syslog-server/tree/master/examples/simple).

## Why cloudwatch-to-syslog-server?

AWS CloudWatch is meant for durable and scalable log archiving. It is tightly
integrated with ECS and, overall, the AWS ecosystem, which makes it an interesting
choice for low-cost, long-term log archiving. However, the browsing experience
is poor, which is something other people have remarked ([AWS CloudWatch logs for Humans][],
[Elasticsearch+Kibana][]). In this context, Papertrail offers a seamless browsing
experience that would be the equivalent, as a service, of a full-blown Elasticsearch
cluster.

[aws cloudwatch logs for humans]: https://github.com/jorgebastida/awslogs
[elasticsearch+kibana]: https://aws.amazon.com/blogs/aws/cloudwatch-logs-subscription-consumer-elasticsearch-kibana-dashboards/

This repository is a full solution for forwarding the CloudWatch logs belonging
to a specific log group to a syslog TCP server. Papertrail exposes such TCP servers, but
our implementation can accommodate any syslog TCP server. (As a side note, Papertrail
also exposes UDP servers, but we are subscribing to the CloudWatch logs, and UPD is not
available on AWS Lambda, see the [FAQ](https://aws.amazon.com/lambda/faqs/),
"What restrictions apply to AWS Lambda function code?")

## Other projects

This is a full solution written in Terraform, including an AWS Lambda function
written for the NodeJS runtime. The code for this function has been adapted
from https://github.com/apiaryio/cloudwatch-to-papertrail to add more reliability.
We also have added end-to-end tests to make sure that the Terraform module as a whole
fulfills its contract.

## Hostnames and programs with syslog, and why it matters for ECS clusters

With the syslog format, the messages are tagged with a hostname and a program. Here,
the hostname is equal to the name of the AWS CloudWatch log group, and the program
is equal to a transformation of the AWS CloudWatch log stream. This transformation
is specifically tailored to ECS clusters: if all the log streams within a cluster
goes to the same log group, you then get one syslog hostname per cluster,
and one syslog program per ECS service.

## License

cloudwatch-to-syslog-server is licensed under the MIT License.
