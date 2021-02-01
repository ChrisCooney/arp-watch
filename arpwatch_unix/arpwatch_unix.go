// +build !windows

package arpwatch_unix


import (
	"log"
	"flag"
	"log/syslog"
)

var RlogServer string

func init() {
	flag.StringVar(&RlogServer, "server", "", "remote server to log to (UDP)")
}


func LogToRemote(message string, rServer string) {
      logWriter, err := syslog.Dial("udp", rServer, syslog.LOG_ERR, "arpwatch")
      defer logWriter.Close()
      if err != nil {
              log.Fatal(err)
      }

      logWriter.Write([]byte(message))
}
