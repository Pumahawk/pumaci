package pumacicmd

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/Pumahawk/pumaci/internal/cmd"
	"github.com/Pumahawk/pumaci/internal/log"
	"github.com/Pumahawk/pumaci/internal/reg"
)

var Manifest = &cmd.Cmd{
	CName: "manifest",
	CRun: func(s []string) error {
		var showIndex bool
		var march, mos string
		fs := flag.NewFlagSet("", flag.ExitOnError)
		fs.BoolVar(&showIndex, "index", false, "")
		fs.StringVar(&march, "arc", "amd64", "")
		fs.StringVar(&mos, "os", "linux", "")
		fs.Parse(s)

		image := fs.Arg(0)
		if image == "" {
			fmt.Fprintf(os.Stderr, "missing image arg\n")
			os.Exit(1)
		}

		digest := fs.Arg(1)

		cl := &reg.Client{}
		img, err := reg.ParseImage(image)
		if err != nil {
			fmt.Fprintf(os.Stderr, "parse image %q: %w\n", image, err)
			os.Exit(1)
		}

		if digest == "" || digest == "config" || strings.HasPrefix(digest, "layer:") {
			log.Debug("retrieve metadata digest=%q", digest)
			data, err := cl.Manifest(img, "")
			if err != nil {
				fmt.Fprintf(os.Stderr, "get manifest: %s\n", err)
				os.Exit(1)
			}

			isIndex := data.IsIndex()
			log.Debug("is index=%v", isIndex)
			if isIndex && !showIndex {
				if digest, ok := data.LookupPlatform(march, mos); ok {
					data, err = cl.Manifest(img, digest)
					if err != nil {
						fmt.Fprintf(os.Stderr, "unable to get manifest from index: %s\n", err)
						os.Exit(1)
					}
				} else {
					fmt.Fprintf(os.Stderr, "not found manifest arch=%q os=%q\n", march, mos)
				}
			}
			if layer, ok := strings.CutPrefix(digest, "layer:"); ok {
				log.Debug("get blob layer=%q", layer)
				n, err := strconv.ParseInt(layer, 10, 64)
				if err != nil {
					fmt.Fprintf(os.Stderr, "unable to parse layer=%q: %s\n", layer, err)
					os.Exit(1)
				}
				ls, err := data.Layers()
				if err != nil {
					fmt.Fprintf(os.Stderr, "unable to extract layers from manifest: %s\n", err)
					os.Exit(1)
				}
				if len(ls) == 0 {
					fmt.Fprintf(os.Stderr, "not found layers from manigest\n")
					os.Exit(1)
				}
				idx := int(n)
				if idx < 0 {
					idx = len(ls) + idx
				}
				if idx < 0 || idx >= len(ls) {
					fmt.Fprintf(os.Stderr, "invalid defined layer index (%d/%d)", idx+1, len(ls))
					os.Exit(1)
				}
				log.Debug("find digest (%d/%d)", idx+1, len(ls))
				digest := ls[idx]
				log.Debug("retrieve blob digest=%q", digest)
				blob, err := cl.Blob(img, digest)
				if err != nil {
					fmt.Fprintf(os.Stderr, "unable to retrieve blob digest=%q: %s\n", digest, err)
					os.Exit(1)
				}
				defer blob.Close()
				if _, err := io.Copy(os.Stdout, blob); err != nil {
					fmt.Fprintf(os.Stderr, "unable to write blob to stdout: %w\n", err)
				}
				return nil
			} else if digest == "config" {
				log.Debug("get config blob")
				if dcg, ok := data.Config(); ok {
					blob, err := cl.Blob(img, dcg)
					if err != nil {
						fmt.Fprintf(os.Stderr, "unable to retrieve blob digest=%q: %s\n", dcg, err)
						os.Exit(1)
					}
					defer blob.Close()
					var jsonMap any
					// TODO check error
					json.NewDecoder(blob).Decode(&jsonMap)
					jen := json.NewEncoder(os.Stdout)
					jen.SetIndent("", " ")
					if err := jen.Encode(jsonMap); err != nil {
						fmt.Fprintf(os.Stderr, "unable to write blob to stdout: %w\n", err)
					}
				} else {
					fmt.Fprintf(os.Stderr, "digest config not found\n")
					os.Exit(1)
				}
			} else {
				fmt.Println(data.Raw())
			}
			return nil
		} else {
			log.Debug("retrieve blob digest=%q", digest)
			blob, err := cl.Blob(img, digest)
			if err != nil {
				fmt.Fprintf(os.Stderr, "unable to retrieve blob digest=%q: %s\n", digest, err)
				os.Exit(1)
			}
			defer blob.Close()
			if _, err := io.Copy(os.Stdout, blob); err != nil {
				fmt.Fprintf(os.Stderr, "unable to write blob to stdout: %w\n", err)
			}
			return nil
		}
	},
}
