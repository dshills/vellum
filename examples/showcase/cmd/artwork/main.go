// Artwork writes original bundle icons into one new, explicitly named directory.
package main

import (
	"flag"
	"fmt"
	"github.com/dshills/vellum/examples/showcase/demo"
	"image/png"
	"os"
	"path/filepath"
)

func main() {
	output := flag.String("output", "", "new output directory")
	flag.Parse()
	if *output == "" {
		fmt.Fprintln(os.Stderr, "-output is required")
		os.Exit(1)
	}
	if e := write(*output); e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}
func write(output string) error {
	if e := os.Mkdir(output, 0755); e != nil {
		return e
	}
	for _, name := range []string{"forma", "luma", "pulse"} {
		img, e := demo.AppIcon(name)
		if e != nil {
			return e
		}
		f, e := os.OpenFile(filepath.Join(output, name+".png"), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
		if e != nil {
			return e
		}
		e = png.Encode(f, img)
		closeErr := f.Close()
		if e != nil {
			return e
		}
		if closeErr != nil {
			return closeErr
		}
	}
	return nil
}
