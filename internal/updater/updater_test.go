package updater

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCompareVersions(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"0.10.0", "0.11.0", -1},
		{"0.11.0", "0.10.0", 1},
		{"0.10.0", "0.10.0", 0},
		{"1.0.0", "0.99.99", 1},
		{"0.9", "0.9.0", 0},
		{"1", "1.0.0", 0},
		{"1.10", "1.9.5", 1},
		{"0.10.0", "1.0.0", -1},
	}
	for _, c := range cases {
		got, err := CompareVersions(c.a, c.b)
		if err != nil {
			t.Errorf("CompareVersions(%q,%q) error: %v", c.a, c.b, err)
			continue
		}
		if got != c.want {
			t.Errorf("CompareVersions(%q,%q) = %d, ingin %d", c.a, c.b, got, c.want)
		}
	}
}

func TestCompareVersionsInvalid(t *testing.T) {
	for _, v := range []string{"", "a.b.c", "1.-1", "1.2.3.4", ".1"} {
		if _, err := CompareVersions(v, "0.1.0"); err == nil {
			t.Errorf("versi %q seharusnya ditolak", v)
		}
		if _, err := CompareVersions("0.1.0", v); err == nil {
			t.Errorf("versi %q seharusnya ditolak", v)
		}
	}
}

func TestResolveDriveLink(t *testing.T) {
	cases := []struct {
		in   string
		want string
		err  bool
	}{
		{
			in:   "https://drive.google.com/file/d/ABC123/view?usp=sharing",
			want: "https://drive.usercontent.google.com/download?id=ABC123&export=download&confirm=t",
		},
		{
			in:   "https://drive.google.com/uc?export=download&id=XYZ789",
			want: "https://drive.usercontent.google.com/download?id=XYZ789&export=download&confirm=t",
		},
		{
			in:   "https://drive.usercontent.google.com/download?id=ID42&export=download",
			want: "https://drive.usercontent.google.com/download?id=ID42&export=download&confirm=t",
		},
		{
			in:   "https://myserver.example.com/SPHManager.exe",
			want: "https://myserver.example.com/SPHManager.exe",
		},
		{in: "", err: true},
		{in: "https://drive.google.com/file/d", err: true},
	}
	for _, c := range cases {
		got, err := ResolveDriveLink(c.in)
		if c.err {
			if err == nil {
				t.Errorf("ResolveDriveLink(%q) seharusnya error", c.in)
			}
			continue
		}
		if err != nil {
			t.Errorf("ResolveDriveLink(%q) error tak terduga: %v", c.in, err)
			continue
		}
		if got != c.want {
			t.Errorf("ResolveDriveLink(%q) = %q, ingin %q", c.in, got, c.want)
		}
	}
}

func TestAppendAndReadMeta(t *testing.T) {
	path := filepath.Join(t.TempDir(), "candidate.exe")
	if err := os.WriteFile(path, []byte("MZ\x90\x00 fake binary body"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := AppendMeta(path, ExeMeta{Version: "0.13.0", Notes: "Fix export"}); err != nil {
		t.Fatalf("AppendMeta error: %v", err)
	}
	meta, err := ReadExeMeta(path)
	if err != nil {
		t.Fatalf("ReadExeMeta error: %v", err)
	}
	if meta.Version != "0.13.0" || meta.Notes != "Fix export" {
		t.Errorf("meta salah: %+v", meta)
	}
}

func TestAppendAndReadMetaEmptyNotes(t *testing.T) {
	path := filepath.Join(t.TempDir(), "candidate.exe")
	if err := os.WriteFile(path, []byte("MZ"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := AppendMeta(path, ExeMeta{Version: "0.14.0"}); err != nil {
		t.Fatal(err)
	}
	meta, err := ReadExeMeta(path)
	if err != nil {
		t.Fatalf("ReadExeMeta error: %v", err)
	}
	if meta.Version != "0.14.0" {
		t.Errorf("versi salah: %q", meta.Version)
	}
}

func TestReadExeMetaMissing(t *testing.T) {
	path := filepath.Join(t.TempDir(), "plain.exe")
	if err := os.WriteFile(path, []byte("plain binary no meta"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadExeMeta(path); err == nil {
		t.Error("file tanpa trailer seharusnya error")
	}
}

func TestReadExeMetaCorrupt(t *testing.T) {
	path := filepath.Join(t.TempDir(), "corrupt.exe")
	data := []byte("abc\n" + metaBegin + "\n{not json\n" + metaEnd + "\n")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadExeMeta(path); err == nil {
		t.Error("trailer rusak seharusnya error")
	}
}

func TestReadExeMetaEmptyVersion(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nover.exe")
	data := []byte("abc\n" + metaBegin + "\n{\"version\":\"\"}\n" + metaEnd + "\n")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadExeMeta(path); err == nil {
		t.Error("trailer tanpa versi seharusnya error")
	}
}

func TestFetchCandidateHTTP(t *testing.T) {
	var body bytes.Buffer
	body.WriteString("MZ\x90\x00 fake candidate")
	if err := appendMetaBlock(&body, ExeMeta{Version: "0.15.0", Notes: "Rilis baru"}); err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(body.Bytes())
	}))
	defer srv.Close()

	dest := filepath.Join(t.TempDir(), "downloaded.exe")
	meta, err := FetchCandidate(srv.URL, dest, 5*time.Second, nil)
	if err != nil {
		t.Fatalf("FetchCandidate error: %v", err)
	}
	if meta.Version != "0.15.0" {
		t.Errorf("versi salah: %q", meta.Version)
	}
	got, err := os.ReadFile(dest)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, body.Bytes()) {
		t.Error("isi file hasil unduhan tidak cocok")
	}
}

func TestFetchCandidateHTTPNoMeta(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("not an update candidate"))
	}))
	defer srv.Close()

	dest := filepath.Join(t.TempDir(), "downloaded.exe")
	if _, err := FetchCandidate(srv.URL, dest, 5*time.Second, nil); err == nil {
		t.Error("kandidat tanpa trailer seharusnya error")
	}
	if _, err := os.Stat(dest); !os.IsNotExist(err) {
		t.Error("file kandidat yang gagal seharusnya dihapus")
	}
}