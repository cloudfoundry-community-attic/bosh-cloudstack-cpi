package bundlecollection

import (
	bosherr "bosh/errors"
	boshsys "bosh/system"
	"os"
	"path/filepath"
)

type FileBundleCollection struct {
	name string
	path string
	fs   boshsys.FileSystem
}

func NewFileBundleCollection(name, path string, fs boshsys.FileSystem) *FileBundleCollection {
	return &FileBundleCollection{
		name: name,
		path: path,
		fs:   fs,
	}
}

// Installed into {{ path }}/data/{{ name }}/{{ bundle.BundleName }}/{{ bundle.BundleVersion }}
func (s *FileBundleCollection) Install(bundle Bundle) (boshsys.FileSystem, string, error) {
	err := s.checkBundle(bundle)
	if err != nil {
		return nil, "", err
	}

	path := s.buildInstallPath(bundle)
	err = s.fs.MkdirAll(path, os.FileMode(0755))
	if err != nil {
		return nil, "", bosherr.WrapError(err, "failed to create install dir")
	}

	return s.fs, path, nil
}

// Symlinked from {{ path }}/{{ name }}/{{ bundle.BundleName }} to installed path
func (s *FileBundleCollection) Enable(bundle Bundle) error {
	err := s.checkBundle(bundle)
	if err != nil {
		return err
	}

	installPath := s.buildInstallPath(bundle)
	if !s.fs.FileExists(installPath) {
		return bosherr.New("bundle must be installed")
	}

	enablePath := s.buildEnablePath(bundle)
	err = s.fs.MkdirAll(filepath.Dir(enablePath), os.FileMode(0755))
	if err != nil {
		return bosherr.WrapError(err, "failed to create enable dir")
	}

	err = s.fs.Symlink(installPath, enablePath)
	if err != nil {
		return bosherr.WrapError(err, "failed to enable")
	}

	return nil
}

func (s *FileBundleCollection) checkBundle(bundle Bundle) error {
	if len(bundle.BundleName()) == 0 {
		return bosherr.New("missing bundle name")
	}
	if len(bundle.BundleVersion()) == 0 {
		return bosherr.New("missing bundle version")
	}
	return nil
}

func (s *FileBundleCollection) buildInstallPath(bundle Bundle) string {
	return filepath.Join(s.path, "data", s.name, bundle.BundleName(), bundle.BundleVersion())
}

func (s *FileBundleCollection) buildEnablePath(bundle Bundle) string {
	return filepath.Join(s.path, s.name, bundle.BundleName())
}
