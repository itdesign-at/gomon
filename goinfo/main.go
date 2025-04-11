package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/itdesign-at/golib/commandLine"
	"github.com/itdesign-at/golib/keyvalue"

	"itdesign.at/gomon/pkg"
)

func main() {
	args := commandLine.Parse(os.Args)
	if printVersionOrHelp(args) {
		os.Exit(0)
	}
	content, err := os.ReadFile(pkg.F_hostsExportedFile)
	if err != nil {
		fmt.Printf("error reading %s: %s\n", pkg.F_hostsExportedFile, err)
		os.Exit(0)
	}
	var data map[string]keyvalue.Record
	err = json.Unmarshal(content, &data)
	if err != nil {
		fmt.Printf("error unmarshalling %s: %s\n", pkg.F_hostsExportedFile, err)
		os.Exit(0)
	}
	host := args.String("h")
	if host == "" {
		fmt.Println("no hostname provided, requires -h <hostname>")
		os.Exit(0)
	}
	var ok bool
	var record keyvalue.Record
	if record, ok = data[host]; ok {
		printToConsole(host, args.String("layout"), record)
	} else {
		fmt.Printf("host %s not found\n", host)
		os.Exit(0)
	}
	os.Exit(0)
}

func printToConsole(host, layout string, record keyvalue.Record) {
	var description = record.String("D")
	switch layout {
	case "withIP":
		ipAddress := record.String("IP")
		if ipAddress != "" {
			if ipAddress != host {
				description = fmt.Sprintf("%s (%s)", host, ipAddress)
			}
		}
	default:
		break
	}
	if description != "" {
		fmt.Println(description)
	} else {
		fmt.Println("no description found")
	}
}

func printVersionOrHelp(args keyvalue.Record) bool {
	if args.Exists("version") {
		fmt.Println(pkg.VERSION)
		return true
	}
	if args.Exists("help") {
		helpText := `goinfo reads and prints the description of one host
from the file /opt/watchit/var/etc/hosts-exported.json

The layout option "withIP" will print the IP address, too (if available).

Usage:
goinfo -h <hostname> [--layout <layout>] [--version] [--help] [--about]
goinfo -h fc-switch.demo.at
goinfo -h fc-switch.demo.at --layout withIP`
		fmt.Println(helpText)
		return true
	}
	if args.Exists("about") {
		fmt.Println("https://github.com/itdesign-at/gomon.git")
		return true
	}
	return false
}
