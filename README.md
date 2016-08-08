#GetEventStore Atom Feed Mock

The Atom feed handler provides an easy way to provide a mock 
atom feed when testing code that reads Atom feeds from GetEventStore.

While I was developing the Go.GetEventStore client for the GetEventStore 
HTTP API, I found that there was a lot of overhead in setting up mock data 
for tests. Creating Atom feeds as string literals was just too fiddly. 

Using the simulator you can set up tests running against realistic data in 
just a few lines of code.

The package also provides a number of fuctions for creating test events and metadata.

###Get the package

```go

    $ go get github.com/jetbasrawi/go.geteventstore.testfeed

```

While unit testing, the handler can be used with the test server in the "net/http/httptest" package

```go 

import(
    "net/http"
	"net/http/httptest"
	"net/url"
	"testing"

    "github.com/jetbasrawi/go.geteventstore.testfeed"
)

var (

	// mux is the HTTP request multiplexer used with the test server
	mux *http.ServeMux

	// server is a test HTTP server used to provide mock API responses
	server *httptest.Server

)

func setup() {
    // Initialize multiplexer
	mux = http.NewServeMux()

    // Initialize test server
	server = httptest.NewServer(mux)

    // Create 50 test events to be served by the mock feed handler
    // The events will be of the types specified in the variadic eventType argument
    es := mock.CreateTestEvents(50, "foostream", server.URL, "FooEventType", "BarEventType")

    // Create a new mock feed handler
    handler, err := mock.NewAtomFeedSimulator(es, u, m, -1)
	if err != nil {
		log.Fatal(err)
	}

    // Add the handler to the multiplexer
	mux.Handle("/", handler)

}

func teardown() {
    // Close the server
	server.Close()
}


```



