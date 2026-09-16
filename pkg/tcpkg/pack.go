package tcpkg

import (
	"archive/zip"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Pack writes only validated definitions and referenced assets. It never overwrites a file.
func Pack(root, out string) error {
	p, err := Load(root, true)
	if err != nil {
		return err
	}
	if out == "" {
		out = fmt.Sprintf("%s-%s.tcpkg", p.Provider.ProviderID, p.Provider.Version)
	}
	f, err := os.CreateTemp(filepath.Dir(out), ".tcpkg-*")
	if err != nil {
		return err
	}
	tmp := f.Name()
	defer os.Remove(tmp)
	z := zip.NewWriter(f)
	names := []string{}
	for n := range p.Files {
		names = append(names, n)
	}
	sort.Strings(names)
	var total int64
	for _, n := range names {
		b := p.Files[n]
		total += int64(len(b))
		if total > MaxPackageSize {
			z.Close()
			f.Close()
			return fmt.Errorf("expanded package exceeds size limit")
		}
		w, e := z.Create(n)
		if e == nil {
			_, e = w.Write(b)
		}
		if e != nil {
			z.Close()
			f.Close()
			return e
		}
	}
	if err = z.Close(); err != nil {
		f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	if _, err = Verify(tmp); err != nil {
		return err
	}
	// A hard link creates the destination atomically and fails if it already exists.
	if err = os.Link(tmp, out); err != nil {
		return fmt.Errorf("create %s (destination must not exist): %w", out, err)
	}
	return nil
}
func Verify(file string) (*Project, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	s, err := f.Stat()
	if err != nil {
		return nil, err
	}
	parsed, err := ParsePackage(f, s.Size())
	if err != nil {
		return nil, err
	}
	// Validate exact paths too: unknown entries must not silently ship secrets.
	z, err := zip.NewReader(f, s.Size())
	if err != nil {
		return nil, err
	}
	files := map[string][]byte{"_provider.yaml": parsed.ProviderYAML}
	for _, e := range z.File {
		if e.FileInfo().IsDir() {
			continue
		}
		n := e.Name
		if n == "_provider.yaml" || n == "_provider.yml" {
			continue
		}
		parts := strings.Split(n, "/")
		if len(parts) != 2 || !plain(parts[1]) {
			return nil, fmt.Errorf("unsupported package path %s", n)
		}
		switch parts[0] {
		case "tools":
			if !isYAMLFile(n) {
				return nil, fmt.Errorf("unexpected tool file %s", n)
			}
		case "wasm":
			if !strings.HasSuffix(n, ".wasm") {
				return nil, fmt.Errorf("unexpected WASM file %s", n)
			}
		case "_assets":
		default:
			return nil, fmt.Errorf("unexpected package file %s", n)
		}
		b, err := readZipEntry(e)
		if err != nil {
			return nil, err
		}
		files[n] = b
	}
	return ValidateFiles(files, true)
}
