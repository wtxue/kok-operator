package supervisor

import (
	"fmt"
	"io"
	"path"

	"github.com/wtxue/kube-on-kube-operator/pkg/util/ssh"
)

const (
	DefaultSystemdUnitFilePath = "/etc/systemd/system"
)

type SystemdSupervisor struct {
	Name string
	SSH  ssh.Interface
}

func (s *SystemdSupervisor) Deploy(data io.Reader) error {
	unitFilePath := path.Join(DefaultSystemdUnitFilePath, fmt.Sprintf("%s.service", s.Name))
	if err := s.SSH.WriteFile(data, unitFilePath); err != nil {
		return err
	}

	cmd := fmt.Sprintf("systemctl -f enable %s", unitFilePath)
	if _, stderr, exit, err := s.SSH.Execf(cmd); err != nil || exit != 0 {
		return fmt.Errorf("exec %q failed:exit %d:stderr %s:error %s", cmd, exit, stderr, err)
	}

	cmd = "systemctl daemon-reload"
	if _, stderr, exit, err := s.SSH.Execf(cmd); err != nil || exit != 0 {
		return fmt.Errorf("exec %q failed:exit %d:stderr %s:error %s", cmd, exit, stderr, err)
	}

	return nil
}

func (s *SystemdSupervisor) Start() error {
	unitName := fmt.Sprintf("%s.service", s.Name)

	cmd := fmt.Sprintf("systemctl restart %s", unitName)
	if _, stderr, exit, err := s.SSH.Execf(cmd); err != nil || exit != 0 {
		cmd = fmt.Sprintf("journalctl --unit %s -n10 --no-pager", unitName)
		jStdout, _, jExit, jErr := s.SSH.Execf(cmd)
		if jErr != nil || jExit != 0 {
			return fmt.Errorf("exec %q:error %s", cmd, err)
		}
		fmt.Printf("log:\n%s", jStdout)

		return fmt.Errorf("Exec %s failed:exit %d:stderr %s:error %s:log:\n%s", cmd, exit, stderr, err, jStdout)
	}

	return nil
}
