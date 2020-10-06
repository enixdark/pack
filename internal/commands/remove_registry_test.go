package commands_test

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/heroku/color"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"

	"github.com/buildpacks/pack/internal/commands"
	"github.com/buildpacks/pack/internal/config"
	ilogging "github.com/buildpacks/pack/internal/logging"
	h "github.com/buildpacks/pack/testhelpers"
)

func TestRemoveRegistry(t *testing.T) {
	color.Disable(true)
	defer color.Disable(false)

	spec.Run(t, "Commands", testRemoveRegistryCommand, spec.Parallel(), spec.Report(report.Terminal{}))
}

func testRemoveRegistryCommand(t *testing.T, when spec.G, it spec.S) {
	when("#RemoveRegistry", func() {
		var (
			outBuf     bytes.Buffer
			logger     = ilogging.NewLogWithWriters(&outBuf, &outBuf)
			tmpDir     string
			configFile string
			cfg        config.Config
			assert     = h.NewAssertionManager(t)
		)

		it.Before(func() {
			var err error
			tmpDir, err = ioutil.TempDir("", "pack-home-*")
			assert.Nil(err)

			cfg = config.Config{
				DefaultRegistryName: "buildpack-registry",
				Registries: []config.Registry{
					{
						Name: "buildpack-registry",
						URL:  "https://github.com/buildpacks/registry-index",
						Type: "github",
					},
					{
						Name: "elbandito-registry",
						URL:  "https://github.com/elbandito/registry-index",
						Type: "github",
					},
				},
			}

			configFile = filepath.Join(tmpDir, "config.toml")
			err = config.Write(cfg, configFile)
			assert.Nil(err)
		})

		it.After(func() {
			_ = os.RemoveAll(tmpDir)
		})

		it("should remove the registry", func() {
			command := commands.RemoveRegistry(logger, cfg, configFile)
			command.SetArgs([]string{"elbandito-registry"})
			assert.Succeeds(command.Execute())

			newCfg, err := config.Read(configFile)
			assert.Nil(err)

			assert.Equal(newCfg, config.Config{
				DefaultRegistryName: "buildpack-registry",
				Registries: []config.Registry{
					{
						Name: "buildpack-registry",
						URL:  "https://github.com/buildpacks/registry-index",
						Type: "github",
					},
				},
			})
		})

		it("should remove the registry and matching default registry name", func() {
			command := commands.RemoveRegistry(logger, cfg, configFile)
			command.SetArgs([]string{"buildpack-registry"})
			assert.Succeeds(command.Execute())

			newCfg, err := config.Read(configFile)
			assert.Nil(err)

			assert.Equal(newCfg, config.Config{
				DefaultRegistryName: config.OfficialRegistryName,
				Registries: []config.Registry{
					{
						Name: "elbandito-registry",
						URL:  "https://github.com/elbandito/registry-index",
						Type: "github",
					},
				},
			})
		})

		it("should return error when registry does NOT already exist", func() {
			command := commands.RemoveRegistry(logger, cfg, configFile)
			command.SetArgs([]string{"missing-registry"})
			assert.Error(command.Execute())

			output := outBuf.String()
			h.AssertContains(t, output, "registry 'missing-registry' does not exist")
		})

		it("should throw error when registry name is official", func() {
			command := commands.RemoveRegistry(logger, config.Config{}, configFile)
			command.SetArgs([]string{"official"})
			assert.Error(command.Execute())

			output := outBuf.String()
			h.AssertContains(t, output, "'official' is a reserved registry name, please provide a different registry")
		})
	})
}
