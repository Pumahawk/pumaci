package reg

import (
	"encoding/json"
	"io"

	"github.com/Pumahawk/pumaci/internal/log"
)

type Manifest struct {
	m any
}

func (m *Manifest) Index() (*IndexManifest, bool) {
	v, ok := m.m.(IndexManifest)
	if ok {
		return &v, true
	} else {
		log.Debug("manifest not index type=%T", m.m)
		return nil, false
	}
}

func (m *Manifest) Image() (*ImageManifest, bool) {
	v, ok := m.m.(ImageManifest)
	if ok {
		return &v, true
	} else {
		log.Debug("manifest not image type=%T", m.m)
		return nil, false
	}
}

func (m *Manifest) Encode(w io.Writer) error {
	jenc := json.NewEncoder(w)
	jenc.SetIndent("", " ")
	return jenc.Encode(m.m)
}

type ImageManifest struct {
	SchemaVersion int            `json:"schemaVersion"`
	MediaType     string         `json:"mediaType"`
	Config        ConfigInfo     `json:"config"`
	Layers        []LayerInfo    `json:"layers"`
	Extra         map[string]any `json:"-"`
}

type ConfigInfo struct {
	MediaType string         `json:"mediaType"`
	Digest    string         `json:"digest"`
	Size      int            `json:"size"`
	Extra     map[string]any `json:"-"`
}

type SubjectInfo struct {
	MediaType string `json:"mediaType"`
	Digest    string `json:"digest"`
	Size      int    `json:"size"`
	Extra     any    `json:"-"`
}

type LayerInfo struct {
	MediaType string `json:"mediaType"`
	Digest    string `json:"digest"`
	Size      int    `json:"size"`
	Extra     any    `json:"-"`
}

type IndexManifest struct {
	SchemaVersion int            `json:"schemaVersion"`
	MediaType     string         `json:"mediaType"`
	Manifests     []ManifestInfo `json:"manifests"`
	Extra         map[string]any `json:"-"`
}

type ManifestInfo struct {
	MediaType string   `json:"mediaType"`
	Size      int      `json:"size"`
	Digest    string   `json:"digest"`
	Platform  Platform `json:"platform"`
	Extra     any      `json:"-"`
}

type Platform struct {
	Architecture string `json:"architecture"`
	Os           string `json:"os"`
	Extra        any    `json:"-"`
}
