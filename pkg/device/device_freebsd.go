package device

import (
	"fmt"
	"net"
	"os/exec"
)

func (d *BSDKernelDevice) AddRoute(dst net.IPNet, gw net.IP, table int) error {
	// TODO: Use proper gateway

	return exec.Command("setfib", fmt.Sprint(table), "route", "add", "-net", dst.String(), "-interface", d.Name()).Run()
}

func (d *BSDKernelDevice) DeleteRoute(dst net.IPNet, table int) error {
	return exec.Command("setfib", fmt.Sprint(table), "route", "delete", "-net", dst.String(), "-interface", d.Name()).Run()
}

func DetectMTU(ip net.IP) (int, error) {
	// TODO: Thats just a guess
	return 1500, nil
}

func DetectDefaultMTU() (int, error) {
	// TODO: Thats just a guess
	return 1500, nil
}
