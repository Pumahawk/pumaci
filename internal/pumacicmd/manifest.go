package pumacicmd

import (
	"fmt"
	"os"

	"github.com/Pumahawk/pumaci/internal/cmd"
	"github.com/Pumahawk/pumaci/internal/reg"
)

var Manifest = &cmd.Cmd{
	CName: "manifest",
	CRun: func(s []string) error {
		if len(s) < 1 {
			fmt.Fprintf(os.Stderr, "missing image arg\n")
			os.Exit(1)
		}
		image := s[0]
		cl := &reg.Client{}
		data, err := cl.Manifest(image)
		if err != nil {
			fmt.Fprintf(os.Stderr, "get manifest: %s\n", err)
			os.Exit(1)
		}
		fmt.Println(data.Raw())
		return nil
	},
}
