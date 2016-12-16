package main

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"

)

type ArpEntry struct {
	IpAddress string
	MacAddress string
}



func main() {
	enableDetection()
}

func enableDetection() {
	fmt.Println("Listening for ARP changes...")
	entries := getCurrentEntries()
	for {
		currentEntries := getCurrentEntries()
		detectChanges(entries, currentEntries)
		entries = currentEntries
		time.Sleep(1000 * time.Millisecond)
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
			if matchedEntry.MacAddress != entry.MacAddress {
				tellTheUser(entry, matchedEntry)
			}
		}
	}
}

func tellTheUser(entry *ArpEntry, matchedEntry *ArpEntry) {
	fmt.Println("Mac address change detected for same IP Address")
	fmt.Printf("IP[%s] - %s => %s\n", matchedEntry.IpAddress, entry.MacAddress, matchedEntry.MacAddress)
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

func regSplit(text string, delimeter string) []string {
	    reg := regexp.MustCompile(delimeter)
	    indexes := reg.FindAllStringIndex(text, -1)
	    laststart := 0
	    result := make([]string, len(indexes) + 1)
	    for i, element := range indexes {
		result[i] = text[laststart:element[0]]
		laststart = element[1]
	    }
	    result[len(indexes)] = text[laststart:len(text)]
	    return result
}

