##ARP WATCH
ARP Watch is a Mac OSX tool for detecting changes in mac addresses. A common attack that takes place is known as an ARP cache poisoning attack. This attack tricks your machine into thinking that it is talking to the router, then it is really talking to a malicious third party. 

###How does it work?

ARP Watch parses the output of the linx `arp` command. It uses this to build an in memory model of the current ARP entries, and every second will check if the IP has remained the same but the MAC address has changed. While this isn't a guarantee of malicious activity, it is one of the classic symptoms.

###Usage

####Using the Dist for OSX

In the Dist folder is a runnable script for OSX.

####Rebuilding for Linux

Clone the contents of this repository. Navigate to the route directory and run

    go install

*NOTE*: You will need to have Go setup on your machine.

###Contributing

Dive right in! It's very primitive at the moment so any improvements or suggestions are more than welcome.
