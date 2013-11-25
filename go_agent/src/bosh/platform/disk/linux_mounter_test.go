package disk

import (
	fakesys "bosh/system/fakes"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLinuxMount(t *testing.T) {
	runner, fs := getLinuxMounterDependencies()
	fs.WriteToFile("/proc/mounts", "")

	mounter := newLinuxMounter(runner, fs)
	err := mounter.Mount("/dev/foo", "/mnt/foo")

	assert.NoError(t, err)
	assert.Equal(t, 1, len(runner.RunCommands))
	assert.Equal(t, []string{"mount", "/dev/foo", "/mnt/foo"}, runner.RunCommands[0])
}

func TestLinuxMountWhenDiskIsAlreadyMountedToTheGoodMountPoint(t *testing.T) {
	runner, fs := getLinuxMounterDependencies()
	fs.WriteToFile("/proc/mounts", "/dev/foo /mnt/foo\n/dev/bar /mnt/bar")

	mounter := newLinuxMounter(runner, fs)
	err := mounter.Mount("/dev/foo", "/mnt/foo")

	assert.NoError(t, err)
	assert.Equal(t, 0, len(runner.RunCommands))
}

func TestLinuxMountWhenDiskIsAlreadyMountedToTheWrongMountPoint(t *testing.T) {
	runner, fs := getLinuxMounterDependencies()
	fs.WriteToFile("/proc/mounts", "/dev/foo /mnt/foobarbaz\n/dev/bar /mnt/bar")

	mounter := newLinuxMounter(runner, fs)
	err := mounter.Mount("/dev/foo", "/mnt/foo")

	assert.Error(t, err)
	assert.Equal(t, 0, len(runner.RunCommands))
}

func TestLinuxMountWhenAnotherDiskIsAlreadyMountedToMountPoint(t *testing.T) {
	runner, fs := getLinuxMounterDependencies()
	fs.WriteToFile("/proc/mounts", "/dev/baz /mnt/foo\n/dev/bar /mnt/bar")

	mounter := newLinuxMounter(runner, fs)
	err := mounter.Mount("/dev/foo", "/mnt/foo")

	assert.Error(t, err)
	assert.Equal(t, 0, len(runner.RunCommands))
}

func TestLinuxSwapOn(t *testing.T) {
	runner, fs := getLinuxMounterDependencies()
	runner.CommandResults = map[string][]string{
		"swapon -s": []string{"Filename				Type		Size	Used	Priority\n", ""},
	}

	mounter := newLinuxMounter(runner, fs)
	mounter.SwapOn("/dev/swap")

	assert.Equal(t, 2, len(runner.RunCommands))
	assert.Equal(t, []string{"swapon", "/dev/swap"}, runner.RunCommands[1])
}

func TestLinuxSwapOnWhenAlreadyOn(t *testing.T) {
	runner, fs := getLinuxMounterDependencies()
	runner.CommandResults = map[string][]string{
		"swapon -s": []string{SWAPON_USAGE_OUTPUT, ""},
	}

	mounter := newLinuxMounter(runner, fs)
	mounter.SwapOn("/dev/swap")
	assert.Equal(t, 1, len(runner.RunCommands))
	assert.Equal(t, []string{"swapon", "-s"}, runner.RunCommands[0])
}

const SWAPON_USAGE_OUTPUT = `Filename				Type		Size	Used	Priority
/dev/swap                              partition	78180316	0	-1
`

func TestLinuxSwapOnWhenAlreadyOnOtherDevice(t *testing.T) {
	runner, fs := getLinuxMounterDependencies()
	runner.CommandResults = map[string][]string{
		"swapon -s": []string{SWAPON_USAGE_OUTPUT_WITH_OTHER_DEVICE, ""},
	}

	mounter := newLinuxMounter(runner, fs)
	mounter.SwapOn("/dev/swap")
	assert.Equal(t, 2, len(runner.RunCommands))
	assert.Equal(t, []string{"swapon", "-s"}, runner.RunCommands[0])
	assert.Equal(t, []string{"swapon", "/dev/swap"}, runner.RunCommands[1])
}

const SWAPON_USAGE_OUTPUT_WITH_OTHER_DEVICE = `Filename				Type		Size	Used	Priority
/dev/swap2                              partition	78180316	0	-1
`

func getLinuxMounterDependencies() (runner *fakesys.FakeCmdRunner, fs *fakesys.FakeFileSystem) {
	runner = &fakesys.FakeCmdRunner{}
	fs = &fakesys.FakeFileSystem{}
	return
}
