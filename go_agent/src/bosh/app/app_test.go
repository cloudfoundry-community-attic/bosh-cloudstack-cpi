package app_test

import (
	"io/ioutil"
	"os"
	"path/filepath"

	boshapp "bosh/app"
	boshlog "bosh/logger"

	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
)

func init() {
	Describe("App", func() {
		var baseDir string

		BeforeEach(func() {
			baseDir, _ = ioutil.TempDir("", "go-agent-test")
			os.Mkdir(filepath.Join(baseDir, "bosh"), os.ModePerm)
			settingsPath := filepath.Join(baseDir, "bosh", "settings.json")

			settingsJson := `{
					"agent_id": "my-agent-id",
					"blobstore": {
						"options": {
							"bucket_name": "george",
							"encryption_key": "optional encryption key",
							"access_key_id": "optional access key id",
							"secret_access_key": "optional secret access key"
						},
						"provider": "dummy"
					},
					"disks": {
						"ephemeral": "/dev/sdb",
						"persistent": {
							"vol-xxxxxx": "/dev/sdf"
						},
						"system": "/dev/sda1"
					},
					"env": {
						"bosh": {
							"password": "some encrypted password"
						}
					},
					"networks": {
						"netA": {
							"default": ["dns", "gateway"],
							"ip": "ww.ww.ww.ww",
							"dns": [
								"xx.xx.xx.xx",
								"yy.yy.yy.yy"
							]
						},
						"netB": {
							"dns": [
								"zz.zz.zz.zz"
							]
						}
					},
					"Mbus": "https://vcap:hello@0.0.0.0:6868",
					"ntp": [
						"0.north-america.pool.ntp.org",
						"1.north-america.pool.ntp.org"
					],
					"vm": {
						"name": "vm-abc-def"
					}
				}`

			ioutil.WriteFile(settingsPath, []byte(settingsJson), 0640)
		})

		AfterEach(func() {
			os.RemoveAll(baseDir)
		})

		It("Sets up device path resolver on platform specific to infrastructure", func() {
			logger := boshlog.NewLogger(boshlog.LEVEL_DEBUG)
			app := boshapp.New(logger)

			err := app.Setup([]string{
				"bosh-agent",
				"-I", "dummy",
				"-P", "dummy",
				"-b", baseDir,
			})

			assert.NoError(GinkgoT(), err)

			assert.Equal(GinkgoT(), app.GetPlatform().GetDevicePathResolver(), app.GetInfrastructure().GetDevicePathResolver())
		})
	})
}
