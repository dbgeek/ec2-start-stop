package main

import (
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/cenkalti/backoff"
	"github.com/sirupsen/logrus"
)

var (
	log     *logrus.Logger
	logging logrus.FieldLogger
)

type (
	// TagFilters tags to filter out ec2 instances
	TagFilters struct {
		Name  string   `json:"name"`
		Value []string `json:"value"`
	}
	// StartStopEvent incomment event
	StartStopEvent struct {
		TagFilters []TagFilters `json:"tagfilters"`
		GroupByTag string       `json:"groupbytag"`
		OrderByTag string       `json:"orderbytag"`
		OrderBy    []string     `json:"orderby"`
		Action     string       `json:"action"`
		DryRun     bool         `json:"dryrun"`
		LogLevel   string       `json:"loglevel"`
	}
	instance struct {
		instanceID string
		tags       map[string]string
	}
	instances        []instance
	groupedInstances map[string]instances
	orderedInstances map[string]instances
)

func init() {
	// Log as JSON instead of the default ASCII formatter.
	log = logrus.New()
	log.SetOutput(os.Stdout)
	log.SetFormatter(&logrus.JSONFormatter{})
	log.SetLevel(logrus.InfoLevel)
	logging = logrus.NewEntry(log)

}

func newInstance(i *ec2.Instance) instance {

	inst := i.InstanceId
	tags := make(map[string]string)

	for _, tag := range i.Tags {
		tags[*tag.Key] = *tag.Value
	}
	return instance{
		instanceID: *inst,
		tags:       tags,
	}
}

func waitToRunning(svc ec2iface.EC2API, instances []string) {
	describeInstancesInput := &ec2.DescribeInstancesInput{
		InstanceIds: aws.StringSlice(instances),
	}

	operation := func() error {
		allRunning := false
		running := []string{}
		notRunning := []string{}
		result, err := svc.DescribeInstances(describeInstancesInput)
		if err != nil {
			logging.Fatalf("waitToRunning: DescribeInstances failed with error: %v", err)
		}
		for _, reservations := range result.Reservations {
			for _, ec2 := range reservations.Instances {
				if *ec2.State.Name == "running" {
					allRunning = true
					running = append(running, *ec2.InstanceId)

				} else {
					allRunning = false
					notRunning = append(notRunning, *ec2.InstanceId)
				}
			}
		}
		if !allRunning {
			return fmt.Errorf("All ec2 instaces is not running! Running: %v, NotRunning: %v", running, notRunning)
		}
		return nil
	}

	err := backoff.Retry(operation, backoff.NewExponentialBackOff())
	if err != nil {
		// Handle error.
		logging.Fatalf("Retry failed with error: %v", err)
	}
}

func startInstances(svc ec2iface.EC2API, instances []string, dryRun bool) {
	startInstancesInput := ec2.StartInstancesInput{
		DryRun:      aws.Bool(dryRun),
		InstanceIds: aws.StringSlice(instances),
	}
	startInstancesOutput, err := svc.StartInstances(&startInstancesInput)
	if err != nil {
		logging.Fatalf("Failed to StartInstances. Gott error: %v", err)
	}
	if dryRun {
		logging.Infof("dry run output: %v", startInstancesOutput)
	}
	waitToRunning(svc, instances)
}

func stopInstances(svc ec2iface.EC2API, instances []string, dryRun bool) {
	logging.Infof("stopping instanceIDs: %v", instances)
	stopInstancesInput := ec2.StopInstancesInput{
		DryRun:      aws.Bool(dryRun),
		Force:       aws.Bool(false),
		Hibernate:   aws.Bool(false),
		InstanceIds: aws.StringSlice(instances),
	}
	stopInstancesOutput, err := svc.StopInstances(&stopInstancesInput)
	if err != nil {
		logging.Fatalf("Failed to stopInstances. Gott error: %v", err)
	}
	if dryRun {
		logging.Infof("dry run output: %v", stopInstancesOutput)
	}
	waitToStopped(svc, instances)
}

func waitToStopped(svc ec2iface.EC2API, instances []string) {

	describeInstancesInput := &ec2.DescribeInstancesInput{
		InstanceIds: aws.StringSlice(instances),
	}

	operation := func() error {
		AllStopped := false
		stopped := []string{}
		notStopped := []string{}
		result, err := svc.DescribeInstances(describeInstancesInput)
		if err != nil {
			logging.Fatalf("waitToRunning: DescribeInstances failed with error: %v", err)
		}
		for _, reservations := range result.Reservations {
			for _, ec2 := range reservations.Instances {
				if *ec2.State.Name == "stopped" {
					AllStopped = true
					stopped = append(stopped, *ec2.InstanceId)

				} else {
					AllStopped = false
					notStopped = append(notStopped, *ec2.InstanceId)
				}
			}
		}
		if !AllStopped {
			return fmt.Errorf("All ec2 instaces is not stopped! Running: %v, NotRunning: %v", stopped, notStopped)
		}
		return nil
	}

	err := backoff.Retry(operation, backoff.NewExponentialBackOff())
	if err != nil {
		// Handle error.
		logging.Fatalf("Retry failed with error: %v", err)
	}

}

