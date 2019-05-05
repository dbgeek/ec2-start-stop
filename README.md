# lambda function start-stop

[![Go Report Card](https://goreportcard.com/badge/github.com/dbgeek/ec2-start-stop)](https://goreportcard.com/report/github.com/dbgeek/ec2-start-stop) [![Build Status][travis image]][travis]

AWS Lambda function that can start and stop ec2 instances that are tagged with `StartStop`.

## Payload specification for the function

### groupbytag

Specify which tag key on ec2 instaces it should group by instances by. Lambda function will execute start/stop one group of instances at time.

Example:

if groupbytag is `group` and you will instances with value group1 and another with value group2. The lambda will execute the action on each group at the time.

### orderbytag

`orderbytag` is used to if you want execute the action on instances in an group in a specifiec order.

### orderby

`orderby` lists the value in which order you want to apply the the action on. Instances without an `orderbytag`or not diffiend `orderby`value will be applied last in each group.

### tagfilters

`tagfilters` is the filter that you want to apply when function does `describeinstances`. Lambda function will apply the action on all the ec2 instances that will be queried out.

### action

#### start

`start`will start all the ec2 instances.

#### stop

`stop`will stop all the ec2 instances.

### dryrun

When dryrun is `true`it will do an dryrun.

### loglevel

It will set logglevel(info, warning, debug.....) for logrus.

### example json document

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

### Invoke sam local

You have to create event.json file for the event you should trigger the lambda with.

```sh
sam local invoke -e event.json StartStop --region <region>
```

## Lambda iam policy

Lambda function need to have this privileges at least

* ec2:StartInstances
* ec2:StopInstances"
* ec2:DescribeInstances

### Example Policy document

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "Stmt1556908945540",
      "Action": [
        "ec2:DescribeInstances",
        "ec2:StartInstances",
        "ec2:StopInstances"
      ],
      "Effect": "Allow",
      "Resource": "arn:aws:ec2:*:*:instance/*"
    }
  ]
}
```

[travis]: https://travis-ci.org/dbgeek/ec2-start-stop
[travis image]: https://travis-ci.org/dbgeek/ec2-start-stop.png?branch=master