# lambda function start-stop

AWS Lambda function that can start and stop ec2 instances that are tagged with `StartStop`.

## Payload specification for the function

```json
{
    "groupbytag": "group-by-tag",
    "orderbytag": "order-by-tag",
    "OrderBy": [
        "tag1",
        "tag2"
    ],
    "tagfilters": [
        {
            "name": "tag:filter-key",
            "value": [
                "tag:filter-valye"
            ]
        }
    ],
    "action": "start",
    "dryrun": false,
    "loglevel": "info"
}
```

## Deploy

### Deploy by AWS CLI

```sh
aws lambda create-function --function-name start-stop \
--zip-file fileb://start-stop/start-stop.zip --handler start-stop --runtime go1.x \
--role arn:aws:iam::123456789012:role/lambda-start-stop-role
```

## Invoke lambda

### Invoke by AWS CLI

```sh
aws lambda invoke --function-name start-stop --log-type Tail \
--payload '{"groupbytag": "group-by-tag", "orderbytag": "order-by-tag", "OrderBy": ["tag1","tag2"], "tagfilters": [{ "name": "tag:filter-key", "value": ["tag:filter-valye"]}], "action": "start" }' \
outputfile.txt
```