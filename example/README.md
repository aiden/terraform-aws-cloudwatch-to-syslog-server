# Terraform configuration example

The example Terraform configuration is both used as an example of how to set up
`cloudwatch_to_papertrail` and also for testing purposes.

This example, instead of setting up a connection to a remote syslog TCP server
such as papertrail, starts up an AWS EC2 instance that uses
[ncat](https://nmap.org/ncat/) in listen mode to record a TCP session.
