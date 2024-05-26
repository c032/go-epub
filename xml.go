package epub

import (
	"errors"
	"strings"
)

var (
	ErrInvalidLinear error = errors.New("invalid linear")
)

type Container struct {
	RootFiles []ContainerRootFile `xml:"rootfiles>rootfile"`
}

type ContainerRootFile struct {
	FullPath  string `xml:"full-path,attr"`
	MediaType string `xml:"media-type,attr"`
}

type Package struct {
	Metadata *PackageMetadata `xml:"metadata"`
	Manifest *PackageManifest `xml:"manifest"`
	Spine    *PackageSpine    `xml:"spine"`
}

type PackageMetadata struct {
	// TODO
}

type PackageManifest struct {
	Items []PackageManifestItem `xml:"item"`
}

func (pm *PackageManifest) Item(id string) (*PackageManifestItem, bool) {
	// TODO: If this `for` loop ever becomes a bottleneck, use a map instead of
	// a slice.

	for _, pmi := range pm.Items {
		if pmi.ID == id {
			return &pmi, true
		}
	}

	return nil, false
}

type PackageManifestItem struct {
	ID        string `xml:"id,attr"`
	Href      string `xml:"href,attr"`
	MediaType string `xml:"media-type,attr"`
}

type PackageSpine struct {
	ItemRefs []PackageSpineItemRef `xml:"itemref"`
}

type PackageSpineItemRef struct {
	IDRef     string `xml:"idref,attr"`
	RawLinear string `xml:"linear,attr"`
}

func (psir *PackageSpineItemRef) IsLinear() (bool, error) {
	linear := strings.ToLower(psir.RawLinear)

	if linear == "yes" {
		return true, nil
	}
	if linear == "no" {
		return false, nil
	}

	return false, ErrInvalidLinear
}
