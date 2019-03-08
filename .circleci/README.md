# CircleCI configuration

## Environment variables to set

These environment variables are only necessary when you do E2E tests:

- `AWS_ACCOUNT_ID`: The ID of the AWS account used to create all the resources during E2E tests.
- `AWS_REGION`: The region where to create all the resources during E2E tests. If not specified,
  a region is chosen at random for every test invocation.
- `AWS_ACCESS_KEY_ID`: The access key ID of the AWS user to use during E2E tests.
- `AWS_SECRET_ACCESS_KEY`: The secret access key of the AWS user to use during E2E tests.
