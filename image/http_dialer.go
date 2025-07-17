package image

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

func newStatistic() *Statistic {
	return &Statistic{}
}

type Statistic struct {
	downBytes atomic.Uint64
	upBytes   atomic.Uint64
	retry     atomic.Uint32
	dial      atomic.Uint32
	startOnce sync.Once
	startTime time.Time
}

func (s *Statistic) markDial() {
	s.startOnce.Do(func() { s.startTime = time.Now() })
	s.dial.Add(1)
}
func (s *Statistic) markRetry() {
	s.startOnce.Do(func() { s.startTime = time.Now() })
	s.dial.Add(1)
}

func (s *Statistic) GetStatistic() (downBytes, upBytes uint64, dial, retry uint32, startTime time.Time) {
	return s.downBytes.Load(), s.upBytes.Load(), s.dial.Load(), s.retry.Load(), s.startTime

}

type statisticConn struct {
	net.Conn
	stats *Statistic
}

func (c *statisticConn) Read(b []byte) (n int, err error) {
	n, err = c.Conn.Read(b)
	if n > 0 {
		c.stats.downBytes.Add(uint64(n))
	}
	return n, err
}

func (c *statisticConn) Write(b []byte) (n int, err error) {
	n, err = c.Conn.Write(b)
	if n > 0 {
		c.stats.upBytes.Add(uint64(n))
	}
	return n, err
}
func GenHTTPStatRounderTripper(proxy string) (*Statistic, http.RoundTripper, error) {
	stats := newStatistic()
	tripper, err := warpProxy(http.DefaultTransport.(*http.Transport).Clone(), proxy)
	if err != nil {
		return nil, nil, fmt.Errorf("warp proxy error: %w", err)
	}

	tripper.DialContext = warpStatisticRounderTripper(tripper.DialContext, stats)
	return stats, tripper, nil
}

func warpStatisticRounderTripper(
	dialContext func(ctx context.Context, network string, addr string) (net.Conn, error), stats *Statistic,
) func(ctx context.Context, network string, addr string) (net.Conn, error) {
	if dialContext == nil {
		dialContext = (&net.Dialer{Timeout: 30 * time.Second, KeepAlive: 30 * time.Second}).DialContext
	}
	return func(ctx context.Context, network string, addr string) (net.Conn, error) {
		conn, err := dialContext(ctx, network, addr)
		if err != nil {
			stats.markRetry()
			return nil, err
		}
		stats.markDial()
		return &statisticConn{Conn: conn, stats: stats}, nil
	}
}

func warpProxy(tripper *http.Transport, proxy string) (*http.Transport, error) {
	if len(proxy) == 0 {
		return tripper, nil
	}
	if !strings.HasPrefix(proxy, "http") {
		proxy = "http://" + proxy
	}
	proxyURL, err := url.Parse(proxy)
	if err != nil {
		return nil, fmt.Errorf("proxy is invalid: %w", err)
	}
	tripper.Proxy = http.ProxyURL(proxyURL)
	return tripper, nil
}
