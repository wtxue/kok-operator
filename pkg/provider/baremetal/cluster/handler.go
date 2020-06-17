package cluster

import (
	"fmt"
	"net/http"
)

func (p *Provider) ping(resp http.ResponseWriter, req *http.Request) {
	fmt.Fprint(resp, "pong")
}
