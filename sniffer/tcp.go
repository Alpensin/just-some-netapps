// Package sniffer sniffs shit
package sniffer

import (
	"log"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
)

func handlePacket(packet gopacket.Packet) {
	ipl := packet.NetworkLayer()
	neFlow := ipl.NetworkFlow()
	tl := packet.TransportLayer()
	trFlow := tl.TransportFlow()
	if len(tl.LayerPayload()) > 0 {
		log.Printf("Message from %s:%v to %s:%v: %s",
			neFlow.Src(), trFlow.Src(),
			neFlow.Dst(), trFlow.Dst(),
			string(tl.LayerPayload()))
	}
}

func getPacketSource() *gopacket.PacketSource {
	var err error

	var handle *pcap.Handle
	if handle, err = pcap.OpenLive("lo", 1600, true, pcap.BlockForever); err != nil {
		panic(err)
	}

	if err = handle.SetBPFFilter("tcp and port 6969"); err != nil {
		panic(err)
	}

	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	return packetSource
}

// Sniff - sniff port 6969 of localhost
func Sniff() {
	for {
		packetSource := getPacketSource()
		for packet := range packetSource.Packets() {
			handlePacket(packet)
		}
	}
}
