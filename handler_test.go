package mock

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	. "gopkg.in/check.v1"
)

type MockSuite struct{}

var (

	// mux is the HTTP request multiplexer used with the test server
	mux *http.ServeMux

	// server is a test HTTP server used to provide mack API responses
	server *httptest.Server

	_ = Suite(&MockSuite{})
)

func Test(t *testing.T) { TestingT(t) }

func (s *MockSuite) SetUpTest(c *C) {
	mux = http.NewServeMux()
	server = httptest.NewServer(mux)
}

func (s *MockSuite) TearDownTest(c *C) {
	server.Close()
}

// Test that an attempt to construct a simulator with no events returns an error
func (s *MockSuite) TestCreateSimulatorWithNoEventsReturnsError(c *C) {
	stream := "noevents-stream"
	es := CreateTestEvents(0, stream, server.URL, "EventTypeY")

	handler, err := NewAtomFeedSimulator(es, nil, nil, 0)

	c.Assert(err, NotNil)
	c.Assert(err, DeepEquals, errors.New("Must provide one or more events."))
	c.Assert(handler, IsNil)
}

func (s *MockSuite) TestGetEventResponse(c *C) {
	stream := "astream-54"
	es := CreateTestEvents(1, stream, server.URL, "EventTypeA")
	e := es[0]

	b, err := json.Marshal(e)
	raw := json.RawMessage(b)

	timeStr := Time(time.Now())

	want := &EventAtomResponse{
		Title:   fmt.Sprintf("%d@%s", e.EventNumber, stream),
		ID:      e.Links[0].URI,
		Updated: timeStr,
		Summary: e.EventType,
		Content: &raw,
	}

	got, err := CreateTestEventAtomResponse(e, &timeStr)
	c.Assert(err, IsNil)
	c.Assert(got, DeepEquals, want)
}

func (s *MockSuite) TestResolveEvent(c *C) {
	stream := "astream5"
	es := CreateTestEvents(100, stream, server.URL, "EventTypeX")
	eu := fmt.Sprintf("%s/streams/%s/%d/", server.URL, stream, 9)

	got, err := resolveEvent(es, eu)

	c.Assert(err, IsNil)
	c.Assert(got, DeepEquals, es[9])
}

func (s *MockSuite) TestGetSliceSectionForwardFromZero(c *C) {
	es := CreateTestEvents(15, "x", "x", "x")

	sl, isF, isL, isH := getSliceSection(es, 0, 10, "forward")

	c.Assert(sl, HasLen, 10)
	c.Assert(isF, Equals, false)
	c.Assert(isL, Equals, true)
	c.Assert(isH, Equals, false)
	c.Assert(sl[0].EventNumber, Equals, 0)
	c.Assert(sl[len(sl)-1].EventNumber, Equals, 9)
}

//Testing a slice from the middle of the strem not exceeding any bounds.
func (s *MockSuite) TestGetSliceSectionForward(c *C) {
	es := CreateTestEvents(100, "x", "x", "x")

	se, isF, isL, isH := getSliceSection(es, 25, 50, "forward")

	c.Assert(se, HasLen, 50)
	c.Assert(isF, Equals, false)
	c.Assert(isL, Equals, false)
	c.Assert(isH, Equals, false)

	c.Assert(se[0].EventNumber, Equals, 25)
	c.Assert(se[len(se)-1].EventNumber, Equals, 74)
}

//Testing a slice from the middle of the stream not exceeding any bounds
func (s *MockSuite) TestGetSliceSectionBackward(c *C) {
	es := CreateTestEvents(100, "x", "x", "x")

	se, isF, isL, isH := getSliceSection(es, 75, 50, "backward")

	c.Assert(se, HasLen, 50)
	c.Assert(isF, Equals, false)
	c.Assert(isL, Equals, false)
	c.Assert(isH, Equals, false)
	c.Assert(se[0].EventNumber, Equals, 26)
	c.Assert(se[len(se)-1].EventNumber, Equals, 75)
}

