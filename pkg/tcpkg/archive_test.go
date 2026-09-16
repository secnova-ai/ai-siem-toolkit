package tcpkg

import (
	"archive/zip"
	"bytes"
	"testing"
)

func TestRejectInvalidArchives(t *testing.T) {
	for _, names := range [][]string{{"../secret"}, {"/absolute"}, {"tools\\a.yaml"}, {"_provider.yaml", "_provider.yaml"}, {"_provider.yaml", "_provider.yml"}} {
		var b bytes.Buffer
		z := zip.NewWriter(&b)
		for _, n := range names {
			w, _ := z.Create(n)
			w.Write([]byte("provider_id: demo"))
		}
		z.Close()
		if _, err := ParsePackage(bytes.NewReader(b.Bytes()), int64(b.Len())); err == nil {
			t.Errorf("accepted %v", names)
		}
	}
}
func TestRejectCorruptCRC(t *testing.T) {
	var b bytes.Buffer
	z := zip.NewWriter(&b)
	w, _ := z.CreateHeader(&zip.FileHeader{Name: "_provider.yaml", Method: zip.Store})
	w.Write([]byte("provider_id: demo"))
	z.Close()
	raw := b.Bytes()
	i := bytes.Index(raw, []byte("provider_id: demo"))
	raw[i] = 'X'
	if _, err := ParsePackage(bytes.NewReader(raw), int64(len(raw))); err == nil {
		t.Fatal("CRC corruption accepted")
	}
}