func (g groupedInstances) stop(svc ec2iface.EC2API, event StartStopEvent) {
	orderByTag := event.OrderByTag
	orderBy := event.OrderBy

	// adding default order bucket last in the slice
	orderBy = append(orderBy, "<no-order>")

	for _, group := range g {
		for _, v := range group.byOrder(orderByTag, orderBy) {
			var instanceList []string
			for _, key := range v {
				instanceList = append(instanceList, key.instanceID)
			}
			stopInstances(svc, instanceList, event.DryRun)
		}
	}
}

func (g groupedInstances) start(svc ec2iface.EC2API, event StartStopEvent) {
	orderByTag := event.OrderByTag
	orderBy := event.OrderBy
	for _, val := range g {
		for _, v := range val.byOrder(orderByTag, orderBy) {
			var instanceList []string
			for _, key := range v {
				instanceList = append(instanceList, key.instanceID)
			}
			startInstances(svc, instanceList, event.DryRun)
		}
	}
}

func (i instances) byOrder(orderTag string, order []string) orderedInstances {

	var missing instances
	var missingValue instances
	byOrderInstances := make(orderedInstances)
	taken := make(map[string]bool)

	for _, tag := range order {
		for _, instance := range i {
			if _, ok := taken[instance.instanceID]; ok {
				continue
			}
			if value, ok := instance.tags[orderTag]; ok {
				if value == tag {
					if _, ok := byOrderInstances[tag]; ok {
						byOrderInstances[tag] = append(byOrderInstances[tag], instance)
						taken[instance.instanceID] = true
					} else {
						byOrderInstances[tag] = instances{instance}
						taken[instance.instanceID] = true
					}
				}
			} else {
				missing = append(missing, instance)
				taken[instance.instanceID] = true
			}
		}
	}

	for _, instance := range i {
		if _, ok := taken[instance.instanceID]; ok {
			continue
		}
		missingValue = append(missingValue, instance)

	}
	if len(missingValue) > 0 {
		byOrderInstances["<no-order>"] = missingValue
	}

	if len(missing) > 0 {
		if _, ok := byOrderInstances["<no-order>"]; ok {
			byOrderInstances["<no-order>"] = append(byOrderInstances["<no-order>"], missing...)
		} else {
			byOrderInstances["<no-order>"] = missing
		}
	}
	return byOrderInstances
}

func (s *StartStopEvent) ec2Filters() []*ec2.Filter {

	filters := []*ec2.Filter{}
	for _, value := range s.TagFilters {
		filters = append(filters, &ec2.Filter{
			Name:   aws.String(fmt.Sprintf("tag:%s", value.Name)),
			Values: aws.StringSlice(value.Value),
		})
	}
	filters = append(filters, &ec2.Filter{
		Name: aws.String(fmt.Sprintf("tag:%s", "StartStop")),
		Values: []*string{
			aws.String(""),
		},
	})
	return filters
}

func getDescribeInstaces(svc ec2iface.EC2API, event StartStopEvent) *ec2.DescribeInstancesOutput {
	input := &ec2.DescribeInstancesInput{
		Filters: event.ec2Filters(),
	}

	describeInstancesOutput, err := svc.DescribeInstances(input)
	if err != nil {
		logging.Fatalf("DescribeInstances failed with error: %v", err)
	}

	return describeInstancesOutput
}

func startStopBuilder(describeInstancesOutput *ec2.DescribeInstancesOutput, event StartStopEvent) groupedInstances {
	var noGroupTag instances

	stopStartInstances := make(groupedInstances)

	for _, reservations := range describeInstancesOutput.Reservations {
		for _, ec2 := range reservations.Instances {
			i := newInstance(ec2)
			logging.Debugf("instanceID: %s tags: %v", i.instanceID, i.tags)
			if _, ok := i.tags[event.GroupByTag]; ok {
				groupTagValue := i.tags[event.GroupByTag]
				if _, ok := stopStartInstances[groupTagValue]; ok {
					stopStartInstances[groupTagValue] = append(stopStartInstances[groupTagValue], i)
				} else {
					stopStartInstances[groupTagValue] = instances{i}
				}
			} else {
				logging.Debugf("%s tag key missing for instanceID: %s", event.GroupByTag, i.instanceID)
				noGroupTag = append(noGroupTag, i)
			}
		}
	}

	if len(noGroupTag) > 0 {
		stopStartInstances["<missing_grp_tag>"] = noGroupTag
	}

	return stopStartInstances
}

func handler(event StartStopEvent) error {
	logging.Infof("Starting handlerfunction for StartStop lambda")
	logging.Infof("start stop event: %v", event)

	if len(event.TagFilters) < 1 {
		logging.Fatalf("No filter applied")
	}

	if event.LogLevel != "" {
		logLvl, err := logrus.ParseLevel(event.LogLevel)
		if err != nil {
			logging.Fatalf("failed to parse log level. Error: %v", err)
		}
		log.SetLevel(logLvl)
	}

	svc := ec2.New(session.New())

	describeInstancesOutput := getDescribeInstaces(svc, event)

	stopStartInstances := startStopBuilder(describeInstancesOutput, event)

	logging.Debugf("stopStartInstances: %v", stopStartInstances)

	switch action := event.Action; action {
	case "stop":
		stopStartInstances.stop(svc, event)
	case "start":
		stopStartInstances.start(svc, event)
	default:
		logging.Fatalf("Unsupported action %s", action)
	}

	return nil
}

func main() {
	lambda.Start(handler)
}