//Version number is in range, but page number means the set will exceed
//the number of events in the stream.
func (s *MockSuite) TestGetSliceSectionBackwardUnder(c *C) {
	es := CreateTestEvents(100, "x", "x", "x")

	se, isF, isL, isH := getSliceSection(es, 25, 50, "backward")

	c.Assert(se, HasLen, 26)
	c.Assert(isF, Equals, false)
	c.Assert(isL, Equals, true)
	c.Assert(isH, Equals, false)
	c.Assert(se[0].EventNumber, Equals, 0)
	c.Assert(se[len(se)-1].EventNumber, Equals, 25)
}

//Testing the case where the version may be over the
//size of the highest version. This will happen when
//polling the head of the stream waiting for changes
func (s *MockSuite) TestGetSliceSectionForwardOut(c *C) {
	es := CreateTestEvents(100, "x", "x", "x")

	se, isF, isL, isH := getSliceSection(es, 101, 50, "forward")

	c.Assert(se, HasLen, 0)
	c.Assert(isF, Equals, true)
	c.Assert(isL, Equals, false)
	c.Assert(isH, Equals, true)
}

// Version number is in range but version plus pagesize is greter the the highest
// event number and so the query exeeds the number of results that can be returned
func (s *MockSuite) TestGetSliceSectionForwardOver(c *C) {
	es := CreateTestEvents(100, "x", "x", "x")

	se, isF, isL, isH := getSliceSection(es, 75, 50, "forward")
	c.Assert(se, HasLen, 25)
	c.Assert(isF, Equals, true)
	c.Assert(isL, Equals, false)
	c.Assert(isH, Equals, true)
	c.Assert(se[0].EventNumber, Equals, 75)
	c.Assert(se[len(se)-1].EventNumber, Equals, 99)
}

// This test covers the case where the version is higher than the highest version
func (s *MockSuite) TestGetSliceSectionTail(c *C) {
	es := CreateTestEvents(100, "x", "x", "x")

	se, isF, isL, isH := getSliceSection(es, 100, 20, "forward")

	c.Assert(se, HasLen, 0)
	c.Assert(isF, Equals, true)
	c.Assert(isL, Equals, false)
	c.Assert(isH, Equals, true)
}

func (s *MockSuite) TestGetSliceSectionAllForward(c *C) {
	es := CreateTestEvents(100, "x", "x", "x")

	se, isF, isL, isH := getSliceSection(es, 0, 100, "forward")

	c.Assert(se, HasLen, 100)
	c.Assert(isF, Equals, true)
	c.Assert(isL, Equals, true)
	c.Assert(isH, Equals, true)
	c.Assert(se[0].EventNumber, Equals, 0)
	c.Assert(se[len(se)-1].EventNumber, Equals, 99)
}

func (s *MockSuite) TestParseURLVersioned(c *C) {
	srv := "http://localhost:2113"
	stream := "An-Qw3334rd-St333"
	ver := 50
	direction := "backward"
	pageSize := 10

	url := fmt.Sprintf("%s/streams/%s/%d/%s/%d", srv, stream, ver, direction, pageSize)

	er, err := parseURL(url)

	c.Assert(err, IsNil)
	c.Assert(er.Host, Equals, srv)
	c.Assert(er.Stream, Equals, stream)
	c.Assert(er.Version, Equals, ver)
	c.Assert(er.Direction, Equals, direction)
	c.Assert(er.PageSize, Equals, pageSize)
}

func (s *MockSuite) TestParseURLInvalidVersion(c *C) {
	srv := "http://localhost:2113"
	stream := "An-Qw3334rd-St333"
	pageSize := 20
	direction := "backward"
	version := -1
	url := fmt.Sprintf("%s/streams/%s/%d/%s/%d", srv, stream, version, direction, pageSize)

	_, err := parseURL(url)

	c.Assert(err, FitsTypeOf, errInvalidVersion(version))
}

