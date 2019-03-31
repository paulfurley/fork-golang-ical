package ics

import (
	"fmt"
	"io"
	"regexp"
	"strings"
	"testing"
	"time"
)

func TestCalendarSerialize(t *testing.T) {
	now := time.Date(2019, 3, 31, 9, 35, 45, 0, time.UTC)
	midday := time.Date(2019, 3, 31, 12, 0, 0, 0, time.UTC)
	one := time.Date(2019, 3, 31, 13, 0, 0, 0, time.UTC)

	cal := NewCalendar()
	cal.SetMethod(MethodRequest)

	event := cal.AddEvent("fake-id")
	event.SetCreatedTime(now)
	event.SetDtStampTime(now)
	event.SetModifiedAt(now)
	event.SetStartAt(midday)
	event.SetEndAt(one)
	event.SetSummary("Test event")
	event.SetLocation("Test address")
	event.SetDescription("simple single-line description")
	event.SetURL("https://example.com/this-is-a-long-url-that-should-hit-the-75-octet-fold-limit")
	event.SetOrganizer("sender@domain", WithCN("This Machine"))
	// TODO: use AddAttendee and fix non-deterministic order of serializing caused by use of map
	// event.AddAttendee("nobody@example.com", CalendarUserTypeIndividual, ParticipationStatusNeedsAction, ParticipationRoleReqParticipant, WithRSVP(true))
	got := cal.Serialize()

	expected := "BEGIN:VCALENDAR\r\n" +
		"VERSION:2.0\r\n" +
		"PRODID:-//Arran Ubels//Golang ICS library\r\n" +
		"METHOD:REQUEST\r\n" +
		"BEGIN:VEVENT\r\n" +
		"UID:fake-id\r\n" +
		"CREATED:20190331T093545Z\r\n" +
		"DTSTAMP:20190331T093545Z\r\n" +
		"LAST-MODIFIED:20190331T093545Z\r\n" +
		"DTSTART:20190331T120000Z\r\n" +
		"DTEND:20190331T130000Z\r\n" +
		"SUMMARY:Test event\r\n" +
		"LOCATION:Test address\r\n" +
		"DESCRIPTION:simple single-line description\r\n" +
		"URL:https://example.com/this-is-a-long-url-that-should-hit-the-75-octet-fol\r\n" +
		" d-limit\r\n" +
		"ORGANIZER;CN=This Machine:sender@domain\r\n" +
		"END:VEVENT\r\n" +
		"END:VCALENDAR\r\n"

	assertEqual(t, expected, got)
}

func TestCalendarStream(t *testing.T) {
	i := `
ATTENDEE;RSVP=TRUE;ROLE=REQ-PARTICIPANT;CUTYPE=GROUP:
 mailto:employee-A@example.com
DESCRIPTION:Project XYZ Review Meeting
CATEGORIES:MEETING
CLASS:PUBLIC
`
	expected := []ContentLine{
		ContentLine("ATTENDEE;RSVP=TRUE;ROLE=REQ-PARTICIPANT;CUTYPE=GROUP:mailto:employee-A@example.com"),
		ContentLine("DESCRIPTION:Project XYZ Review Meeting"),
		ContentLine("CATEGORIES:MEETING"),
		ContentLine("CLASS:PUBLIC"),
	}
	c := NewCalendarStream(strings.NewReader(i))
	cont := true
	for i := 0; cont; i++ {
		l, err := c.ReadLine()
		if err != nil {
			switch err {
			case io.EOF:
				cont = false
			default:
				t.Logf("Unknown error; %v", err)
				t.Fail()
				return
			}
		}
		if l == nil {
			if err == io.EOF && i == len(expected) {
				cont = false
			} else {
				t.Logf("Nil response...")
				t.Fail()
				return
			}
		}
		if i < len(expected) {
			if string(*l) != string(expected[i]) {
				t.Logf("Got %s expected %s", string(*l), string(expected[i]))
				t.Fail()
			}
		} else if l != nil {
			t.Logf("Larger than expected")
			t.Fail()
			return
		}
	}
}

