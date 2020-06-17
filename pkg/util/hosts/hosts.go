package hosts

import (
	"fmt"
	"regexp"
)

const (
	linuxHostfile = "/etc/hosts"
)

// Hostser for hosts
type Hostser interface {
	Data() ([]byte, error)
	Set(ip string) error
}

func hostFile() string {
	return linuxHostfile
}

func setHosts(data []byte, host, ip string) ([]byte, error) {
	item := fmt.Sprintf("%s %s", ip, host)
	var re = regexp.MustCompile(fmt.Sprintf(".* %s", host))
	var newData string
	if re.Match(data) {
		newData = re.ReplaceAllString(string(data), item)
	} else {
		newData = fmt.Sprintf("%s\n%s\n", data, item)
	}

	return []byte(newData), nil
}
