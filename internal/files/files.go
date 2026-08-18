package files

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/rockpanel/rockpanel/internal/config"
	"github.com/rockpanel/rockpanel/pkg/types"
)

func basePath() string {
	return config.C.DataDir
}

func resolve(userPath string) (string, error) {
	base := basePath()
	clean := filepath.Clean(userPath)
	if clean == "." {
		clean = ""
	}
	full := filepath.Join(base, clean)
	rel, err := filepath.Rel(base, full)
	if err != nil || strings.HasPrefix(rel, "..") {
		return "", os.ErrPermission
	}
	return full, nil
}

func List(dir string) ([]types.FileEntry, error) {
	full, err := resolve(dir)
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(full)
	if err != nil {
		return nil, err
	}
	result := make([]types.FileEntry, 0, len(entries))
	for _, e := range entries {
		info, err := e.Info()
		if err != nil {
			continue
		}
		result = append(result, types.FileEntry{
			Name:    e.Name(),
			Path:    filepath.Join(dir, e.Name()),
			Size:    info.Size(),
			ModTime: info.ModTime().Unix(),
			IsDir:   e.IsDir(),
			Mode:    info.Mode().String(),
		})
	}
	return result, nil
}

func Read(path string) ([]byte, error) {
	full, err := resolve(path)
	if err != nil {
		return nil, err
	}
	return os.ReadFile(full)
}

func Write(path string, data []byte) error {
	full, err := resolve(path)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		return err
	}
	return os.WriteFile(full, data, 0o644)
}

func Mkdir(path string) error {
	full, err := resolve(path)
	if err != nil {
		return err
	}
	return os.MkdirAll(full, 0o755)
}

func Remove(path string) error {
	full, err := resolve(path)
	if err != nil {
		return err
	}
	return os.RemoveAll(full)
}

func Rename(oldPath, newPath string) error {
	oldFull, err := resolve(oldPath)
	if err != nil {
		return err
	}
	newFull, err := resolve(newPath)
	if err != nil {
		return err
	}
	return os.Rename(oldFull, newFull)
}

func Copy(src, dst string) error {
	srcFull, err := resolve(src)
	if err != nil {
		return err
	}
	dstFull, err := resolve(dst)
	if err != nil {
		return err
	}
	info, err := os.Stat(srcFull)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return copyDir(srcFull, dstFull)
	}
	return copyFile(srcFull, dstFull)
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}

func copyDir(src, dst string) error {
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return err
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, e := range entries {
		srcPath := filepath.Join(src, e.Name())
		dstPath := filepath.Join(dst, e.Name())
		if e.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}
	return nil
}

func Extract(archivePath, destDir string) error {
	archFull, err := resolve(archivePath)
	if err != nil {
		return err
	}
	destFull, err := resolve(destDir)
	if err != nil {
		return err
	}
	if strings.HasSuffix(archFull, ".zip") {
		return extractZip(archFull, destFull)
	}
	if strings.HasSuffix(archFull, ".tar.gz") || strings.HasSuffix(archFull, ".tgz") {
		return extractTarGz(archFull, destFull)
	}
	return os.ErrInvalid
}

func extractZip(src, dst string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()
	for _, f := range r.File {
		fpath := filepath.Join(dst, f.Name)
		if !strings.HasPrefix(filepath.Clean(fpath), filepath.Clean(dst)+string(os.PathSeparator)) {
			continue
		}
		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, f.Mode())
			continue
		}
		if err := os.MkdirAll(filepath.Dir(fpath), 0o755); err != nil {
			return err
		}
		rc, err := f.Open()
		if err != nil {
			return err
		}
		out, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			rc.Close()
			return err
		}
		io.Copy(out, rc)
		out.Close()
		rc.Close()
	}
	return nil
}

func extractTarGz(src, dst string) error {
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		fpath := filepath.Join(dst, hdr.Name)
		if !strings.HasPrefix(filepath.Clean(fpath), filepath.Clean(dst)+string(os.PathSeparator)) {
			continue
		}
		switch hdr.Typeflag {
		case tar.TypeDir:
			os.MkdirAll(fpath, os.FileMode(hdr.Mode))
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(fpath), 0o755); err != nil {
				return err
			}
			out, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.FileMode(hdr.Mode))
			if err != nil {
				return err
			}
			io.Copy(out, tr)
			out.Close()
		}
	}
	return nil
}

func Compress(srcPaths []string, destPath string) error {
	destFull, err := resolve(destPath)
	if err != nil {
		return err
	}
	if strings.HasSuffix(destFull, ".zip") {
		return createZip(srcPaths, destFull)
	}
	return createTarGz(srcPaths, destFull)
}

func createZip(srcPaths []string, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	f, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer f.Close()
	zw := zip.NewWriter(f)
	defer zw.Close()
	for _, src := range srcPaths {
		srcFull, err := resolve(src)
		if err != nil {
			return err
		}
		base := filepath.Base(srcFull)
		filepath.Walk(srcFull, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			rel, _ := filepath.Rel(srcFull, path)
			if rel == "." {
				return nil
			}
			hdr, err := zip.FileInfoHeader(info)
			if err != nil {
				return err
			}
			hdr.Name = filepath.Join(base, rel)
			hdr.Method = zip.Deflate
			w, err := zw.CreateHeader(hdr)
			if err != nil {
				return err
			}
			if !info.IsDir() {
				data, err := os.ReadFile(path)
				if err != nil {
					return err
				}
				w.Write(data)
			}
			return nil
		})
	}
	return nil
}

func createTarGz(srcPaths []string, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	f, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer f.Close()
	gz := gzip.NewWriter(f)
	defer gz.Close()
	tw := tar.NewWriter(gz)
	defer tw.Close()
	for _, src := range srcPaths {
		srcFull, err := resolve(src)
		if err != nil {
			return err
		}
		base := filepath.Base(srcFull)
		filepath.Walk(srcFull, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			rel, _ := filepath.Rel(srcFull, path)
			if rel == "." {
				return nil
			}
			hdr, err := tar.FileInfoHeader(info, "")
			if err != nil {
				return err
			}
			hdr.Name = filepath.Join(base, rel)
			if err := tw.WriteHeader(hdr); err != nil {
				return err
			}
			if !info.IsDir() {
				data, err := os.ReadFile(path)
				if err != nil {
					return err
				}
				tw.Write(data)
			}
			return nil
		})
	}
	return nil
}

func Stat(path string) (*types.FileEntry, error) {
	full, err := resolve(path)
	if err != nil {
		return nil, err
	}
	info, err := os.Stat(full)
	if err != nil {
		return nil, err
	}
	return &types.FileEntry{
		Name:    info.Name(),
		Path:    path,
		Size:    info.Size(),
		ModTime: info.ModTime().Unix(),
		IsDir:   info.IsDir(),
		Mode:    info.Mode().String(),
	}, nil
}