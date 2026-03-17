package main

import (
	"compress/gzip"
	"expvar"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// We will create a new type right before we use it in our metrics middleware
type metricsResponseWriter struct {
    wrapped    http.ResponseWriter   // the original http.ResponseWriter
    statusCode int         // this will contain the status code we need
    headerWritten bool    // has the response headers already been written?
}

// Create an new instance of our custom http.ResponseWriter once
// we are provided with the original http.ResponseWriter. We will set
// the status code to 200 by default since that is what Golang does as well
// the headerWritten is false by default so no need to specify
func newMetricsResponseWriter(w http.ResponseWriter) *metricsResponseWriter {
    return &metricsResponseWriter {
        wrapped: w,
        statusCode: http.StatusOK,
    }
}

// Remember that the http.Header type is a map (key: value) of the headers
// Our custom http.ResponseWriter does not need to change the way the Header()
// method works, so all we do is call the original http.ResponseWriter's Header() 
// method when our custom http.ResponseWriter's Header() method is called
func (mw *metricsResponseWriter) Header() http.Header {
    return mw.wrapped.Header()
}

// Let's write the status code that is provided
// Again the original http.ResponseWriter's WriteHeader() methods knows
// how to do this
func (mw *metricsResponseWriter) WriteHeader(statusCode int) {
    mw.wrapped.WriteHeader(statusCode)
    // After the call to WriteHeader() returns, we record
    // the first status code for use in our metrics
    // NOTE: Because we only want the first status code sent, we will
    // ignore any other status code that gets written. For example,
    // mw.WriteHeader(404) followed by mw.WriteHeader(500). The client
    // will receive a 404, the 500 will never be sent
    if !mw.headerWritten {
        mw.statusCode = statusCode
        mw.headerWritten = true
    }
}

// The write() method simply calls the original http.ResponseWriter's
// Write() method which write the data to the connection
func (mw *metricsResponseWriter) Write(b []byte) (int, error) {
    mw.headerWritten = true
    return mw.wrapped.Write(b)
}

// We need a function to get the original http.ResponseWriter
func (mw *metricsResponseWriter) Unwrap() http.ResponseWriter {
    return mw.wrapped
}

// add an entry for recording our status code metrics
// this middleware will run for every request received
func (a *applicationDependencies) metrics (next http.Handler) http.Handler {                             
	// Setup our variable to track the metrics
	var (
		totalRequestsReceived = expvar.NewInt("total_requests_received")
		totalResponsesSent    = expvar.NewInt("total_responses_sent")
		totalProcessingTimeMicroseconds = expvar.NewInt("total_processing_time_μs")
		totalResponsesSentByStatus = expvar.NewMap("total_responses_sent_by_status")
	)

 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // start is when we receive the request and start processing it
        start := time.Now()
        // update our request received counter
        totalRequestsReceived.Add(1)
		// create a custom responseWriter
		mw := newMetricsResponseWriter(w)
		// we send our custom responseWriter down the middleware chain
		next.ServeHTTP(mw, r)
        // remember the middleware chain goes in both directions, so we can
        // do things when we return back to our middleware.We will increment
        // the responses sent counter
        totalResponsesSent.Add(1)
		// extract the status code for use in our metrics since we have returned 
		// from the middleware chain. The map uses strings so we need to convert the
		// status codes from their integer values to strings
		totalResponsesSentByStatus.Add(strconv.Itoa(mw.statusCode), 1)
		// calculate the processing time for this request. Remember we set start
        // at the beginning, so now since we are back in the middleware we can
        // compute the time taken
        duration := time.Since(start).Microseconds()
        totalProcessingTimeMicroseconds.Add(duration)
    })
}



func (a *applicationDependencies) logRequest(next http.Handler) http.Handler{
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a.logger.Info("request",
			"method", r.Method,
			"url", r.URL.String(),
			"remote_ip", r.RemoteAddr,
			"user_agent", r.UserAgent(),
			"time", time.Now().UTC(),
		)

		next.ServeHTTP(w, r)
	})
}

func (a *applicationDependencies) compressResponse(next http.Handler) http.Handler{
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		w.Header().Set("Content-Encoding", "gzip")

		gzw := gzip.NewWriter(w)
		defer gzw.Close()
		gzrw := gzipResponseWriter{Writer: gzw, ResponseWriter: w}

		next.ServeHTTP(gzrw, r)
	})
}

type gzipResponseWriter struct{
	http.ResponseWriter
	io.Writer
}

func (gzrw gzipResponseWriter) Write(data []byte) (int, error) {
	return gzrw.Writer.Write(data)
}

func (a *applicationDependencies) enableCORS (next http.Handler) http.Handler {                             
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.Header().Add("Vary", "Origin")
		// The request method can vary so don't rely on cache
		w.Header().Add("Vary", "Access-Control-Request-Method")
		origin := r.Header.Get("Origin")

		if origin != "" {
			for i := range a.config.cors.trustedOrigins {
				if origin == a.config.cors.trustedOrigins[i] {
					w.Header().Set("Access-Control-Allow-Origin", origin)  
					// check if it is a Preflight CORS request
					if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
						if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
							w.Header().Set("Access-Control-Allow-Methods", "OPTIONS, PUT, PATCH, DELETE")
							w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")

							// we need to send a 200 OK status. Also since there
							// is no need to continue the middleware chain we
							// we leave  - remember it is not a real 'comments' request but
							// only a preflight CORS request 
							w.WriteHeader(http.StatusOK)
							return
						}
						return
					}
					break
				}
			}
		}

        next.ServeHTTP(w, r)
    })
}

func (a *applicationDependencies) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// defer will be called when the stack unwinds
		defer func() {
			// recover() checks for panics
			err := recover()
			if err != nil {
				w.Header().Set("Connection", "close")
				a.serverErrorResponse(w, r, fmt.Errorf("%s", err))
				return
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// Delete all previous code in rateLimit. We will start from scratch
func (a *applicationDependencies) rateLimit(next http.Handler) http.Handler {
	// Define a rate limiter struct
	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time // remove map entries that are stale
	}
	var mu sync.Mutex                      // use to synchronize the map
	var clients = make(map[string]*client) // the actual map
	// A goroutine to remove stale entries from the map
	go func() {
		for {
			time.Sleep(time.Minute)
			mu.Lock() // begin cleanup
			// delete any entry not seen in three minutes
			for ip, client := range clients {
				if time.Since(client.lastSeen) > 3*time.Minute {
					delete(clients, ip)
				}
			}
			mu.Unlock() // finish clean up
		}
	}()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// we will wrap all our logic in an if statement
		if a.config.limiter.enabled {
			// get the IP address
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				a.serverErrorResponse(w, r, err)
				return
			}

			mu.Lock() // exclusive access to the map
			// check if ip address already in map, if not add it
			_, found := clients[ip]
			if !found {
				clients[ip] = &client{limiter: rate.NewLimiter(rate.Limit(a.config.limiter.rps), a.config.limiter.burst)}
			}

			// Update the last seen for the client
			clients[ip].lastSeen = time.Now()

			// Check the rate limit status
			if !clients[ip].limiter.Allow() {
				mu.Unlock() // no longer need exclusive access to the map
				a.rateLimitExceededResponse(w, r)
				return
			}

			mu.Unlock() // others are free to get exclusive access to the map
		}
		next.ServeHTTP(w, r)
	})
}