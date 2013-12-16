package action

import (
	boshassert "bosh/assert"
	fakeplatform "bosh/platform/fakes"
	fakesettings "bosh/settings/fakes"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMountDisk(t *testing.T) {
	settings := &fakesettings.FakeSettingsService{}
	settings.Disks.Persistent = map[string]string{"vol-123": "/dev/sdf"}
	platform, mountDisk := buildMountDiskAction(settings)

	payload := `{"arguments":["vol-123"]}`
	result, err := mountDisk.Run([]byte(payload))
	assert.NoError(t, err)
	boshassert.MatchesJsonString(t, result, "{}")

	assert.True(t, settings.SettingsWereRefreshed)

	assert.Equal(t, platform.MountPersistentDiskDevicePath, "/dev/sdf")
	assert.Equal(t, platform.MountPersistentDiskMountPoint, "/var/vcap/store")
}

func TestMountDiskWhenStoreAlreadyMounted(t *testing.T) {
	settings := &fakesettings.FakeSettingsService{}
	settings.Disks.Persistent = map[string]string{"vol-123": "/dev/sdf"}
	platform, mountDisk := buildMountDiskAction(settings)

	platform.IsMountPointResult = true

	payload := `{"arguments":["vol-123"]}`
	result, err := mountDisk.Run([]byte(payload))
	assert.NoError(t, err)
	boshassert.MatchesJsonString(t, result, "{}")

	assert.Equal(t, platform.IsMountPointPath, "/var/vcap/store")

	assert.Equal(t, platform.MountPersistentDiskDevicePath, "/dev/sdf")
	assert.Equal(t, platform.MountPersistentDiskMountPoint, "/var/vcap/store_migration_target")
}

func TestMountDiskWithMissingVolumeId(t *testing.T) {
	settings := &fakesettings.FakeSettingsService{}
	_, mountDisk := buildMountDiskAction(settings)

	payload := `{"arguments":[]}`
	_, err := mountDisk.Run([]byte(payload))
	assert.Error(t, err)
}

func TestMountDiskWhenDevicePathNotFound(t *testing.T) {
	settings := &fakesettings.FakeSettingsService{}
	settings.Disks.Persistent = map[string]string{"vol-123": "/dev/sdf"}
	_, mountDisk := buildMountDiskAction(settings)

	payload := `{"arguments":["vol-456"]}`
	_, err := mountDisk.Run([]byte(payload))
	assert.Error(t, err)
}

func buildMountDiskAction(settings *fakesettings.FakeSettingsService) (*fakeplatform.FakePlatform, mountDiskAction) {
	platform := fakeplatform.NewFakePlatform()
	action := newMountDisk(settings, platform)
	return platform, action
}
