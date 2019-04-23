package main

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
	"flag"
)

type ArpEntry struct {
	IpAddress string
	MacAddress string
}

var outFile string
var quiet bool
func init() {
	flag.StringVar(&outFile, "outfile", "", "file to write logs to")
	flag.BoolVar(&quiet, "quiet", false, "supress output")
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
	cmd := exec.Command("arp", "-a")
	output, err := cmd.CombinedOutput()

	if err != nil {
		panic(err)
	}

	return parseArpTable(string(output))
}

func detectChanges(oldEntries []*ArpEntry, newEntries[]*ArpEntry)  {
	if oldEntries == nil {
		return
	}

	for _,entry := range oldEntries {
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
			if isError(err) { return }
			defer file.Close()
		}
		text := "timeStamp=" + timeValue + " ip=" + matchedEntry.IpAddress + " oldMac=" + entry.MacAddress + " newMac=" + matchedEntry.MacAddress + " Message='MAC address change detected'\n"
		f, err := os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			panic(err)
		}

		defer f.Close()

		if _, err = f.WriteString(text); err != nil {
			panic(err)
		}
	}
}

func getMatchingEntry(entry *ArpEntry, entries []*ArpEntry) *ArpEntry {
	for _,potentialMatch := range entries {
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

	regex := regexp.MustCompile(`(\d+.\d+.\d+.\d+).* at (.*) on`)

	var entries = []*ArpEntry{}

	for _,line := range lines {
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
	return strings.Split(arpOutput, "[ethernet]")
}
