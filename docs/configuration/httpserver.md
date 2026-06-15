# HTTP Server

## Settings

```yaml
################################
##### HTTP SERVER SETTINGS #####
################################

# Settings for INCOMING http server connections received by GoToSocial.
#
# WARNING: all of the below are internal HTTP server configuration flags,
# you should not uncomment + tweak these unless you know what you're doing!
# They are set to sensible defaults and you shouldn't need to change them.
http-server:

  #######################################################################
  ##### Options passed directly to our HTTP router's gin.Engine{}: #####
  #######################################################################

  # MaxMultipartMemory value of 'maxMemory' param that is given to
  # http.Request's ParseMultipartForm method call.
  #
  # Default: 40MiB
  #max-multipart-memory: "40MiB"

  # Enable h2c (http2 without TLS).
  #
  # May be useful if your reverse proxy supports connecting to GoToSocial via HTTP2.
  #
  # Default: false
  #use-h2c: false

  #######################################################################
  ##### Options passed directly to our base http.Server{} instance: #####
  #######################################################################

  # ReadTimeout is the maximum duration for reading the entire request,
  # including the body. A zero or negative value means there will be no timeout.
  #
  # Because ReadTimeout does not let Handlers make per-request decisions on
  # each request body's acceptable deadline or upload rate, most users will prefer
  # to use ReadHeaderTimeout. It is valid to use them both.
  #
  # Default: 60s
  #read-timeout: "60s"

  # ReadHeaderTimeout is the amount of time allowed to read request headers.
  # The connection's read deadline is reset after reading the headers and the
  # Handler can decide what is considered too slow for the body. If zero, the
  # value of ReadTimeout is used. If negative, or if zero and ReadTimeout is
  # zero or negative, there is no timeout.
  #
  # Default: 0 (use read-timeout duration).
  #read-header-timeout: "0"

  # WriteTimeout is the maximum duration before timing out writes of the response.
  # It is reset whenever a new request's header is read. Like ReadTimeout, it does
  # not let Handlers make decisions on a per-request basis. A zero or negative value
  # means there will be no timeout.
  #
  # Default: 30s
  #write-timeout: "30s"

  # IdleTimeout is the maximum amount of time to wait for the next request when
  # keep-alives are enabled. If zero, the value of ReadTimeout is used. If negative,
  # or if zero and ReadTimeout is zero or negative, there is no timeout.
  #
  # Default: 30s
  #idle-timeout: "30s"

  # MaxHeaderBytes controls the maximum number of bytes the server will read parsing
  # the request header's keys and values, including the request line. It does not limit
  # the size of the request body. If zero, DefaultMaxHeaderBytes is used.
  #
  # Default: 0 (use built-in default 1048576, ie., 1 mebibyte)
  #max-header-bytes: "0"

  ############################################################################
  ##### Options passed directly to our HTTP server's http.HTTP2Config{}: #####
  ############################################################################

  # MaxConcurrentStreams optionally specifies the number of concurrent streams that a
  # peer may have open at a time. If zero, MaxConcurrentStreams defaults to at least 100.
  #
  # Default: 0 (at least 100)
  #max-concurrent-streams: "0"

  # MaxDecoderHeaderTableSize optionally specifies an upper limit for the size of the header
  # compression table used for decoding headers sent by the peer. A valid value is less than
  # 4MiB. If zero or invalid, a default value is used.
  #
  # Default: 0 (use built-in default 4096, ie., 4 kibibytes)
  #max-decoder-header-table-size: "0"

  # MaxEncoderHeaderTableSize optionally specifies an upper limit for the header compression
  # table used for sending headers to the peer. A valid value is less than 4MiB. If zero or
  # invalid, a default value is used.
  #
  # Default: 0 (use built-in default 4096, ie., 4 kibibytes)
  #max-encoder-header-table-size: "0"

  # MaxReadFrameSize optionally specifies the largest frame this endpoint is willing to read.
  # A valid value is between 16KiB and 16MiB, inclusive. If zero or invalid, a default value is used.
  #
  # Default: 0 (use built-in default 1048576, ie., 1 mebibyte)
  #max-read-frame-size: "0"

  # MaxReceiveBufferPerConnection is the maximum size of the flow control window for data received
  # on a connection. A valid value is at least 64KiB and less than 4MiB. If invalid, a default value
  # is used.
  #
  # Default: 0 (use built-in default 1048576, ie., 1 mebibyte)
  #max-receive-buffer-per-connection: "0"

  # MaxReceiveBufferPerStream is the maximum size of the flow control window for data received on a
  # stream (request). A valid value is less than 4MiB. If zero or invalid, a default value is used.
  #
  # Default: 0 (use built-in default 1048576, ie., 1 mebibyte)
  #max-receive-buffer-per-stream: "0"

  # SendPingTimeout is the timeout after which a health check using a ping frame will be carried out
  # if no frame is received on a connection. If zero, no health check is performed.
  #
  # Default: 0
  #send-ping-timeout: "0"

  # PingTimeout is the timeout after which a connection will be closed if a response to a ping is not
  # received. If zero, a default of 15 seconds is used.
  #
  # Default: 0 (use built-in default 15s)
  #ping-timeout: "0"

  # WriteByteTimeout is the timeout after which a connection will be closed if no data can be written to it.
  # The timeout begins when data is available to write, and is extended whenever any bytes are written.
  #
  # Default: 0
  #write-byte-timeout: "0"
```
