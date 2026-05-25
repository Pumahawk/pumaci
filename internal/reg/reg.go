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
)

type Client struct {
	token string
	mx    sync.Mutex
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
	req, err := http.NewRequest("GET", rawUrl, nil)
	if err != nil {
		return nil, err
	}
	res, err := cl.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode == 401 {
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
		if err := json.NewDecoder(res2.Body).Decode(&resB); err != nil {
			return "", err
		}
	} else {
		if err := json.NewDecoder(res.Body).Decode(&resB); err != nil {
			return "", err
		}
	}
	return resB, nil
}

func (c *Client) tokenize(wwwAuth string) (string, error) {
	c.mx.Lock()
	defer c.mx.Unlock()
	if c.token != "" {
		return c.token, nil
	}

	if !regexp.MustCompile("^Bearer ").MatchString(wwwAuth) {
		return "", fmt.Errorf("expected brearer wwwAuth=%q", wwwAuth)
	}

	wwwAuth = strings.TrimPrefix(wwwAuth, "Bearer ")

	ps := strings.Split(wwwAuth, ",")
	rx := regexp.MustCompile("^([A-Za-z]*)=\"([^\"]*)\"")
	var realm string
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
		if v[1] == "realm" {
			realm = v[2]
			continue
		}
		urlv.Add(v[1], v[2])
	}
	if realm == "" {
		return "", fmt.Errorf("missing realm paramater wwwAuth=%q", wwwAuth)
	}
	u, err := url.Parse(realm)
	if err != nil {
		return "", fmt.Errorf("parse realm realm=%q: %w", err)
	}
	u.RawQuery = urlv.Encode()
	res, err := http.Get(u.String())
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	bf := &bytes.Buffer{}
	io.Copy(bf, res.Body)
	type tok struct{ Token string }
	dto := &tok{}
	if err := json.NewDecoder(bf).Decode(dto); err != nil {
		return "", fmt.Errorf("parse response token: %w", err)
	}
	c.token = dto.Token
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
