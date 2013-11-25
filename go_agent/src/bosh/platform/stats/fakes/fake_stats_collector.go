package fakes

import boshstats "bosh/platform/stats"

type FakeStatsCollector struct {
	CpuLoad   boshstats.CpuLoad
	CpuStats  boshstats.CpuStats
	MemStats  boshstats.MemStats
	SwapStats boshstats.MemStats
	DiskStats map[string]boshstats.DiskStats
}

func (c *FakeStatsCollector) GetCpuLoad() (load boshstats.CpuLoad, err error) {
	load = c.CpuLoad
	return
}

func (c *FakeStatsCollector) GetCpuStats() (stats boshstats.CpuStats, err error) {
	stats = c.CpuStats
	return
}

func (c *FakeStatsCollector) GetMemStats() (stats boshstats.MemStats, err error) {
	stats = c.MemStats
	return
}

func (c *FakeStatsCollector) GetSwapStats() (stats boshstats.MemStats, err error) {
	stats = c.SwapStats
	return
}

func (c *FakeStatsCollector) GetDiskStats(devicePath string) (stats boshstats.DiskStats, err error) {
	stats = c.DiskStats[devicePath]
	return
}
