package validation

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"github.com/wtxue/kube-on-kube-operator/pkg/util/ipallocator"
)

// IsHTTPSReachle tests that https://host:port is reachble in timeout.
func IsHTTPSReachle(host string, port int32, timeout time.Duration) error {
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout: timeout,
			}).DialContext,
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	url := fmt.Sprintf("https://%s:%d", host, port)
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	_, err = client.Do(request)
	if err != nil {
		return err
	}

	return nil
}

// IsSubNetOverlapped test if two subnets are overlapped
func IsSubNetOverlapped(net1, net2 *net.IPNet) error {
	if net1 == nil || net2 == nil {
		return nil
	}
	net1FirstIP, _ := ipallocator.GetFirstIP(net1)
	net1LastIP, _ := ipallocator.GetLastIP(net1)

	net2FirstIP, _ := ipallocator.GetFirstIP(net2)
	net2LastIP, _ := ipallocator.GetLastIP(net2)

	if net1.Contains(net2FirstIP) || net1.Contains(net2LastIP) ||
		net2.Contains(net1FirstIP) || net2.Contains(net1LastIP) {
		return errors.Errorf("subnet %v and %v are overlapped", net1, net2)
	}
	return nil
}