func (s *MockSuite) TestParseURLBase(c *C) {
	srv := "http://localhost:2113"
	stream := "An-Qw3334rd-St333"
	pageSize := 20
	direction := "backward"

	url := fmt.Sprintf("%s/streams/%s", srv, stream)

	er, err := parseURL(url)

	c.Assert(err, IsNil)
	c.Assert(er.Host, Equals, srv)
	c.Assert(er.Stream, Equals, stream)
	c.Assert(er.Version, Equals, 0)
	c.Assert(er.Direction, Equals, direction)
	c.Assert(er.PageSize, Equals, pageSize)
}

func (s *MockSuite) TestParseURLHead(c *C) {
	srv := "http://localhost:2113"
	stream := "An-Qw3334rd-St333"
	direction := "backward"
	pageSize := 100

	url := fmt.Sprintf("%s/streams/%s/%s/%s/%d", srv, stream, "head", direction, pageSize)

	er, err := parseURL(url)

	c.Assert(err, IsNil)
	c.Assert(er.Host, Equals, srv)
	c.Assert(er.Stream, Equals, stream)
	c.Assert(er.Version, Equals, 0)
	c.Assert(er.Direction, Equals, direction)
	c.Assert(er.PageSize, Equals, pageSize)
}

func (s *MockSuite) TestCreateFeedLinksBackward(c *C) {
	stream := "astream"
	ver := 50
	url := fmt.Sprintf("%s/streams/%s/%d/backward/20", server.URL, stream, ver)

	selfWant := fmt.Sprintf("%s/streams/%s", server.URL, stream)
	firstWant := fmt.Sprintf("%s/streams/%s/head/backward/20", server.URL, stream)
	lastWant := fmt.Sprintf("%s/streams/%s/0/forward/20", server.URL, stream)
	nextWant := fmt.Sprintf("%s/streams/%s/30/backward/20", server.URL, stream)
	prevWant := fmt.Sprintf("%s/streams/%s/51/forward/20", server.URL, stream)
	metaWant := fmt.Sprintf("%s/streams/%s/metadata", server.URL, stream)

	es := CreateTestEvents(100, stream, server.URL, "EventTypeX")
	m, _ := CreateTestFeed(es, url)

	var self, first, next, last, prev, meta bool
	for _, v := range m.Link {

		switch v.Rel {
		case "self":
			self = true
			c.Assert(v.Href, Equals, selfWant)
		case "first":
			first = true
			c.Assert(v.Href, Equals, firstWant)
		case "next":
			next = true
			c.Assert(v.Href, Equals, nextWant)
		case "last":
			last = true
			c.Assert(v.Href, Equals, lastWant)
		case "previous":
			prev = true
			c.Assert(v.Href, Equals, prevWant)
		case "metadata":
			meta = true
			c.Assert(v.Href, Equals, metaWant)
		}
	}

	c.Assert(self, Equals, true)
	c.Assert(first, Equals, true)
	c.Assert(next, Equals, true)
	c.Assert(last, Equals, true)
	c.Assert(prev, Equals, true)
	c.Assert(meta, Equals, true)
}

func (s *MockSuite) TestCreateFeedLinksLast(c *C) {
	stream := "astream"
	url := fmt.Sprintf("%s/streams/%s/0/forward/20", server.URL, stream)

	selfWant := fmt.Sprintf("%s/streams/%s", server.URL, stream)
	firstWant := fmt.Sprintf("%s/streams/%s/head/backward/20", server.URL, stream)
	prevWant := fmt.Sprintf("%s/streams/%s/20/forward/20", server.URL, stream)
	metaWant := fmt.Sprintf("%s/streams/%s/metadata", server.URL, stream)

	es := CreateTestEvents(100, stream, server.URL, "EventTypeX")
	m, _ := CreateTestFeed(es, url)

	var self, first, next, last, prev, meta bool
	for _, v := range m.Link {

		switch v.Rel {
		case "self":
			self = true
			c.Assert(v.Href, Equals, selfWant)
		case "first":
			first = true
			c.Assert(v.Href, Equals, firstWant)
		case "next":
			next = true
		case "last":
			last = true
		case "previous":
			prev = true
			c.Assert(v.Href, Equals, prevWant)
		case "metadata":
			meta = true
			c.Assert(v.Href, Equals, metaWant)
		}
	}

	c.Assert(self, Equals, true)
	c.Assert(first, Equals, true)
	c.Assert(next, Equals, false)
	c.Assert(last, Equals, false)
	c.Assert(prev, Equals, true)
	c.Assert(meta, Equals, true)
}

