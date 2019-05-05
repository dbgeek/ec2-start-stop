package main

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func TestStartStopBuilder(t *testing.T) {
	tt := []struct {
		name  string
		input ec2.DescribeInstancesOutput
		event StartStopEvent
		out   groupedInstances
	}{
		{
			name: "two_groups",
			input: ec2.DescribeInstancesOutput{
				Reservations: []*ec2.Reservation{
					&ec2.Reservation{
						Instances: []*ec2.Instance{
							&ec2.Instance{InstanceId: aws.String("i-123456789"),
								PrivateIpAddress: aws.String("127.0.0.1"),
								Tags: []*ec2.Tag{
									&ec2.Tag{Key: aws.String("groupTag"), Value: aws.String("grp1")},
									&ec2.Tag{Key: aws.String("orderByTag"), Value: aws.String("1")},
								},
							},
							&ec2.Instance{InstanceId: aws.String("i-987654321"),
								PrivateIpAddress: aws.String("127.0.0.1"),
								Tags: []*ec2.Tag{
									&ec2.Tag{Key: aws.String("groupTag"), Value: aws.String("grp1")},
									&ec2.Tag{Key: aws.String("orderByTag"), Value: aws.String("2")},
								},
							},
							&ec2.Instance{InstanceId: aws.String("i-567891234"),
								PrivateIpAddress: aws.String("127.0.0.1"),
								Tags: []*ec2.Tag{
									&ec2.Tag{Key: aws.String("groupTag"), Value: aws.String("grp2")},
									&ec2.Tag{Key: aws.String("orderByTag"), Value: aws.String("1")},
								},
							},
						},
					},
				},
				NextToken: aws.String(""),
			},
			out: groupedInstances{
				"grp1": instances{
					instance{
						instanceID: "i-123456789",
						tags: map[string]string{
							"groupTag":   "grp1",
							"orderByTag": "1",
						},
					},
					instance{
						instanceID: "i-987654321",
						tags: map[string]string{
							"groupTag":   "grp1",
							"orderByTag": "2",
						},
					},
				},
				"grp2": instances{
					instance{
						instanceID: "i-567891234",
						tags: map[string]string{
							"groupTag":   "grp2",
							"orderByTag": "1",
						},
					},
				},
			},
			event: StartStopEvent{
				GroupByTag: "groupTag",
				OrderByTag: "orderByTag",
				OrderBy: []string{
					"1",
					"2",
				},
			},
		},
		{
			name: "missing-groupby-tag",
			input: ec2.DescribeInstancesOutput{
				Reservations: []*ec2.Reservation{
					&ec2.Reservation{
						Instances: []*ec2.Instance{
							&ec2.Instance{InstanceId: aws.String("i-123456789"),
								PrivateIpAddress: aws.String("127.0.0.1"),
								Tags: []*ec2.Tag{
									&ec2.Tag{Key: aws.String("orderByTag"), Value: aws.String("1")},
								},
							},
						},
					},
				},
				NextToken: aws.String(""),
			},
			out: groupedInstances{
				"<missing_grp_tag>": instances{
					instance{
						instanceID: "i-123456789",
						tags: map[string]string{
							"orderByTag": "1",
						},
					},
				},
			},
			event: StartStopEvent{
				GroupByTag: "groupTag",
				OrderByTag: "orderByTag",
				OrderBy: []string{
					"1",
					"2",
				},
			},
		},
		{
			name: "one-with-grp-tag-one-missing-grp-tag",
			input: ec2.DescribeInstancesOutput{
				Reservations: []*ec2.Reservation{
					&ec2.Reservation{
						Instances: []*ec2.Instance{
							&ec2.Instance{InstanceId: aws.String("i-123456789"),
								PrivateIpAddress: aws.String("127.0.0.1"),
								Tags: []*ec2.Tag{
									&ec2.Tag{Key: aws.String("orderByTag"), Value: aws.String("1")},
								},
							},
							&ec2.Instance{InstanceId: aws.String("i-567891234"),
								PrivateIpAddress: aws.String("127.0.0.1"),
								Tags: []*ec2.Tag{
									&ec2.Tag{Key: aws.String("groupTag"), Value: aws.String("grp2")},
									&ec2.Tag{Key: aws.String("orderByTag"), Value: aws.String("1")},
								},
							},
						},
					},
				},
				NextToken: aws.String(""),
			},
			out: groupedInstances{
				"<missing_grp_tag>": instances{
					instance{
						instanceID: "i-123456789",
						tags: map[string]string{
							"orderByTag": "1",
						},
					},
				},
				"grp2": instances{
					instance{
						instanceID: "i-567891234",
						tags: map[string]string{
							"orderByTag": "1",
							"groupTag":   "grp2",
						},
					},
				},
			},
			event: StartStopEvent{
				GroupByTag: "groupTag",
				OrderByTag: "orderByTag",
				OrderBy: []string{
					"1",
					"2",
				},
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			gott := startStopBuilder(&tc.input, tc.event)
			if !reflect.DeepEqual(gott, tc.out) {
				t.Fatalf("gott: %v, wanted: %v", gott, tc.out)
			}
		})
	}
}

