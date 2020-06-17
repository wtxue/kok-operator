package hosts

import "io/ioutil"

// LocalHosts for local hosts
type LocalHosts struct {
	Host string
	File string
}

// Set sets hosts
func (h *LocalHosts) Set(ip string) error {
	if h.File == "" {
		h.File = hostFile()
	}
	data, err := ioutil.ReadFile(h.File)
	if err != nil {
		return err
	}
	data, err = setHosts(data, h.Host, ip)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(h.File, data, 0644)
}