func (s *MockSuite) TestCreateFeedLinksTail(c *C) {
	stream := "astream"
	url := fmt.Sprintf("%s/streams/%s/100/forward/20", server.URL, stream)

	selfWant := fmt.Sprintf("%s/streams/%s", server.URL, stream)
	firstWant := fmt.Sprintf("%s/streams/%s/head/backward/20", server.URL, stream)
	lastWant := fmt.Sprintf("%s/streams/%s/0/forward/20", server.URL, stream)
	nextWant := fmt.Sprintf("%s/streams/%s/99/backward/20", server.URL, stream)
	metaWant := fmt.Sprintf("%s/streams/%s/metadata", server.URL, stream)

	es := CreateTestEvents(100, stream, server.URL, "EventTypeX")
	m, _ := CreateTestFeed(es, url)

	var self, first, next, last, prev, meta bool
	for _, v := range m.Link {

		switch v.Rel {
		case "self":
			self = true
			c.Assert(v.Href, Equals, selfWant)
		case "first":
			first = true
			c.Assert(v.Href, Equals, firstWant)
		case "next":
			next = true
			c.Assert(v.Href, Equals, nextWant)
		case "last":
			last = true
			c.Assert(v.Href, Equals, lastWant)
		case "previous":
			prev = true
		case "metadata":
			meta = true
			c.Assert(v.Href, Equals, metaWant)
		}
	}

	c.Assert(self, Equals, true)
	c.Assert(first, Equals, true)
	c.Assert(next, Equals, true)
	c.Assert(last, Equals, true)
	c.Assert(prev, Equals, false)
	c.Assert(meta, Equals, true)
}

func (s *MockSuite) TestCreateFeedLinksHead(c *C) {
	stream := "astream"
	url := fmt.Sprintf("%s/streams/%s/head/backward/20", server.URL, stream)

	selfWant := fmt.Sprintf("%s/streams/%s", server.URL, stream)
	firstWant := fmt.Sprintf("%s/streams/%s/head/backward/20", server.URL, stream)
	lastWant := fmt.Sprintf("%s/streams/%s/0/forward/20", server.URL, stream)
	nextWant := fmt.Sprintf("%s/streams/%s/79/backward/20", server.URL, stream)
	prevWant := fmt.Sprintf("%s/streams/%s/100/forward/20", server.URL, stream)
	metaWant := fmt.Sprintf("%s/streams/%s/metadata", server.URL, stream)

	es := CreateTestEvents(100, stream, server.URL, "EventTypeX")
	m, _ := CreateTestFeed(es, url)

	var self, first, next, last, prev, meta bool
	for _, v := range m.Link {
		switch v.Rel {
		case "self":
			self = true
			c.Assert(v.Href, Equals, selfWant)
		case "first":
			first = true
			c.Assert(v.Href, Equals, firstWant)
		case "next":
			next = true
			c.Assert(v.Href, Equals, nextWant)
		case "last":
			last = true
			c.Assert(v.Href, Equals, lastWant)
		case "previous":
			prev = true
			c.Assert(v.Href, Equals, prevWant)
		case "metadata":
			meta = true
			c.Assert(v.Href, Equals, metaWant)
		}
	}

	c.Assert(self, Equals, true)
	c.Assert(first, Equals, true)
	c.Assert(next, Equals, true)
	c.Assert(last, Equals, true)
	c.Assert(prev, Equals, true)
	c.Assert(meta, Equals, true)

}

