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
	"./arpwatch_unix"
)

type ArpEntry struct {
	IpAddress  string
	MacAddress string
}

//general var initialization
var version string = "1.0.4"
var fileName string
var timeStamp, ipAddress, oldMac, newMac string

//var flag initialization
var outFile = flag.String("outfile", "", "file to write logs to")
var quiet = flag.Bool("quiet", false, "supress output")
var RlogServer string
var versionFlag = flag.Bool("version", false, "print version")


func init() {
	flag.StringVar(outFile, "o", "", "file to write logs to")
	flag.BoolVar(quiet, "q", false, "supress output")
	flag.BoolVar(versionFlag, "v", false, "print version")
}


func main() {
	flag.Parse()
	if *versionFlag == true {
		fmt.Println("arpwatch-go", version)
		os.Exit(0)
	}
	if *quiet != true {
		fmt.Println("Listening for ARP changes...")
	}
	entries := getCurrentEntries()
	for {
		currentEntries := getCurrentEntries()
		detectAndAlertChanges(entries, currentEntries)
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

func detectAndAlertChanges(oldEntries []*ArpEntry, newEntries []*ArpEntry) {
	if oldEntries == nil {
		return
	}

	for _, entry := range oldEntries {
		matchedEntry := getMatchingEntry(entry, newEntries)

		if matchedEntry != nil {
			var entryChange bool = entry.MacAddress != matchedEntry.MacAddress && matchedEntry.MacAddress != "(incomplete)" && entry.MacAddress != "(incomplete)"
			if entryChange {
				t := time.Now()
				changeTime := t.Format(time.RFC3339)
				tellTheUser(entry, matchedEntry, changeTime)
			}
		}
	}
}

func tellTheUser(entry *ArpEntry, matchedEntry *ArpEntry, timeValue string) {
	if *quiet != true {
		printToStdout(matchedEntry.IpAddress, entry.MacAddress, matchedEntry.MacAddress)
	}

	text := "timeStamp=" + timeValue + " ip=" + matchedEntry.IpAddress + " oldMac=" + entry.MacAddress + " newMac=" + matchedEntry.MacAddress + " Message='MAC address change detected'"
	if *outFile != "" {
		logToFile(text, *outFile)
	}

	if arpwatch_unix.RlogServer != "" {
		arpwatch_unix.LogToRemote(text, arpwatch_unix.RlogServer)
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



func printToStdout(ipAddress string, oldMac string, newMac string) {
	fmt.Println("Mac address change detected for same IP Address")
	fmt.Printf("IP[%s] - %s => %s\n", ipAddress, oldMac, newMac)
}

func logToFile(message string, fileName string) {
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		var file, err = os.Create(fileName)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()
	}
	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		panic(err)
	}
	logger := log.New(file, "", log.LstdFlags|log.Lshortfile)

	logger.Println(message)
}
