package hardware

import (
	"fmt"
	"runtime"

	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/mem"
)

func GetSystemSection() (string, error) {
	runTimeOs := runtime.GOOS

	vmStat, err := mem.VirtualMemory()
	if err != nil {
		return "", err
	}

	hostStat, err := host.Info()
	if err != nil {
		return "", err
	}

	output := fmt.Sprintf("Hostname: %s\nTotal Memory: %d\nUsed Memory: %d\nOS: %s", hostStat.Hostname, vmStat.Total, vmStat.Used, runTimeOs)
	return output, nil
}
