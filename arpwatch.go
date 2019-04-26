package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"time"
	"log"
)

type ArpEntry struct {
	IpAddress  string
	MacAddress string
}

var outFile string
var quiet bool
var rlogServer string

func init() {
	flag.StringVar(&outFile, "outfile", "", "file to write logs to")
	flag.BoolVar(&quiet, "quiet", false, "supress output")
	flag.StringVar(&rlogServer, "server", "", "remote server to log to")
}

func main() {
	flag.Parse()
	enableDetection()
}

func enableDetection() {
	if quiet != true {
		fmt.Println("Listening for ARP changes...")
	}
	entries := getCurrentEntries()
	for {
		currentEntries := getCurrentEntries()
		detectChanges(entries, currentEntries)
		entries = currentEntries
		time.Sleep(100 * time.Millisecond)
	}
}

func getCurrentEntries() []*ArpEntry {
	if runtime.GOOS == "linux" {
		b, err := ioutil.ReadFile("/proc/net/arp")
		if err != nil {
			panic(err)
		}
		output := string(b)
		return parseArpTable(output)
	} else {
		cmd := exec.Command("arp", "-a")
		output, err := cmd.CombinedOutput()
		if err != nil {
			panic(err)
		}
		return parseArpTable(string(output))
	}

}

func detectChanges(oldEntries []*ArpEntry, newEntries []*ArpEntry) {
	if oldEntries == nil {
		return
	}

	for _, entry := range oldEntries {
		matchedEntry := getMatchingEntry(entry, newEntries)

		if matchedEntry != nil {
			if entryHasChanged(entry, matchedEntry) {
				t := time.Now()
				changeTime := t.Format(time.RFC3339)
				tellTheUser(entry, matchedEntry, changeTime)
			}
		}
	}
}

func entryHasChanged(oldEntry *ArpEntry, newEntry *ArpEntry) bool {
	return oldEntry.MacAddress != newEntry.MacAddress && newEntry.MacAddress != "(incomplete)" && oldEntry.MacAddress != "(incomplete)"
}

func isError(err error) bool {
	if err != nil {
		fmt.Println(err.Error())
	}

	return (err != nil)
}

func tellTheUser(entry *ArpEntry, matchedEntry *ArpEntry, timeValue string) {
	if quiet != true {
		fmt.Println("Mac address change detected for same IP Address")
		fmt.Printf("IP[%s] - %s => %s\n", matchedEntry.IpAddress, entry.MacAddress, matchedEntry.MacAddress)
	}
	if outFile != "" {
		fileName := outFile
		if _, err := os.Stat(fileName); os.IsNotExist(err) {
			var file, err = os.Create(fileName)
			if isError(err) {
				return
			}
			defer file.Close()
		}
		file, err := os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			panic(err)
		}
		logger := log.New(file, "", log.LstdFlags|log.Lshortfile)

		text := "timeStamp=" + timeValue + " ip=" + matchedEntry.IpAddress + " oldMac=" + entry.MacAddress + " newMac=" + matchedEntry.MacAddress + " Message='MAC address change detected'"
		logger.Println(text)
	}
}

func getMatchingEntry(entry *ArpEntry, entries []*ArpEntry) *ArpEntry {
	for _, potentialMatch := range entries {
		if potentialMatch.IpAddress == entry.IpAddress {
			return potentialMatch
		}
	}

	return nil
}

func parseArpTable(arpOutput string) []*ArpEntry {
	lines := splitOutputIntoArray(arpOutput)
	entries := mapLinesToObjects(lines)
	return entries
}

func mapLinesToObjects(lines []string) []*ArpEntry {
	var entries = []*ArpEntry{}
	var regex *regexp.Regexp

	if runtime.GOOS == "darwin" {
		regex = regexp.MustCompile(`(\d+.\d+.\d+.\d+).* at (.*) on`)

	} else {
		regex = regexp.MustCompile(`(\d+.\d+.\d+.\d+).*([0-9a-fA-F]{2}[:-][0-9a-fA-F]{2}[:-][0-9a-fA-F]{2}[:-][0-9a-fA-F]{2}[:-][0-9a-fA-F]{2}[:-][0-9a-fA-F]{2})`)
	}

	for _, line := range lines {
		values := regex.FindStringSubmatch(line)
		if len(values) > 0 {
			entry := new(ArpEntry)
			entry.IpAddress = values[1]
			entry.MacAddress = values[2]

			entries = append(entries, entry)
		}
	}
	return entries
}

func splitOutputIntoArray(arpOutput string) []string {
	var stringArray []string
	if runtime.GOOS == "darwin" {
		stringArray = strings.Split(arpOutput, "[ethernet]")
	} else {
		stringArray = strings.Split(arpOutput, " \n")
	}

	return stringArray
}
