package ics

import "testing"

func TestPropertyParse(t *testing.T) {
	tests := []struct {
		Input    string
		Expected func(output *BaseProperty) bool
	}{
		{Input: "ATTENDEE;RSVP=TRUE;ROLE=REQ-PARTICIPANT;CUTYPE=GROUP:mailto:employee-A@example.com", Expected: func(output *BaseProperty) bool {
			return output.IANAToken == "ATTENDEE" && output.Value == "mailto:employee-A@example.com"
		}},
		{Input: "ATTENDEE;RSVP=\"TRUE\";ROLE=REQ-PARTICIPANT;CUTYPE=GROUP:mailto:employee-A@example.com", Expected: func(output *BaseProperty) bool {
			return output.IANAToken == "ATTENDEE" && output.Value == "mailto:employee-A@example.com"
		}},
		{Input: "ATTENDEE;RSVP=T\"RUE\";ROLE=REQ-PARTICIPANT;CUTYPE=GROUP:mailto:employee-A@example.com", Expected: func(output *BaseProperty) bool { return output == nil }},
	}
	for i, test := range tests {
		output := ParseProperty(ContentLine(test.Input))
		if !test.Expected(output) {
			t.Logf("Got: %#v", output)
			t.Logf("Failed %d %#v", i, test)
			t.Fail()
		}
	}
}

func TestPropertySerialize(t *testing.T) {
	tests := []struct {
		property BaseProperty
		expected string
	}{
		{
			property: BaseProperty{
				IANAToken:      "DESCRIPTION",
				ICalParameters: map[string][]string{},
				Value:          "foo",
			},
			expected: "DESCRIPTION:foo\r\n",
		},
		{
			property: BaseProperty{
				IANAToken:      "FOO",
				ICalParameters: map[string][]string{},
				Value:          "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			},
			expected: "FOO:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa\r\n a\r\n",
		},
	}

	for _, test := range tests {
		t.Run(test.expected, func(t *testing.T) {
			buf := bytes.NewBuffer(nil)
			test.property.serialize(buf)
			got := buf.String()

			assertEqual(t, test.expected, got)
		})
	}

}
