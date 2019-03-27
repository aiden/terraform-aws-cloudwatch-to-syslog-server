package test

import (
	"errors"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	aws_sdk "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/gruntwork-io/terratest/modules/aws"
	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/gruntwork-io/terratest/modules/terraform"
	test_structure "github.com/gruntwork-io/terratest/modules/test-structure"
)

// Common prefix used by the AWS resources that we create (additionally, the resources
// contain a UniqueID that is different for each test invocation).
const prefix = "cw2pt_" // cw2pt = cloudwatch to papertrail

// String echoed at the end of the user-data shell script, meaning that the ncat
// TCP server is up and ready to receive syslogs.
const userDataReadyString = "--- READY ---"

// Test that logs to cloudwatch are properly sent to a TCP syslog server using the
// example terraform configuration in `../examples/simple`.
func TestExample(t *testing.T) {
	t.Parallel()

	// A unique ID we can use to namespace resources so we don't clash with anything
	// already in the AWS account or tests running in parallel
	uniqueID := random.UniqueId()

	// Pick a random AWS region to test in, or use the AWS_REGION env var.
	awsRegion := os.Getenv("AWS_REGION")
	if awsRegion == "" {
		awsRegion = aws.GetRandomStableRegion(t, nil, nil)
	}

	terraformDir := test_structure.CopyTerraformFolderToTemp(t, "../", "examples/simple")

	// At the end of the test, run `terraform destroy` to clean up any resources that
	// were created.
	defer test_structure.RunTestStage(t, "teardown", func() {
		terraformOptions := test_structure.LoadTerraformOptions(t, terraformDir)
		terraform.Destroy(t, terraformOptions)

		keyPair := test_structure.LoadEc2KeyPair(t, terraformDir)
		aws.DeleteEC2KeyPair(t, keyPair)

		// Delete the log group used for the logs of the AWS Lambda function itself.
		// This log group is automatically created by AWS, and not managed by
		// Terraform.
		logsClient := aws.NewCloudWatchLogsClient(t, awsRegion)
		lambdaArn := test_structure.LoadString(t, terraformDir, "lambdaArn")
		if _, err := logsClient.DeleteLogGroup(&cloudwatchlogs.DeleteLogGroupInput{
			LogGroupName: aws_sdk.String(awsLambdaLogGroup(lambdaArn)),
		}); err != nil {
			t.Errorf("cannot delete AWS Lambda log group: %v", err)
		}
	})

	test_structure.RunTestStage(t, "setup", func() {
		keyPairName := prefix + uniqueID
		keyPair := aws.CreateAndImportEC2KeyPair(t, awsRegion, keyPairName)

		awsAccountID := os.Getenv("AWS_ACCOUNT_ID")
		if awsAccountID == "" {
			t.Fatal("AWS_ACCOUNT_ID env variable should be set")
		}

		terraformOptions := &terraform.Options{
			// The path to where our Terraform code is located
			TerraformDir: terraformDir,

			// Variables to pass to our Terraform code using -var options
			Vars: map[string]interface{}{
				"region":                      awsRegion,
				"log_group":                   prefix + "lg_" + uniqueID,
				"log_stream":                  prefix + "ls_" + uniqueID,
				"account_id":                  awsAccountID,
				"key_pair_name":               keyPairName,
				"syslog_server_instance_name": prefix + "syslog_server_" + uniqueID,
			},
		}

		// Save the options and key pair so later test stages can use them
		test_structure.SaveTerraformOptions(t, terraformDir, terraformOptions)
		test_structure.SaveEc2KeyPair(t, terraformDir, keyPair)

		// This will run `terraform init` and `terraform apply` and fail the test if there are any errors
		terraform.InitAndApply(t, terraformOptions)
	})

	test_structure.RunTestStage(t, "validate", func() {
		// Load the options and key pair from the setup stage
		terraformOptions := test_structure.LoadTerraformOptions(t, terraformDir)
		keyPair := test_structure.LoadEc2KeyPair(t, terraformDir)

		// Check that the lambda_arn is outputted
		terraform.Output(t, terraformOptions, "lambda_arn")

		// We get the instance ID of the syslog server to extract the logs for the
		// user-data shell script (executed at startup) and the ncat TCP session dump.
		syslogServerInstanceID := terraform.Output(t, terraformOptions, "syslog_server_instance_id")

		// We wait for the user-data logs to be accessible, meaning that the machine is
		// up and that sshd is running, and then we wait for the logs to contain the
		// userDataReadyString which signals that ncat has been launched in the background.
		userDataLogs, err := retry.DoWithRetryE(t, "get user-data.log", 50, 2*time.Second, func() (string, error) {
			logs, err := aws.FetchContentsOfFileFromInstanceE(
				t,
				awsRegion,
				"ubuntu",
				keyPair,
				syslogServerInstanceID,
				true,
				"/var/log/user-data.log",
			)
			if err != nil {
				return "", err
			}
			if !strings.Contains(logs, userDataReadyString) {
				return logs, errors.New("cannot find ready signal in logs")
			}
			return logs, nil
		})
		if err != nil {
			t.Fatalf(
				"cannot find ready signal '%s' in logs: <<%s>>",
				userDataReadyString,
				userDataLogs,
			)
		}

		// Wait for ncat, actually launched in the background, to be fully up.
		// TODO: Remove this potential source of flakiness (but for doing that
		// we need to inspec the logs of the AWS Lambda function to know whether
		// the function was able to connect).
		time.Sleep(3 * time.Second)

		lambdaArn := terraform.Output(t, terraformOptions, "lambda_arn")
		logGroup := terraform.Output(t, terraformOptions, "log_group")
		logStream := terraform.Output(t, terraformOptions, "log_stream")

		test_structure.SaveString(t, terraformDir, "lambdaArn", lambdaArn)

		logsClient := aws.NewCloudWatchLogsClient(t, awsRegion)
		if _, err := logsClient.PutLogEvents(&cloudwatchlogs.PutLogEventsInput{
			LogGroupName:  &logGroup,
			LogStreamName: &logStream,
			LogEvents: []*cloudwatchlogs.InputLogEvent{
				&cloudwatchlogs.InputLogEvent{
					Message: aws_sdk.String("Hello, world"),
					// The timestamp does not matter.  It is not currently picked up by
					// the lambda function when converting to syslogs and sending
					// over the TCP connection.
					Timestamp: aws_sdk.Int64(time.Now().UnixNano() / 1000000),
				},
			},
		}); err != nil {
			t.Error(err)
		}

		ncatSession, err := retry.DoWithRetryE(t, "get ncat-session.log", 10, 1*time.Second, func() (string, error) {
			// Get the TCP session dump captured by ncat
			ncatSession, err := aws.FetchContentsOfFileFromInstanceE(
				t,
				awsRegion,
				"ubuntu",
				keyPair,
				syslogServerInstanceID,
				true,
				"/var/log/ncat-session.log",
			)
			if err != nil {
				return "", err
			}

			// Check that the TCP session dump contains the log event that we sent on cloudwatch,
			// and that the format is indeed that of a syslog.
			hostname := logGroup
			program := logStream
			regexStr := `<30>1 \d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{1,3}Z ` + hostname + ` ` + program + ` - - - Hello, world.*`
			if !regexp.MustCompile(regexStr).MatchString(ncatSession) {
				return ncatSession, errors.New("ncatSession does not match regex <<" + regexStr + ">>")
			}
			return ncatSession, nil
		})
		if err != nil {
			t.Errorf("ncatSession does not match regex: <<%s>>", ncatSession)
		}

		// In case of failure, it might be useful to get the logs for the AWS
		// Lambda function
		awsLambdaLogs, err := retry.DoWithRetryE(t, "get AWS Lambda Logs", 5, 1*time.Second, func() (string, error) {
			logGroupName := aws_sdk.String(awsLambdaLogGroup(lambdaArn))

			// Get the name of the first log stream we encounter in the log group
			logStreams, err := logsClient.DescribeLogStreams(&cloudwatchlogs.DescribeLogStreamsInput{
				LogGroupName: logGroupName,
			})
			if len(logStreams.LogStreams) == 0 {
				return "", errors.New("no log stream found in log group ")
			}
			logStreamName := logStreams.LogStreams[0].LogStreamName

			// Join all the log messages in the log stream
			var awsLambdaLogs string
			if err != nil {
				return "", err
			} else if err := logsClient.GetLogEventsPages(&cloudwatchlogs.GetLogEventsInput{
				LogGroupName:  logGroupName,
				LogStreamName: logStreamName,
			}, func(output *cloudwatchlogs.GetLogEventsOutput, lastPage bool) bool {
				for _, event := range output.Events {
					awsLambdaLogs += *event.Message
				}
				return true
			}); err != nil {
				return "", err
			}

			// Check that the log messages are complete.  An AWS Lambda invocation by
			// defaults end with a "REPORT" log message.
			if !strings.Contains(awsLambdaLogs, "\nREPORT") {
				return "", errors.New("awsLambdaLogs does not match regex")
			}
			return awsLambdaLogs, nil
		})

		if err != nil {
			t.Error(err)
		}
		logger.Logf(t, "AWS Lambda logs:\n------\n%s------", awsLambdaLogs)
	})
}

// awsLambdaNameFromArn gets the name of a Lambda function from its ARN.
func awsLambdaNameFromArn(lambdaArn string) string {
	return lambdaArn[strings.LastIndex(lambdaArn, ":")+1:]
}

// awsLambdaLogGroup returns the log group where the logs for a Lambda function
// with the given ARN are outputted.
func awsLambdaLogGroup(lambdaArn string) string {
	return "/aws/lambda/" + awsLambdaNameFromArn(lambdaArn)
}