func TestInstances(t *testing.T) {
	tt := []struct {
		name      string
		instances instances
		expected  orderedInstances
		orderTag  string
		order     []string
	}{
		{
			"all-ordertag-and-tag-two",
			instances{
				instance{
					instanceID: "i-123456789",
					tags:       map[string]string{"one": "two"},
				},
				instance{
					instanceID: "i-987654321",
					tags:       map[string]string{"one": "two"},
				},
			},
			orderedInstances{
				"two": instances{
					instance{
						instanceID: "i-123456789",
						tags:       map[string]string{"one": "two"},
					},
					instance{instanceID: "i-987654321",
						tags: map[string]string{"one": "two"},
					},
				},
			},
			"one",
			[]string{"two", "one"},
		},
		{
			"all-ordertag-missing-four",
			instances{
				instance{
					instanceID: "i-123456789",
					tags:       map[string]string{"one": "two"},
				},
				instance{
					instanceID: "i-987654321",
					tags:       map[string]string{"one": "four"},
				},
			},
			orderedInstances{
				"two": instances{
					instance{
						instanceID: "i-123456789",
						tags:       map[string]string{"one": "two"},
					},
				},
				"<no-order>": instances{
					instance{instanceID: "i-987654321",
						tags: map[string]string{"one": "four"},
					},
				},
			},
			"one",
			[]string{"two", "one"},
		},
		{
			"one-ordertag-one-no-ordertag",
			instances{
				instance{
					instanceID: "i-123456789",
					tags:       map[string]string{"one": "one"},
				},
				instance{
					instanceID: "i-987654321",
					tags:       map[string]string{"two": "two"},
				},
			},
			orderedInstances{
				"one": instances{
					instance{
						instanceID: "i-123456789",
						tags:       map[string]string{"one": "one"},
					},
				},
				"<no-order>": instances{
					instance{instanceID: "i-987654321",
						tags: map[string]string{"two": "two"},
					},
				},
			},
			"one",
			[]string{"two", "one"},
		},
		{
			"all-no-ordertag",
			instances{
				instance{
					instanceID: "i-123456789",
					tags:       map[string]string{"two": "one"},
				},
				instance{
					instanceID: "i-987654321",
					tags:       map[string]string{"two": "two"},
				},
			},
			orderedInstances{
				"<no-order>": instances{
					instance{
						instanceID: "i-123456789",
						tags:       map[string]string{"two": "one"},
					},
					instance{
						instanceID: "i-987654321",
						tags:       map[string]string{"two": "two"},
					},
				},
			},
			"one",
			[]string{"two", "one"},
		},
		{
			"all-ordertag-and-tag-two",
			instances{
				instance{
					instanceID: "i-123456789",
					tags:       map[string]string{"one": "two"},
				},
				instance{
					instanceID: "i-987654321",
					tags:       map[string]string{"one": "two"},
				},
				instance{
					instanceID: "i-2222222222",
					tags:       map[string]string{"one": "one"},
				},
				instance{
					instanceID: "i-432198765",
					tags:       map[string]string{"one": "two"},
				},
				instance{
					instanceID: "i-111111111",
					tags:       map[string]string{"one": "one"},
				},
			},
			orderedInstances{
				"two": instances{
					instance{
						instanceID: "i-123456789",
						tags:       map[string]string{"one": "two"},
					},
					instance{
						instanceID: "i-987654321",
						tags:       map[string]string{"one": "two"},
					},
					instance{
						instanceID: "i-432198765",
						tags:       map[string]string{"one": "two"},
					},
				},
				"one": instances{
					instance{
						instanceID: "i-2222222222",
						tags:       map[string]string{"one": "one"},
					},
					instance{
						instanceID: "i-111111111",
						tags:       map[string]string{"one": "one"},
					},
				},
			},
			"one",
			[]string{"two", "one"},
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			order := tc.instances.byOrder(tc.orderTag, tc.order)

			if !reflect.DeepEqual(order, tc.expected) {
				t.Fatalf("want: %v, gott: %v", tc.expected, order)
			}
		})
	}
}
