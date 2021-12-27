package main

import (
	"compress/gzip"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"
)

var (
	inFilename  = flag.String("in", "", "Input filename")
	outFilename = flag.String("out", "", "Output filename")
)

func main() {
	statusCode := run()
	os.Exit(statusCode)
}

func run() (statusCode int) {
	if flagsAreInvalid() {
		return 1
	}

	inFile, err := os.Open(*inFilename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not open input file: %v\n", err)
		return 1
	}
	defer func() {
		if err := inFile.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "closing input file: %v\n", err)
		}
	}()

	var charJSON []string
	if err = json.NewDecoder(inFile).Decode(&charJSON); err != nil {
		fmt.Fprintf(os.Stderr, "could not decode input as a JSON string array: %v\n", err)
		return 1
	}

	var charmap []byte
	for _, char := range charJSON {
		decoded, err := hex.DecodeString(char)
		if err != nil {
			fmt.Fprintf(os.Stderr, "could not decode character as hex: %v\n", err)
			return 1
		}
		charmap = append(charmap, decoded...)
	}

	outFile, err := os.Create(*outFilename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not open output file: %v\n", err)
		return 1
	}
	defer func() {
		if err := outFile.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "closing output file: %v\n", err)
		}
	}()

	gz := gzip.NewWriter(outFile)
	defer func() {
		if err := gz.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "closing gzip writer: %v\n", err)
		}
	}()

	if _, err = gz.Write(charmap); err != nil {
		fmt.Fprintf(os.Stderr, "could not write output file: %v\n", err)
		return 1
	}

	return 0
}

func flagsAreInvalid() (invalid bool) {
	flag.Parse()

	if *inFilename == "" {
		fmt.Fprintf(os.Stderr, "argument required for -in flag\n")
		invalid = true
	} else {
		fi, err := os.Stat(*inFilename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			invalid = true
		} else if fi.IsDir() {
			fmt.Fprintf(os.Stderr, "input is a directory\n")
			invalid = true
		}
	}

	if *outFilename == "" {
		fmt.Fprintf(os.Stderr, "argument required for -out flag\n")
		invalid = true
	} else {
		fi, err := os.Stat(*inFilename)
		if err == nil && fi.IsDir() {
			fmt.Fprintf(os.Stderr, "output is a directory\n")
			invalid = true
		}
	}

	return
}
