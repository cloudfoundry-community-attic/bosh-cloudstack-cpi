package action

import (
	fakeappl "bosh/agent/applier/fakes"
	faketask "bosh/agent/task/fakes"
	fakeblobstore "bosh/blobstore/fakes"
	fakeplatform "bosh/platform/fakes"
	fakesettings "bosh/settings/fakes"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewFactory(t *testing.T) {
	actions := []string{
		"apply",
		"drain",
		"fetch_logs",
		"get_task",
		"get_state",
		"list_disk",
		"migrate_disk",
		"mount_disk",
		"ping",
		"ssh",
		"start",
		"stop",
		"unmount_disk",
	}

	_, _, _, _, _, factory := buildFactory()

	for _, actionName := range actions {
		action := factory.Create(actionName)
		assert.NotNil(t, action)
	}
}

func TestNewFactoryApply(t *testing.T) {
	_, platform, _, _, applier, factory := buildFactory()
	action := factory.Create("apply")
	assert.NotNil(t, action)
	assert.Equal(t, newApply(applier, platform.Fs, platform), action)
}

func TestNewFactoryFetchLogs(t *testing.T) {
	_, platform, blobstore, _, _, factory := buildFactory()
	action := factory.Create("fetch_logs")
	assert.NotNil(t, action)
	assert.Equal(t, newLogs(platform.GetCompressor(), blobstore), action)
}

func TestNewFactoryGetTask(t *testing.T) {
	_, _, _, taskService, _, factory := buildFactory()
	action := factory.Create("get_task")
	assert.NotNil(t, action)
	assert.Equal(t, newGetTask(taskService), action)
}

func TestNewFactoryGetState(t *testing.T) {
	settings, platform, _, _, _, factory := buildFactory()
	action := factory.Create("get_state")
	assert.NotNil(t, action)
	assert.Equal(t, newGetState(settings, platform.GetFs()), action)
}

func TestNewFactoryListDisk(t *testing.T) {
	settings, platform, _, _, _, factory := buildFactory()
	action := factory.Create("list_disk")
	assert.NotNil(t, action)
	assert.Equal(t, newListDisk(settings, platform), action)
}

func TestNewFactoryMigrateDisk(t *testing.T) {
	settings, platform, _, _, _, factory := buildFactory()
	action := factory.Create("migrate_disk")
	assert.NotNil(t, action)
	assert.Equal(t, newMigrateDisk(settings, platform), action)
}

func TestNewFactoryMountDisk(t *testing.T) {
	settings, platform, _, _, _, factory := buildFactory()
	action := factory.Create("mount_disk")
	assert.NotNil(t, action)
	assert.Equal(t, newMountDisk(settings, platform), action)
}

func TestNewFactorySsh(t *testing.T) {
	settings, platform, _, _, _, factory := buildFactory()
	action := factory.Create("ssh")
	assert.NotNil(t, action)
	assert.Equal(t, newSsh(settings, platform), action)
}

func TestNewFactoryUnmountDisk(t *testing.T) {
	settings, platform, _, _, _, factory := buildFactory()
	action := factory.Create("unmount_disk")
	assert.NotNil(t, action)
	assert.Equal(t, newUnmountDisk(settings, platform), action)
}

func buildFactory() (
	settings *fakesettings.FakeSettingsService,
	platform *fakeplatform.FakePlatform,
	blobstore *fakeblobstore.FakeBlobstore,
	taskService *faketask.FakeService,
	applier *fakeappl.FakeApplier,
	factory Factory) {

	settings = &fakesettings.FakeSettingsService{}
	platform = fakeplatform.NewFakePlatform()
	blobstore = &fakeblobstore.FakeBlobstore{}
	taskService = &faketask.FakeService{}
	applier = fakeappl.NewFakeApplier()

	factory = NewFactory(settings, platform, blobstore, taskService, applier)
	return
}
