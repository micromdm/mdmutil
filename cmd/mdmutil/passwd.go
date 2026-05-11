package main

import (
	"crypto/rand"
	"encoding/base64"
	"flag"
	"fmt"
	"os"

	mupasswd "github.com/micromdm/mdmutil/passwd"
	"github.com/micromdm/plist"
)

func passwd(name string, args []string, usage func()) int {
	f := flag.NewFlagSet(name, flag.ExitOnError)
	var (
		flB64      = f.Bool("b64", false, "Output base64-encoded Plist")
		flPassword = f.String("password", os.Getenv("MDMUTIL_PASSWORD"), "password to hash (also as MDMUTIL_PASSWORD environment var)")
	)
	cmdUsage(f, usage, nil, "")

	if err := f.Parse(args); err != nil {
		flagUsageExit(f, "failed to parse args", 2)
	}

	if *flPassword == "" {
		flagUsageExit(f, "no password supplied", 2)
	}

	ph, err := mupasswd.HashPassword(rand.Reader, *flPassword)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	plist, err := plist.MarshalIndent(ph, "  ")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if *flB64 {
		fmt.Println(base64.StdEncoding.EncodeToString(plist))
	} else {
		fmt.Print(string(plist))
	}

	return 0
}
