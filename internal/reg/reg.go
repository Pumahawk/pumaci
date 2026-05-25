package reg

import (
	"encoding/json"
	"fmt"
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

func (c *Client) Manifest(image string) (any, error) {
	cl := &http.Client{}
	img, err := ParseImage(image)
	var resB any
	if err != nil {
		return nil, fmt.Errorf("parse image %q: %w", image, err)
	}

	rawUrl, err := url.JoinPath(img.Url, img.Path, "manifests", img.Tag)
	if err != nil {
		return nil, err
	}
	log.Debug("manifest raw_url=%q", rawUrl)

	req, err := http.NewRequest("GET", rawUrl, nil)
	if err != nil {
		return nil, err
	}

	scope := getScopeFromImage(img, "pull")
	log.Debug("lookup tk scope=%q", scope)
	if tk := c.getTk(scope); tk != "" {
		log.Debug("find stored token scope=%q", scope)
		req.Header.Add("authorization", "Bearer "+tk)
	}

	res, err := cl.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode == 401 {
		log.Debug("authentication required, status_code=%d", res.StatusCode)
		tokenize, err := c.tokenize(res.Header.Get("www-authenticate"))
		if err != nil {
			return "", err
		}
		req.Header.Add("authorization", "Bearer "+tokenize)

		res2, err := cl.Do(req)
		if err != nil {
			return nil, fmt.Errorf("do request: %w", err)
		}
		defer res2.Body.Close()
		if res2.StatusCode < 200 || res2.StatusCode >= 300 {
			return "", fmt.Errorf("unable to retry request, status_code=%d", res2.StatusCode)
		}

		if err := json.NewDecoder(res2.Body).Decode(&resB); err != nil {
			return "", err
		}
	} else if res.StatusCode >= 200 && res.StatusCode < 300 {
		if err := json.NewDecoder(res.Body).Decode(&resB); err != nil {
			return "", err
		}
	} else {
		return "", fmt.Errorf("unexpected status_code=%d", res.StatusCode)
	}
	return resB, nil
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
