package api

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	// DefaultUserAgent is used for API and content-fetching requests.
	DefaultUserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"

	// MaxResponseBody is the maximum response body size we'll read (4 MB).
	MaxResponseBody = 4 << 20

	// MaxContentBody is the maximum content body size we'll read (8 MB).
	MaxContentBody = 8 << 20
)

// NewHTTPClient creates an HTTP client with the given timeout, using the default
// transport with HTTP/2 and proxy support.
func NewHTTPClient(timeout time.Duration) *http.Client {
	t, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		return &http.Client{Timeout: timeout}
	}
	transport := t.Clone()
	transport.Proxy = http.ProxyFromEnvironment
	transport.ForceAttemptHTTP2 = true
	return &http.Client{
		Timeout:   timeout,
		Transport: transport,
	}
}

// NewSafeContentClient creates an HTTP client that blocks requests to private/local
// IP addresses. Use for fetching content from untrusted URLs.
func NewSafeContentClient(timeout time.Duration) *http.Client {
	var transport *http.Transport
	if base, ok := http.DefaultTransport.(*http.Transport); ok {
		transport = base.Clone()
	} else {
		transport = &http.Transport{}
	}
	transport.Proxy = nil
	transport.ForceAttemptHTTP2 = true

	dialer := &net.Dialer{Timeout: 15 * time.Second, KeepAlive: 30 * time.Second}
	transport.DialContext = func(ctx context.Context, network, address string) (net.Conn, error) {
		host, port, err := net.SplitHostPort(address)
		if err != nil {
			return nil, err
		}

		if ip := net.ParseIP(host); ip != nil {
			if IsBlockedIP(ip) {
				return nil, fmt.Errorf("blocked private or local IP address: %s", ip)
			}
			return dialer.DialContext(ctx, network, address)
		}

		ips, err := net.DefaultResolver.LookupIPAddr(ctx, host)
		if err != nil {
			return nil, err
		}

		allowed := make([]string, 0, len(ips))
		for _, ipAddr := range ips {
			if !IsBlockedIP(ipAddr.IP) {
				allowed = append(allowed, ipAddr.IP.String())
			}
		}
		if len(allowed) == 0 {
			return nil, fmt.Errorf("blocked host %q: resolves to private or local IP", host)
		}

		var lastErr error
		for _, ip := range allowed {
			conn, err := dialer.DialContext(ctx, network, net.JoinHostPort(ip, port))
			if err == nil {
				return conn, nil
			}
			lastErr = err
		}
		if lastErr != nil {
			return nil, lastErr
		}
		return nil, fmt.Errorf("failed to dial host %q", host)
	}

	client := &http.Client{
		Timeout:   timeout,
		Transport: transport,
	}
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if len(via) >= 10 {
			return errors.New("stopped after 10 redirects")
		}
		_, err := ValidateRemoteFetchURL(req.URL.String())
		return err
	}
	return client
}

// DoAPIRequest executes an HTTP request and returns the response body bytes.
// It checks the status code and parses API errors on non-2xx responses.
func DoAPIRequest(client *http.Client, req *http.Request) ([]byte, error) {
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, MaxResponseBody))
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, ParseAPIError(resp.StatusCode, body)
	}

	return body, nil
}

// ValidateRemoteFetchURL validates that a URL is safe for fetching content.
func ValidateRemoteFetchURL(rawURL string) (*url.URL, error) {
	u, err := url.ParseRequestURI(strings.TrimSpace(rawURL))
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, fmt.Errorf("invalid URL scheme %q (only http/https are allowed)", u.Scheme)
	}
	host := u.Hostname()
	if host == "" {
		return nil, errors.New("invalid URL: missing hostname")
	}
	if ip := net.ParseIP(host); ip != nil && IsBlockedIP(ip) {
		return nil, fmt.Errorf("blocked private or local IP address: %s", ip)
	}
	return u, nil
}

// IsBlockedIP returns true if the IP is a private, loopback, or otherwise
// unsafe address that should not be fetched from untrusted input.
func IsBlockedIP(ip net.IP) bool {
	if ip == nil {
		return true
	}
	if ip.IsLoopback() || ip.IsUnspecified() || ip.IsMulticast() ||
		ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsPrivate() {
		return true
	}
	if ip4 := ip.To4(); ip4 != nil {
		if ip4[0] == 0 || ip4[0] >= 224 {
			return true
		}
		if ip4[0] == 100 && ip4[1] >= 64 && ip4[1] <= 127 {
			return true // 100.64.0.0/10
		}
	}
	return false
}
