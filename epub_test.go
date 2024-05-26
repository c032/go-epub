package epub_test

import (
	"errors"
	"io/fs"
	"os"
	"testing"

	"github.com/c032/go-epub"
)

const fileYoujoSenkiV01 = "testdata/youjo-senki-v01.epub"

func TestParse(t *testing.T) {
	const inputFile = fileYoujoSenkiV01

	f, err := os.Open(inputFile)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			t.Skipf("missing optional test file %#v", inputFile)

			return
		} else {
			t.Fatal(err)
		}
	}
	defer f.Close()

	ef, err := epub.Open(f)
	if err != nil {
		t.Fatal(err)
	}
	defer ef.Close()

	if ef == nil {
		t.Fatal("ef = nil; want non-nil")
	}

	if got, want := len(ef.Container.RootFiles), 1; got == want {
		rootFile := ef.Container.RootFiles[0]
		if got, want := rootFile.FullPath, "OEBPS/content.opf"; got != want {
			t.Errorf("rootFile.FullPath = %#v; want %#v", got, want)
		}
		if got, want := rootFile.MediaType, "application/oebps-package+xml"; got != want {
			t.Errorf("rootFile.MediaType = %#v; want %#v", got, want)
		}
	} else {
		t.Fatalf("len(ef.Container.RootFiles) = %#v; want %#v", got, want)
	}

	if ef.Package == nil {
		t.Fatal("ef.Package = nil; want non-nil")
	}

	text, err := ef.Text()
	if err != nil {
		t.Fatalf(err.Error())
	}

	if text == "" {
		t.Fatalf("ef.Text() is an empty string; want non-empty string")
	}
}
