package datasource

import (
	"crypto/tls"
	"net"
	"net/http"
	"sync"
	"time"
)

type DataSource struct {
	Id      int
	Name              string
	Type              string
	Url               string
	Password          string
	User              string
	Database          string
	BasicAuth         bool
	BasicAuthUser     string
	BasicAuthPassword string
	WithCredentials   bool
	IsDefault         bool
	ReadOnly          bool
	Created time.Time
	Updated time.Time
}

type proxyTransportCache struct {
	cache map[int]cachedTransport
	sync.Mutex
}

type cachedTransport struct {
	updated time.Time

	*http.Transport
}

type DataSourceParam struct {
	DataSource  *DataSource
	Host  string
	Token  string
}

var ptc = proxyTransportCache{
	cache: make(map[int]cachedTransport),
}

func (ds *DataSource) GetHttpClient() (*http.Client, error) {
	transport, err := ds.GetHttpTransport()

	if err != nil {
		return nil, err
	}

	return &http.Client{
		Timeout:   30 * time.Second,
		Transport: transport,
	}, nil
}

func (ds *DataSource) GetHttpTransport() (*http.Transport, error) {
	ptc.Lock()
	defer ptc.Unlock()

	if t, present := ptc.cache[ds.Id]; present && ds.Updated.Equal(t.updated) {
		return t.Transport, nil
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			//InsecureSkipVerify: tlsSkipVerify,
			InsecureSkipVerify: false,
			Renegotiation:      tls.RenegotiateFreelyAsClient,
		},
		Proxy: http.ProxyFromEnvironment,
		Dial: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).Dial,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
	}

	ptc.cache[ds.Id] = cachedTransport{
		Transport: transport,
		updated:   ds.Updated,
	}

	return transport, nil
}
