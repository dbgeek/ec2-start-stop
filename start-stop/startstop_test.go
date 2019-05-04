package main

import (
	"reflect"
	"testing"
)

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
