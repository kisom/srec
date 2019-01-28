package main

import (
	"flag"
	"os"
	"path/filepath"
	"strings"

	"github.com/kisom/goutils/die"
	"github.com/kisom/srec"
)

func dumpFile(path string, mode32 bool) error {
	ext := filepath.Ext(path)
	outPath := path
	if ext != "" {
		outPath = strings.TrimSuffix(outPath, ext)
	}
	outPath += ".sr"

	inFile, err := os.Open(path)
	if err != nil {
		return err
	}
	defer inFile.Close()

	outFile, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	if mode32 {
		err = srec.Copy32([]byte("HDR"), 0, inFile, outFile)
	} else {
		err = srec.Copy16([]byte("HDR"), 0, inFile, outFile)
	}

	if err != nil {
		return err
	}

	return nil
}

func main() {
	var mode32 bool

	flag.BoolVar(&mode32, "32", false, "32-bit mode")
	flag.Parse()

	for _, path := range flag.Args() {
		err := dumpFile(path, mode32)
		die.If(err)
	}
}
