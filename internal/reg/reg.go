package reg

import (
	"net/url"
	"strings"
)

type Client struct {
}

func (c *Client) Manifest(image string) {
}

type baseUrl struct {
	base  string
	image string
	tag   string
}

func getBaseUrl(image string) (*baseUrl, error) {
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
	return &baseUrl{
		base:  u.String(),
		image: imageName,
		tag:   imageTag,
	}, nil
}