func (s *MockSuite) TestCreateFeedEntriesLast(c *C) {
	stream := "astream"
	url := fmt.Sprintf("%s/streams/%s/0/forward/20", server.URL, stream)
	es := CreateTestEvents(100, stream, server.URL, "EventTypeX")
	m, _ := CreateTestFeed(es, url)

	c.Assert(m.Entry, HasLen, 20)
	for k, v := range m.Entry {
		num := (20 - 1) - k
		ti := fmt.Sprintf("%d@%s", num, stream)
		c.Assert(v.Title, Equals, ti)
	}
}

func (s *MockSuite) TestCreateFeedEntries(c *C) {
	stream := "astream"
	url := fmt.Sprintf("%s/streams/%s/20/forward/20", server.URL, stream)
	es := CreateTestEvents(100, stream, server.URL, "EventTypeX")
	m, _ := CreateTestFeed(es, url)

	c.Assert(m.Entry, HasLen, 20)
	for k, v := range m.Entry {
		num := (40 - 1) - k
		ti := fmt.Sprintf("%d@%s", num, stream)
		c.Assert(v.Title, Equals, ti)
	}
}

func (s *MockSuite) TestCreateFeedEntriesTail(c *C) {
	stream := "astream"
	url := fmt.Sprintf("%s/streams/%s/100/forward/20", server.URL, stream)
	es := CreateTestEvents(100, stream, server.URL, "EventTypeX")
	m, _ := CreateTestFeed(es, url)
	c.Assert(m.Entry, HasLen, 0)
}

func (s *MockSuite) TestCreateFeedEntriesHead(c *C) {
	stream := "astream"
	url := fmt.Sprintf("%s/streams/%s/head/backward/20", server.URL, stream)
	es := CreateTestEvents(100, stream, server.URL, "EventTypeX")
	m, _ := CreateTestFeed(es, url)

	c.Assert(m.Entry, HasLen, 20)
	for k, v := range m.Entry {
		num := (len(es) - 1) - k
		ti := fmt.Sprintf("%d@%s", num, stream)
		c.Assert(v.Title, Equals, ti)
	}
}

func (s *MockSuite) TestCreateEvents(c *C) {
	es := CreateTestEvents(100, "astream", server.URL, "EventTypeX")

	c.Assert(es, HasLen, 100)
	for i := 0; i <= 99; i++ {
		c.Assert(es[i].EventNumber, Equals, i)
	}
}

func (s *MockSuite) TestReverseSlice(c *C) {
	es := CreateTestEvents(100, "astream", server.URL, "EventTypeX")
	rs := reverseEventSlice(es)

	c.Assert(rs, HasLen, 100)
	top := len(es) - 1
	for i := 0; i <= top; i++ {
		c.Assert(rs[i].EventNumber, Equals, top-i)
	}
}

func (s *MockSuite) TestHeadOfStreamSetTrueWhenAtHeadOfStream(c *C) {
	stream := "astream"
	url := fmt.Sprintf("%s/streams/%s/90/forward/20", server.URL, stream)
	es := CreateTestEvents(100, stream, server.URL, "EventTypeX")
	m, _ := CreateTestFeed(es, url)

	c.Assert(m.HeadOfStream, Equals, true)
}

// If the result is the first page but has not exceeded the number of events
// the stream is not at the head of the stream. Only when the query exceeds the
// number of results is the reader at the head of the stream
func (s *MockSuite) TestHeadOfStreamSetFalseWhenNotAtHeadOfStream(c *C) {
	stream := "astream"
	url := fmt.Sprintf("%s/streams/%s/79/forward/20", server.URL, stream)
	es := CreateTestEvents(100, stream, server.URL, "EventTypeX")
	m, _ := CreateTestFeed(es, url)

	c.Assert(m.HeadOfStream, Equals, false)
}

func (s *MockSuite) TestSetStreamID(c *C) {
	stream := "some-stream"
	url := fmt.Sprintf("%s/streams/%s/90/forward/20", server.URL, stream)
	es := CreateTestEvents(100, stream, server.URL, "EventTypeX")
	m, _ := CreateTestFeed(es, url)

	c.Assert(m.StreamID, Equals, stream)
}
