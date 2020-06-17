package ssh

import (
	"strconv"
	"strings"
)

// GetNetworkInterface return network interface name by ip
func GetNetworkInterface(s Interface, ip string) string {
	stdout, _, _, _ := s.Execf("ip a | grep '%s' |awk '{print $NF}'", ip)

	return stdout
}

// Timestamp returns target node timestamp.
func Timestamp(s Interface) (int, error) {
	stdout, err := s.CombinedOutput("date +%s")
	if err != nil {
		return 0, err
	}

	return strconv.Atoi(strings.TrimSpace(string(stdout)))
}
