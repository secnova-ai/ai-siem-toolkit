package tcpkg

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"path"
	"sort"
	"strings"
)

// PackageContent holds the parsed contents of a .tcpkg archive.
type PackageContent struct {
	ProviderYAML []byte
	ToolYAMLs    [][]byte          // sorted by path for deterministic checksum
	WASMFiles    map[string][]byte // basename → bytes
	AssetFiles   map[string][]byte // _assets/ basename → bytes (icons etc.)
}

const (
	MaxPackageSize  int64 = 256 << 20 // 256 MiB per .tcpkg archive
	maxPkgEntrySize       = 32 << 20  // 32 MiB per entry
)

// ParsePackage reads a .tcpkg (zip) archive and extracts the provider YAML,
// tool YAMLs, and any WASM binaries. Expected layout:
//
//	_provider.yaml
//	tools/*.yaml
//	wasm/*.wasm   (optional)
//
// Returns an error if the archive is malformed or missing _provider.yaml.
func ParsePackage(r io.ReaderAt, size int64) (*PackageContent, error) {
	if size > MaxPackageSize {
		return nil, fmt.Errorf("package exceeds size limit (%d bytes)", MaxPackageSize)
	}
	zr, err := zip.NewReader(r, size)
	if err != nil {
		return nil, fmt.Errorf("open zip: %w", err)
	}

	out := &PackageContent{
		WASMFiles:  make(map[string][]byte),
		AssetFiles: make(map[string][]byte),
	}

	type namedData struct {
		name string
		data []byte
	}
	var toolFiles []namedData
	var totalSize uint64
	seenPaths := make(map[string]bool)

	for _, f := range zr.File {
		if f.FileInfo().IsDir() {
			continue
		}
		// Guard against path traversal.
		cleaned := path.Clean(f.Name)
		if path.IsAbs(f.Name) || cleaned != f.Name || strings.HasPrefix(cleaned, "..") || strings.ContainsRune(f.Name, '\\') {
			return nil, fmt.Errorf("invalid path in package: %q", f.Name)
		}
		if f.UncompressedSize64 > maxPkgEntrySize {
			return nil, fmt.Errorf("entry %q exceeds size limit (%d bytes)", f.Name, maxPkgEntrySize)
		}

		totalSize += f.UncompressedSize64
		if totalSize > uint64(MaxPackageSize) {
			return nil, fmt.Errorf("expanded package exceeds size limit")
		}
		if seenPaths[f.Name] {
			return nil, fmt.Errorf("duplicate package entry %q", f.Name)
		}
		seenPaths[f.Name] = true
		data, err := readZipEntry(f)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", f.Name, err)
		}

		switch {
		case f.Name == "_provider.yaml" || f.Name == "_provider.yml":
			if out.ProviderYAML != nil {
				return nil, fmt.Errorf("duplicate provider declaration")
			}
			out.ProviderYAML = data
		case strings.HasPrefix(f.Name, "tools/") && isYAMLFile(f.Name):
			toolFiles = append(toolFiles, namedData{f.Name, data})
		case strings.HasPrefix(f.Name, "wasm/") && strings.HasSuffix(f.Name, ".wasm"):
			if _, exists := out.WASMFiles[path.Base(f.Name)]; exists {
				return nil, fmt.Errorf("duplicate WASM filename")
			}
			out.WASMFiles[path.Base(f.Name)] = data
		case strings.HasPrefix(f.Name, "_assets/"):
			if _, exists := out.AssetFiles[path.Base(f.Name)]; exists {
				return nil, fmt.Errorf("duplicate asset filename")
			}
			out.AssetFiles[path.Base(f.Name)] = data
		}
	}

	if len(out.ProviderYAML) == 0 {
		return nil, fmt.Errorf("package missing _provider.yaml")
	}

	// Sort tool files by path for deterministic checksum computation.
	sort.Slice(toolFiles, func(i, j int) bool {
		return toolFiles[i].name < toolFiles[j].name
	})
	for _, tf := range toolFiles {
		out.ToolYAMLs = append(out.ToolYAMLs, tf.data)
	}

	return out, nil
}

func readZipEntry(f *zip.File) ([]byte, error) {
	rc, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, io.LimitReader(rc, maxPkgEntrySize+1)); err != nil {
		return nil, err
	}
	if buf.Len() > maxPkgEntrySize {
		return nil, fmt.Errorf("entry exceeds size limit")
	}
	return buf.Bytes(), nil
}

func isYAMLFile(s string) bool { return strings.HasSuffix(s, ".yaml") || strings.HasSuffix(s, ".yml") }