func TestParseCalendarForRfc5545Sec4Examples(t *testing.T) {
	inputs := []string{
		`
BEGIN:VCALENDAR
PRODID:-//xyz Corp//NONSGML PDA Calendar Version 1.0//EN
VERSION:2.0
BEGIN:VEVENT
DTSTAMP:19960704T120000Z
UID:uid1@example.com
ORGANIZER:mailto:jsmith@example.com
DTSTART:19960918T143000Z
DTEND:19960920T220000Z
STATUS:CONFIRMED
CATEGORIES:CONFERENCE
SUMMARY:Networld+Interop Conference
DESCRIPTION:Networld+Interop Conference
  and Exhibit\nAtlanta World Congress Center\n
 Atlanta\, Georgia
END:VEVENT
END:VCALENDAR
`,
		`BEGIN:VCALENDAR
PRODID:-//RDU Software//NONSGML HandCal//EN
VERSION:2.0
BEGIN:VTIMEZONE
TZID:America/New_York
BEGIN:STANDARD
DTSTART:19981025T020000
TZOFFSETFROM:-0400
TZOFFSETTO:-0500
TZNAME:EST
END:STANDARD
BEGIN:DAYLIGHT
DTSTART:19990404T020000
TZOFFSETFROM:-0500
TZOFFSETTO:-0400
TZNAME:EDT
END:DAYLIGHT
END:VTIMEZONE
BEGIN:VEVENT
DTSTAMP:19980309T231000Z
UID:guid-1.example.com
ORGANIZER:mailto:mrbig@example.com
ATTENDEE;RSVP=TRUE;ROLE=REQ-PARTICIPANT;CUTYPE=GROUP:
 mailto:employee-A@example.com
DESCRIPTION:Project XYZ Review Meeting
CATEGORIES:MEETING
CLASS:PUBLIC
CREATED:19980309T130000Z
SUMMARY:XYZ Project Review
DTSTART;TZID=America/New_York:19980312T083000
DTEND;TZID=America/New_York:19980312T093000
LOCATION:1CP Conference Room 4350
END:VEVENT
END:VCALENDAR
`,
		`BEGIN:VCALENDAR
VERSION:2.0
PRODID:-//ABC Corporation//NONSGML My Product//EN
BEGIN:VTODO
DTSTAMP:19980130T134500Z
SEQUENCE:2
UID:uid4@example.com
ORGANIZER:mailto:unclesam@example.com
ATTENDEE;PARTSTAT=ACCEPTED:mailto:jqpublic@example.com
DUE:19980415T000000
STATUS:NEEDS-ACTION
SUMMARY:Submit Income Taxes
BEGIN:VALARM
ACTION:AUDIO
TRIGGER:19980403T120000Z
ATTACH;FMTTYPE=audio/basic:http://example.com/pub/audio-
 files/ssbanner.aud
REPEAT:4
DURATION:PT1H
END:VALARM
END:VTODO
END:VCALENDAR
`,
		`BEGIN:VCALENDAR
VERSION:2.0
PRODID:-//ABC Corporation//NONSGML My Product//EN
BEGIN:VJOURNAL
DTSTAMP:19970324T120000Z
UID:uid5@example.com
ORGANIZER:mailto:jsmith@example.com
STATUS:DRAFT
CLASS:PUBLIC
CATEGORIES:Project Report,XYZ,Weekly Meeting
DESCRIPTION:Project xyz Review Meeting Minutes\n
 Agenda\n1. Review of project version 1.0 requirements.\n2.
  Definition
 of project processes.\n3. Review of project schedule.\n
 Participants: John Smith\, Jane Doe\, Jim Dandy\n-It was
  decided that the requirements need to be signed off by
  product marketing.\n-Project processes were accepted.\n
 -Project schedule needs to account for scheduled holidays
  and employee vacation time. Check with HR for specific
  dates.\n-New schedule will be distributed by Friday.\n-
 Next weeks meeting is cancelled. No meeting until 3/23.
END:VJOURNAL
END:VCALENDAR`,
		`BEGIN:VCALENDAR
VERSION:2.0
PRODID:-//RDU Software//NONSGML HandCal//EN
BEGIN:VFREEBUSY
ORGANIZER:mailto:jsmith@example.com
DTSTART:19980313T141711Z
DTEND:19980410T141711Z
FREEBUSY:19980314T233000Z/19980315T003000Z
FREEBUSY:19980316T153000Z/19980316T163000Z
FREEBUSY:19980318T030000Z/19980318T040000Z
URL:http://www.example.com/calendar/busytime/jsmith.ifb
END:VFREEBUSY
END:VCALENDAR
`,
	}

	rnReplace := regexp.MustCompile("\r?\n")
	for i, input := range inputs {
		t.Run(fmt.Sprintf("RFC 5455 example #%d", i), func(t *testing.T) {

			input = rnReplace.ReplaceAllString(input, "\r\n")
			structure, err := ParseCalendar(strings.NewReader(input))
			assertNoError(t, err)

			if structure == nil {
				t.Fatalf("got nil error, but empty return value from ParseCalendar")
			}
		})
	}
}

func assertEqual(t *testing.T, expected string, got string) {
	if got != expected {
		t.Errorf("\n--- expected ---\n%s\n--- got ---\n%s\n---\n", expected, got)
	}
}

// assertNoError tests that got is nil and calls t.Fatal if not
func assertNoError(t *testing.T, got error) {
	t.Helper()
	if got != nil {
		t.Fatalf("got an error but didnt want one '%s'", got)
	}
}
