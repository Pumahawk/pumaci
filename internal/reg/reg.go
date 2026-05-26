package reg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"

	"github.com/Pumahawk/pumaci/internal/log"
)

type Client struct {
	tokens map[string]string
	mxTk   sync.RWMutex
	mx     sync.Mutex
}

func (c *Client) getTk(scope string) string {
	c.mxTk.RLock()
	defer c.mxTk.RUnlock()
	return c.tokens[scope]
}

func (c *Client) setTk(scope, token string) {
	c.mxTk.Lock()
	defer c.mxTk.Unlock()
	if c.tokens == nil {
		c.tokens = make(map[string]string)
	}
	c.tokens[scope] = token
}

type ManifestResDto struct {
	raw any
}

func (m *ManifestResDto) IsIndex() bool {
	if mapv, ok := m.raw.(map[string]any); ok {
		if mediaType, ok := mapv["mediaType"]; ok {
			return mediaType == "application/vnd.oci.image.index.v1+json"
		} else {
			log.Debug("not found media type in manifest")
		}
	} else {
		log.Debug("unexpected raw type")
	}
	return false
}

func (m *ManifestResDto) Raw() string {
	bf := &bytes.Buffer{}
	jd := json.NewEncoder(bf)
	jd.SetIndent("", "  ")
	if err := jd.Encode(m.raw); err != nil {
		panic(err)
	}
	return bf.String()
}

func (m *ManifestResDto) Config() (string, bool) {
	if mapv, ok := m.raw.(map[string]any); ok {
		if configR, ok := mapv["config"]; ok {
			if config, ok := configR.(map[string]any); ok {
				if dgsR, ok := config["digest"]; ok {
					if dgs, ok := dgsR.(string); ok {
						return dgs, true
					}
				} else {
					log.Warn("not found .config.digest")
				}
			} else {
				log.Warn("unexpected type .config")
			}
		} else {
			log.Warn("not found config in manifest")
		}
	} else {
		log.Warn("unexpected raw type")
	}
	return "", false
}

func (m *ManifestResDto) LookupPlatform(marc, mos string) (string, bool) {
	if mapv, ok := m.raw.(map[string]any); ok {
		if manifestsRaw, ok := mapv["manifests"]; ok {
			if manifests, ok := manifestsRaw.([]any); ok {
				for i, mnR := range manifests {
					var cmarc, cmos string
					if mn, ok := mnR.(map[string]any); ok {
						if plR, ok := mn["platform"]; ok {
							if pl, ok := plR.(map[string]any); ok {
								if marcR, ok := pl["architecture"]; ok {
									if cmarc, ok = marcR.(string); !ok {
										log.Warn("invalid type manifest.manifests[%d].platform.architecture, expected string", i)
									}
								} else {
									log.Warn("missing manifest.manifests[%d].platform.architecture", i)
								}
								if mosR, ok := pl["os"]; ok {
									if cmos, ok = mosR.(string); !ok {
										log.Warn("invalid type manifest.manifests[%d].platform.os, expected string", i)
									}
								} else {
									log.Warn("missing manifest.manifests[%d].platform.os", i)
								}
							} else {
								log.Warn("invalid type manifest.manifests[%d].platform, expected map[any]", i)
							}
							if cmarc == marc && cmos == mos {
								if digestR, ok := mn["digest"]; ok {
									if digest, ok := digestR.(string); ok {
										log.Debug("find manifest arc=%q os=%q digest=%q", marc, mos, digest)
										return digest, true
									} else {
										log.Warn("invalid type manifest[%d].digest, expected string", i)
									}
								} else {
									log.Warn("missing digest in manifests[%d]", i)
								}
							}
						} else {
							log.Warn("not found manifest.manifests[%d].platform", i)
						}
					} else {
						log.Warn("unexpected type manifest.manifests[%d]", i)
					}

				}
			} else {
				log.Warn("unexpected type manifest.manifests[]")
			}
		} else {
			log.Warn("manifests property not present")
		}
	} else {
		log.Warn("unmatch type manifest")
	}
	return "", false
}

func (c *Client) Manifest(img *Image, digest string) (*ManifestResDto, error) {
	var resB any

	if digest == "" {
		log.Debug("digest is empty, tag=%q", img.Tag)
		digest = img.Tag
	}

	rawUrl, err := url.JoinPath(img.Url, img.Path, "manifests", digest)
	if err != nil {
		return nil, err
	}
	log.Debug("manifest raw_url=%q", rawUrl)

	req, err := c.newRequest(img, "pull", "GET", rawUrl, nil)

	res, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode >= 200 && res.StatusCode < 300 {
		if err := json.NewDecoder(res.Body).Decode(&resB); err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("unexpected status_code=%d", res.StatusCode)
	}

	return &ManifestResDto{resB}, nil
}

