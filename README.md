## transaction-signer

This lambda takes a raw TX from the sqs queue and signs it using the right key from the TSM nodes. Essentially, the tx becomes valid and ready to be broadcasted onchain.
The TX is then serialized and added to another queue that would trigger the lambda that sends the TX onchain

## Requirements

* AWS CLI already configured with the right permission
* [Docker installed](https://www.docker.com/community-edition)
* [Golang](https://golang.org)
* SAM CLI

## Setup process

### Installing dependencies & building the target

In this example we use the built-in `sam build` to automatically download all the dependencies and package our build target.
Read more about [SAM Build here](https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/sam-cli-command-reference-sam-build.html)

The `sam build` command is wrapped inside of the `Makefile`. To execute this simply run

```shell
make
```

## Integration test

```bash
$ make integration_test
```

## Running locally

The first step is to update the `template.yaml` file:

```yaml
      Environment:
        Variables:
          LOG_LEVEL: debug
          ENVIRONMENT: local
          DD_LOG_LEVEL: "DEBUG"
          DD_API_KEY: fbb9ed85a32ec775518768a8ebf1d5a7
          DD_SITE: "datadoghq.eu"
          SEPIOR_SECRET_NAME: sepior-user
          SEPIOR_SECRET_AWS_REGION: eu-west-2
          WALLETS_POSTGRES_DSN: "postgres://mara:mara@host.docker.internal:4432/wallets?sslmode=disable"
          TRANSACTIONS_POSTGRES_DSN: "postgres://mara:mara@host.docker.internal:4432/transactions?sslmode=disable"
          SQS_REGION: eu-west-2
          SQS_LOCALSTACK_ENDPOINT: http://host.docker.internal:4566
          SQS_WRITE_QUEUE_NAME: new_signed_transaction_queue
          SERVICE_NAME: transaction_signer

```

If you have noticed, we used `new_signed_transaction_queue` as the queue we will be writing to. It is important to create that queue.
That can be done with `aws --endpoint-url=http://localhost:4566 sqs create-queue --queue-name new_signed_transaction_queue --region eu-west-2`

A sample events file is provided at `events/event.example.json` to run the lambda 

> It might make sense to replace the `wallet_row_id` bit in the events file so it matches up with a valid `id` in your database. As this lambda uses that value
to fetch the right derivation path to connect to the tsm nodes so we can successfully sign the TX for the right user/address

```bash

$ make run-local

```

## Packaging and deployment

```bash
$ make deploy
```

