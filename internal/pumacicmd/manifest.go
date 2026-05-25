package pumacicmd

import (
	"encoding/json"
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
		if err := json.NewEncoder(os.Stdout).Encode(data); err != nil {
			fmt.Fprintf(os.Stderr, "decode data: %s\n", err)
			os.Exit(1)
		}
		return nil
	},
}