func (c *Client) Blob(img *Image, digest string) (io.ReadCloser, error) {
	rawUrl, err := url.JoinPath(img.Url, img.Path, "blobs", digest)
	if err != nil {
		return nil, err
	}
	log.Debug("blob raw_url=%q", rawUrl)

	req, err := c.newRequest(img, "pull", "GET", rawUrl, nil)

	res, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	if res.StatusCode >= 200 && res.StatusCode < 300 {
		return res.Body, nil
	} else {
		defer res.Body.Close()
		return nil, fmt.Errorf("unexpected status_code=%d", res.StatusCode)
	}
}

func (c *Client) tokenize(wwwAuth string) (string, error) {
	log.Debug("start tokenize www=%q", wwwAuth)
	c.mx.Lock()
	defer c.mx.Unlock()

	if !regexp.MustCompile("^Bearer ").MatchString(wwwAuth) {
		return "", fmt.Errorf("expected brearer wwwAuth=%q", wwwAuth)
	}

	wwwAuth = strings.TrimPrefix(wwwAuth, "Bearer ")

	ps := strings.Split(wwwAuth, ",")
	rx := regexp.MustCompile("^([A-Za-z]*)=\"([^\"]*)\"")
	var realm, scope string
	urlv := make(url.Values)
	for _, p := range ps {
		vs := rx.FindAllStringSubmatch(p, -1)
		if len(vs) != 1 {
			return "", fmt.Errorf("unexpected token param number parsing param=%q size=%d", p, len(vs))
		}
		v := vs[0]
		if len(v) != 3 {
			return "", fmt.Errorf("unexpected token param groups parsing param=%q size=%d", p, len(v))
		}
		key, value := v[1], v[2]
		if key == "realm" {
			realm = value
			log.Debug("find realm=%q", realm)
			continue
		}
		if key == "scope" {
			scope = value
			log.Debug("find scope=%q", scope)
		}
		log.Debug("set parameter %q=%q", key, value)
		urlv.Add(key, value)
	}
	if realm == "" {
		return "", fmt.Errorf("missing realm paramater wwwAuth=%q", wwwAuth)
	}
	u, err := url.Parse(realm)
	if err != nil {
		return "", fmt.Errorf("parse realm realm=%q: %w", err)
	}
	u.RawQuery = urlv.Encode()
	rawUrl := u.String()
	log.Debug("tokenize rawUrl=%q", rawUrl)
	res, err := http.Get(rawUrl)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	type tok struct{ Token string }
	dto := &tok{}
	if err := json.NewDecoder(res.Body).Decode(dto); err != nil {
		return "", fmt.Errorf("parse response token: %w", err)
	}
	c.setTk(scope, dto.Token)
	return dto.Token, nil
}

func (c *Client) newRequest(img *Image, tokenMethod string, method string, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("accept", "*/*")

	scope := getScopeFromImage(img, tokenMethod)
	log.Debug("lookup tk scope=%q", scope)
	if tk := c.getTk(scope); tk != "" {
		log.Debug("find stored token scope=%q", scope)
		req.Header.Add("authorization", "Bearer "+tk)
	}

	return req, nil
}

func (c *Client) doRequest(req *http.Request) (*http.Response, error) {
	cl := &http.Client{}
	res, err := cl.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	if res.StatusCode == 401 {
		defer res.Body.Close()
		log.Debug("authentication required, status_code=%d", res.StatusCode)
		tokenize, err := c.tokenize(res.Header.Get("www-authenticate"))
		if err != nil {
			return nil, err
		}
		req.Header.Add("authorization", "Bearer "+tokenize)

		res2, err := cl.Do(req)
		if err != nil {
			return nil, fmt.Errorf("do request: %w", err)
		}
		if res2.StatusCode < 200 || res2.StatusCode >= 300 {
			defer res2.Body.Close()
			return nil, fmt.Errorf("unable to retry request, status_code=%d", res2.StatusCode)
		}
		return res2, nil
	} else {
		return res, nil
	}
}

type Image struct {
	Url   string
	Path  string // TODO testing
	Image string
	Tag   string
}

func ParseImage(image string) (*Image, error) {
	var imageName, imageTag string
	pt := strings.Split(image, "/")
	u := &url.URL{}

	u.Host = pt[0]

	var imageParts []string
	for i, p := range pt {
		if len(pt)-1 == i {
			imgs := strings.Split(p, ":")
			imageParts = append(imageParts, imgs[0])
		} else {
			imageParts = append(imageParts, p)
		}
	}

	imageName = strings.Join(imageParts, "/")

	ppt := strings.Split(pt[len(pt)-1], ":")
	if len(ppt) > 1 {
		imageTag = ppt[1]
	} else {
		imageTag = "latest"
	}

	if u.Scheme == "" {
		u.Scheme = "https"
	}
	u.Path = "v2"
	return &Image{
		Url:   u.String(),
		Path:  strings.Join(imageParts[1:], "/"), // TODO refactor and testing
		Image: imageName,
		Tag:   imageTag,
	}, nil
}

func getScopeFromImage(image *Image, method string) string {
	return fmt.Sprintf("repository:%s:%s", image.Path, method)
}
