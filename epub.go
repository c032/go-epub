package epub

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var (
	ErrNoSpine    error = errors.New("no spine")
	ErrNoManifest error = errors.New("no manifest")
)

var _ io.Closer = (*EpubFile)(nil)

type EpubFile struct {
	zr *zip.Reader

	Container *Container
	Package   *Package
}

func (ef *EpubFile) Close() error {
	return nil
}

func (ef *EpubFile) parse() error {
	var err error

	err = ef.parseContainer()
	if err != nil {
		return fmt.Errorf("could not parse container: %w", err)
	}

	err = ef.parsePackage()
	if err != nil {
		return fmt.Errorf("could not parse package: %w", err)
	}

	return nil
}

func (ef *EpubFile) parseXml(file string, output any) error {
	var (
		err error
		f   fs.File
	)

	f, err = ef.zr.Open(file)
	if err != nil {
		return fmt.Errorf("could not open file (%s): %w", file, err)
	}
	defer f.Close()

	dec := xml.NewDecoder(f)

	err = dec.Decode(output)
	if err != nil {
		return fmt.Errorf("could not decode XML: %w", err)
	}

	return nil
}

func (ef *EpubFile) parseContainer() error {
	err := ef.parseXml("META-INF/container.xml", &ef.Container)
	if err != nil {
		return fmt.Errorf("could not parse container file: %w", err)
	}

	return nil
}

func (ef *EpubFile) extractText(file string) (string, error) {
	var (
		err error
		f   fs.File
	)

	f, err = ef.zr.Open(file)
	if err != nil {
		return "", fmt.Errorf("could not open file (%s): %w", file, err)
	}
	defer f.Close()

	var doc *goquery.Document
	doc, err = goquery.NewDocumentFromReader(f)
	if err != nil {
		return "", fmt.Errorf("could not parse HTML: %w", err)
	}

	var paragraphs []string

	doc.Find("p").Each(func(_ int, pSel *goquery.Selection) {
		content := pSel.Text()
		content = strings.TrimSpace(content)

		if content == "" {
			return
		}

		paragraphs = append(paragraphs, content)
	})

	result := strings.Join(paragraphs, "\n\n")

	return result, nil
}

func (ef *EpubFile) parsePackage() error {
	c := ef.Container
	if c == nil {
		return nil
	}
	if len(c.RootFiles) == 0 {
		return nil
	}

	rf := c.RootFiles[0]
	rfp := rf.FullPath
	if rfp == "" {
		return fmt.Errorf("root file has empty path")
	}

	err := ef.parseXml(rfp, &ef.Package)
	if err != nil {
		return fmt.Errorf("could not parse root file: %w", err)
	}

	return nil
}

func (ef *EpubFile) Text() (string, error) {
	pkg := ef.Package

	spine := pkg.Spine
	if spine == nil {
		return "", ErrNoSpine
	}

	manifest := pkg.Manifest
	if manifest == nil {
		return "", ErrNoManifest
	}

	const prefix = "OEBPS"

	var parts []string
	for _, ref := range spine.ItemRefs {
		itemID := ref.IDRef
		item, ok := manifest.Item(itemID)
		if !ok {
			return "", fmt.Errorf("missing item with ID %#v", itemID)
		}

		// There might be more types.
		if item.MediaType != "application/xhtml+xml" {
			continue
		}

		itemPath := fmt.Sprintf("%s/%s", prefix, item.Href)

		var (
			err      error
			itemText string
		)
		itemText, err = ef.extractText(itemPath)
		if err != nil {
			return "", fmt.Errorf("could not extract text from item (%#v): %w", itemPath, err)
		}

		itemText = strings.TrimSpace(itemText)
		if itemText == "" {
			continue
		}

		parts = append(parts, itemText)
	}

	result := strings.Join(parts, "\n\n")

	return result, nil
}

func Open(r io.Reader) (*EpubFile, error) {
	var (
		err error
		buf []byte
	)

	buf, err = ioutil.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("could not read: %w", err)
	}

	ra := bytes.NewReader(buf)

	var zr *zip.Reader
	zr, err = zip.NewReader(ra, int64(len(buf)))
	if err != nil {
		return nil, fmt.Errorf("could not create zip reader: %w", err)
	}

	ef := &EpubFile{
		zr: zr,
	}

	err = ef.parse()
	if err != nil {
		return nil, fmt.Errorf("could not parse epub: %w", err)
	}

	return ef, nil
}
