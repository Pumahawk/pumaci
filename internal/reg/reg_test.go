package reg

import "testing"

func TestGetBaseUrl(t *testing.T) {
	type testData struct {
		imageIn string
		url     string
		image   string
		tag     string
	}
	in := []testData{
		{
			imageIn: "registry-1.docker.io/library/alpine",
			url:     "https://registry-1.docker.io/v2",
			image:   "registry-1.docker.io/library/alpine",
			tag:     "latest",
		},
		{
			imageIn: "registry-1.docker.io/library/alpine:1.2.3",
			url:     "https://registry-1.docker.io/v2",
			image:   "registry-1.docker.io/library/alpine",
			tag:     "1.2.3",
		},
		{
			imageIn: "customrepo/project/alpine",
			url:     "https://customrepo/v2",
			image:   "customrepo/project/alpine",
			tag:     "latest",
		},
		{
			imageIn: "customrepo:4567/project/alpine",
			url:     "https://customrepo:4567/v2",
			image:   "customrepo:4567/project/alpine",
			tag:     "latest",
		},
	}
	for _, i := range in {
		bu, err := getBaseUrl(i.imageIn)
		if err != nil {
			t.Errorf("error getBaseUrl: %s", err)
			continue
		}
		if i.tag != bu.tag {
			t.Errorf("tags: expected=%q actual %q", i.tag, bu.tag)
		}
		if i.image != bu.image {
			t.Errorf("image: expected=%q actual %q", i.image, bu.image)
		}
		if i.url != bu.base {
			t.Errorf("url: expected=%q actual=%q", i.url, bu.base)
		}
	}
}
