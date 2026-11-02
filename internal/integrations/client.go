package integrations

import (
	"net/http"
	"time"
)

// defaultHTTPClient is a shared, thread-safe HTTP client with connection pooling and standard timeouts.
var defaultHTTPClient = &http.Client{
	Timeout: 10 * time.Second,
}

