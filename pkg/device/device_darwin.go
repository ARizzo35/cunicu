package device

import (
	"net"
	"os/exec"

	"github.com/stv0g/cunicu/pkg/errors"
)

func (d *BSDKernelDevice) AddRoute(dst net.IPNet, table int) error {
	if table != 0 {
		return errors.ErrNotSupported
	}

	return exec.Command("route", "add", "-net", dst.String(), "-interface", d.Name()).Run()
}

func (d *BSDKernelDevice) DeleteRoute(dst net.IPNet, table int) error {
	if table != 0 {
		return errors.ErrNotSupported
	}

	return exec.Command("route", "delete", "-net", dst.String(), "-interface", d.Name()).Run()
}

func DetectMTU(ip net.IP) (int, error) {
	return -1, errors.ErrNotSupported
}

func DetectDefaultMTU() (int, error) {
	return -1, errors.ErrNotSupported
}
