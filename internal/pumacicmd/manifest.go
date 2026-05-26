package pumacicmd

import (
	"flag"
	"fmt"
	"os"

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
		cl := &reg.Client{}
		img, err := reg.ParseImage(image)
		if err != nil {
			fmt.Fprintf(os.Stderr, "parse image %q: %w\n", image, err)
			os.Exit(1)
		}
		data, err := cl.Manifest(img, "")
		if err != nil {
			fmt.Fprintf(os.Stderr, "get manifest: %s\n", err)
			os.Exit(1)
		}

		isIndex := data.IsIndex()
		log.Debug("is index=%v", isIndex)
		if isIndex && !showIndex {
			if digest, ok := data.LookupPlatform(march, mos); ok {
				data, err := cl.Manifest(img, digest)
				if err != nil {
					fmt.Fprintf(os.Stderr, "unable to get manifest from index: %s\n", err)
					os.Exit(1)
				}
				fmt.Println(data.Raw())
			} else {
				fmt.Fprintf(os.Stderr, "not found manifest arch=%q os=%q\n", march, mos)
			}
		} else {
			fmt.Println(data.Raw())
		}
		return nil
	},
}
