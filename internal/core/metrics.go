package core

import (
	"time"

	"github.com/rockpanel/rockpanel/pkg/types"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
)

var (
	prevNetBytes  [2]int64
	prevNetTime   time.Time
)

func CollectMetrics() types.SystemMetrics {
	cpuPercent, _ := cpu.Percent(200*time.Millisecond, false)
	cpuVal := 0.0
	if len(cpuPercent) > 0 {
		cpuVal = cpuPercent[0]
	}

	vm, _ := mem.VirtualMemory()
	var ramUsed, ramTotal, ramFree int64
	if vm != nil {
		ramTotal = int64(vm.Total)
		ramUsed = int64(vm.Used)
		ramFree = int64(vm.Available)
	}

	du, _ := disk.Usage("/")
	var diskTotal, diskUsed, diskFree int64
	if du != nil {
		diskTotal = int64(du.Total)
		diskUsed = int64(du.Used)
		diskFree = int64(du.Free)
	}

	ioCounters, _ := net.IOCounters(false)
	var rxBytes, txBytes, rxSpeed, txSpeed int64
	if len(ioCounters) > 0 {
		rxBytes = int64(ioCounters[0].BytesRecv)
		txBytes = int64(ioCounters[0].BytesSent)
		now := time.Now()
		if !prevNetTime.IsZero() {
			elapsed := now.Sub(prevNetTime).Seconds()
			if elapsed > 0 {
				rxSpeed = int64(float64(rxBytes-prevNetBytes[0]) / elapsed)
				txSpeed = int64(float64(txBytes-prevNetBytes[1]) / elapsed)
			}
		}
		prevNetBytes[0] = rxBytes
		prevNetBytes[1] = txBytes
		prevNetTime = time.Now()
	}

	ld, _ := load.Avg()
	var load1, load5, load15 float64
	if ld != nil {
		load1 = ld.Load1
		load5 = ld.Load5
		load15 = ld.Load15
	}

	return types.SystemMetrics{
		CPU: cpuVal,
		RAM: types.RAMInfo{
			Total: ramTotal,
			Used:  ramUsed,
			Free:  ramFree,
		},
		Disk: types.DiskInfo{
			Total: diskTotal,
			Used:  diskUsed,
			Free:  diskFree,
		},
		Net: types.NetInfo{
			RxBytes: rxBytes,
			TxBytes: txBytes,
			RxSpeed: rxSpeed,
			TxSpeed: txSpeed,
		},
		Load: types.LoadInfo{
			Load1:  load1,
			Load5:  load5,
			Load15: load15,
		},
		Uptime: int64(time.Since(time.Unix(0, 0)).Seconds()),
	}
}