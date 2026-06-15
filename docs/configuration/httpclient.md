# HTTP Client

## Settings

```yaml
################################
##### HTTP CLIENT SETTINGS #####
################################

# Settings for OUTGOING http client connections used by GoToSocial to make
# requests to remote resources (status GETs, media GETs, inbox POSTs, etc).
http-client:

  # Duration. Timeout to use for outgoing HTTP requests. If the timeout
  # is exceeded, the connection to the remote server will be dropped.
  # A value of 0s indicates no timeout: this is not advised!
  # Examples: ["5s", "30s", "0s"]
  # Default: "1m"
  timeout: "1m"

  ########################################
  #### RESERVED IP RANGE EXCEPTIONS ######
  ########################################
  #
  # Explicitly allow or block outgoing dialing within the provided IPv4/v6 CIDR ranges.
  #
  # By default, as a basic security precaution, GoToSocial blocks outgoing dialing within most "special-purpose"
  # IP ranges. However, it may be desirable for admins with more exotic setups (proxies, funky NAT, etc) to
  # explicitly override one or more of these otherwise blocked ranges.
  #
  # Each of the below allow/block config options accepts an array of IPv4 and/or IPv6 CIDR strings.
  # For example, to override the hardcoded block of IPv4 and IPv6 dialing to localhost, set:
  #
  #   allow-ips: ["127.0.0.1/32", "::1/128"].
  #
  # You can also use YAML multi-line arrays to define these, but be diligent with indentation.
  #
  # When dialing, GoToSocial will first check if the destination falls within explicitly allowed IP ranges,
  # then explicitly blocked IP ranges, then the default (hardcoded) blocked IP ranges, returning OK on the
  # first allowed match, not OK on the first blocked match, or just defaulting to OK if nothing is matched.
  #
  # As with all security settings, it is better to start too restrictive and then ease off depending on
  # your use case, than to start too permissive and try to close the stable door after the horse has
  # already bolted. With this in mind:
  # - Don't touch these settings unless you have a good reason to, and only if you know what you're doing.
  # - When adding explicitly allowed exceptions, use the narrowest possible CIDR for your use case.
  #
  # For reserved / special ranges, see:
  # - https://www.iana.org/assignments/iana-ipv4-special-registry/iana-ipv4-special-registry.xhtml
  # - https://www.iana.org/assignments/iana-ipv6-special-registry/iana-ipv6-special-registry.xhtml
  #
  # Both allow-ips and block-ips default to an empty array.
  allow-ips: []
  block-ips: []

  # Bool. Disable verification of TLS certificates of remote servers.
  # With this set to 'true', GoToSocial will not error when a remote
  # server presents an invalid or self-signed certificate.
  #
  # THIS SETTING SHOULD BE USED FOR TESTING ONLY! IF YOU TURN THIS
  # ON WHILE RUNNING IN PRODUCTION YOU ARE LEAVING YOUR SERVER WIDE
  # OPEN TO MAN IN THE MIDDLE ATTACKS! DO NOT CHANGE THIS SETTING
  # UNLESS YOU KNOW EXACTLY WHAT YOU'RE DOING AND WHY YOU'RE DOING IT.
  #
  # Default: false
  tls-insecure-skip-verify: false

  # Bool. Sets outgoing queries to webfinger, host-meta and nodeinfo to use
  # HTTP instead of HTTPS.
  #
  # THIS SETTING SHOULD BE USED FOR TESTING ONLY! DO NOT CHANGE THIS SETTING
  # UNLESS YOU KNOW EXACTLY WHAT YOU'RE DOING AND WHY YOU'RE DOING IT.
  #
  # Default: false
  insecure-outgoing: false
  
  ##########################################################################
  ##### Options passed directly to the HTTP client's http.Transport{}: #####
  ##########################################################################

  # NOTE: all of the below are internal HTTP client configuration flags,
  # you should not uncomment + tweak these unless you know what you're doing!
  #
  # The commented-out values are set to the sensible values GtS uses by default.

  # DisableKeepAlives, if true, disables HTTP keep-alives and will only use the
  # connection to the server for a single HTTP request.
  #
  # This is unrelated to the similarly named TCP keep-alives.
  #
  # Default: false
  #disable-keep-alives: false

  # MaxConnsPerHost optionally limits the total number of connections per host,
  # including connections in the dialing, active, and idle states. On limit
  # violation, dials will block.
  #
  # Default: 6
  #max-conns-per-host: 6

  # MaxIdleConns controls the maximum number of idle (keep-alive) connections across
  # all hosts. Zero means no limit.
  #
  # Default: 0 (no limit)
  #max-idle-conns: 0
  
  # MaxIdleConnsPerHost, if non-zero, controls the maximum idle (keep-alive)
  # connections to keep per-host. If zero, DefaultMaxIdleConnsPerHost (2) is used.
  #
  # Default: 0 (use built-in default 2)
  #max-idle-conns-per-host: 0
  
  # IdleConnTimeout is the maximum amount of time an idle (keep-alive) connection
  # will remain idle before closing itself. Zero means no limit.
  #
  # Default: 5 minutes
  #idle-conn-timeout: "5 minutes"
  
  # TLSHandshakeTimeout specifies the maximum amount of time to wait for a TLS
  # handshake. Zero means no timeout.
  #
  # Default: 0 (no limit)
  #tls-handshake-timeout: "0"
  
  # ResponseHeaderTimeout, if non-zero, specifies the amount of time to wait for a
  # server's response headers after fully writing the request (including its body, if any).
  # This time does not include the time to read the response body.
  #
  # Default: 0 (no limit)
  #response-header-timeout: "0"
  
  # ReadBufferSize specifies the size of the read buffer used when reading from the transport.
  # If zero, a default (currently 4KB) is used.
  #
  # Default: 0 (use built-in default 4KB)
  #read-buffer-size: "0"
  
  # WriteBufferSize specifies the size of the write buffer used when writing to the transport.
  # If zero, a default (currently 4KB) is used.
  #
  # Default: 0 (use built-in default 4KB)
  #write-buffer-size: "0"
```
