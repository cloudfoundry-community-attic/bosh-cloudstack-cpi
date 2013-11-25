package fakes

type FakeMounter struct {
	MountPartitionPaths  []string
	MountMountPoints     []string
	SwapOnPartitionPaths []string
}

func (m *FakeMounter) Mount(partitionPath, mountPoint string) (err error) {
	m.MountPartitionPaths = append(m.MountPartitionPaths, partitionPath)
	m.MountMountPoints = append(m.MountMountPoints, mountPoint)
	return
}

func (m *FakeMounter) SwapOn(partitionPath string) (err error) {
	m.SwapOnPartitionPaths = append(m.SwapOnPartitionPaths, partitionPath)
	return
}
