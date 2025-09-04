package stream

import (
	"fmt"
	"time"
)

// ============================================================================
// NETWORK DATA STRUCTURES
// ============================================================================

// NetFlow represents a network flow record optimized for processing
type NetFlow struct {
	// Core flow identifiers
	SrcIP    uint32 // Source IP address
	DstIP    uint32 // Destination IP address  
	SrcPort  uint16 // Source port
	DstPort  uint16 // Destination port
	Protocol uint8  // IP protocol (TCP=6, UDP=17, etc.)
	
	// Timing information
	StartTime time.Time     // Flow start time
	Duration  time.Duration // Flow duration
	
	// Traffic counters
	Bytes   uint64 // Total bytes
	Packets uint64 // Total packets
	
	// Additional metadata
	TCPFlags uint8  // TCP flags (SYN, ACK, FIN, etc.)
	ToS      uint8  // Type of Service
	ASN      uint16 // AS number
}

// FlowKey represents the unique identifier for a flow
type FlowKey struct {
	SrcIP    uint32
	DstIP    uint32
	SrcPort  uint16
	DstPort  uint16
	Protocol uint8
}

// Key generates a FlowKey from a NetFlow
func (nf *NetFlow) Key() FlowKey {
	return FlowKey{
		SrcIP:    nf.SrcIP,
		DstIP:    nf.DstIP,
		SrcPort:  nf.SrcPort,
		DstPort:  nf.DstPort,
		Protocol: nf.Protocol,
	}
}

// String returns human-readable flow representation
func (nf *NetFlow) String() string {
	return fmt.Sprintf("%s:%d -> %s:%d (%s) %d bytes/%d packets",
		intToIP(nf.SrcIP), nf.SrcPort,
		intToIP(nf.DstIP), nf.DstPort,
		protocolToString(nf.Protocol),
		nf.Bytes, nf.Packets)
}

// ToRecord converts NetFlow to Record for generic processing
func (nf *NetFlow) ToRecord() Record {
	return R(
		"src_ip", nf.SrcIP,
		"dst_ip", nf.DstIP,
		"src_port", nf.SrcPort,
		"dst_port", nf.DstPort,
		"protocol", nf.Protocol,
		"bytes", nf.Bytes,
		"packets", nf.Packets,
		"start_time", nf.StartTime,
		"duration", nf.Duration,
		"tcp_flags", nf.TCPFlags,
		"tos", nf.ToS,
		"asn", nf.ASN,
	)
}

// FromRecord creates NetFlow from Record
func NetFlowFromRecord(r Record) NetFlow {
	return NetFlow{
		SrcIP:     GetOr(r, "src_ip", uint32(0)),
		DstIP:     GetOr(r, "dst_ip", uint32(0)),
		SrcPort:   GetOr(r, "src_port", uint16(0)),
		DstPort:   GetOr(r, "dst_port", uint16(0)),
		Protocol:  GetOr(r, "protocol", uint8(0)),
		Bytes:     GetOr(r, "bytes", uint64(0)),
		Packets:   GetOr(r, "packets", uint64(0)),
		StartTime: GetOr(r, "start_time", time.Time{}),
		Duration:  GetOr(r, "duration", time.Duration(0)),
		TCPFlags:  GetOr(r, "tcp_flags", uint8(0)),
		ToS:       GetOr(r, "tos", uint8(0)),
		ASN:       GetOr(r, "asn", uint16(0)),
	}
}

// ============================================================================
// NETWORK ANALYTICS RESULTS
// ============================================================================

// TopTalker represents a high-traffic IP address
type TopTalker struct {
	IP    uint32 // IP address as uint32
	Bytes uint64 // Total bytes
}

// String returns human-readable representation
func (tt TopTalker) String() string {
	return fmt.Sprintf("%s: %d bytes", intToIP(tt.IP), tt.Bytes)
}

// PortScanAlert represents a detected port scan
type PortScanAlert struct {
	SrcIP     uint32        // Source IP performing the scan
	PortCount int           // Number of unique ports scanned
	Timestamp time.Time     // When the scan started
	Duration  time.Duration // How long the scan lasted
}

// String returns human-readable representation
func (psa PortScanAlert) String() string {
	return fmt.Sprintf("Port scan from %s: %d ports scanned in %v",
		intToIP(psa.SrcIP), psa.PortCount, psa.Duration)
}

// DDoSAlert represents a detected DDoS attack
type DDoSAlert struct {
	DstIP     uint32        // Destination IP under attack
	PPS       uint64        // Packets per second
	BPS       uint64        // Bytes per second  
	Severity  float64       // Attack severity score
	Duration  time.Duration // Attack duration
	FlowCount int           // Number of flows involved
}

// String returns human-readable representation
func (da DDoSAlert) String() string {
	return fmt.Sprintf("DDoS attack on %s: %.0f PPS, severity %.2f",
		intToIP(da.DstIP), float64(da.PPS), da.Severity)
}

// TrafficMatrix represents network traffic flows between sources/destinations
type TrafficMatrix struct {
	Dimension int                    // Matrix size (N x N)
	Sources   []uint32              // Source IP addresses
	Flows     map[string]FlowSummary // [srcIP,dstIP] -> summary
}

// FlowSummary summarizes traffic between two endpoints
type FlowSummary struct {
	Bytes      uint64        // Total bytes
	Packets    uint64        // Total packets
	Flows      int           // Number of flows
	FirstSeen  time.Time     // First flow timestamp
	LastSeen   time.Time     // Last flow timestamp
	Protocols  map[uint8]int // Protocol distribution
}

// ============================================================================
// NETWORK STREAM OPERATIONS
// ============================================================================

// NetFlowToRecords converts NetFlow stream to Record stream
func NetFlowToRecords() Filter[NetFlow, Record] {
	return Map(func(flow NetFlow) Record {
		return flow.ToRecord()
	})
}

// RecordsToNetFlow converts Record stream to NetFlow stream
func RecordsToNetFlow() Filter[Record, NetFlow] {
	return Map(func(r Record) NetFlow {
		return NetFlowFromRecord(r)
	})
}

// FilterByProtocol filters flows by IP protocol
func FilterByProtocol(protocol uint8) Filter[NetFlow, NetFlow] {
	return Where(func(flow NetFlow) bool {
		return flow.Protocol == protocol
	})
}

// FilterByPort filters flows by source or destination port
func FilterByPort(port uint16) Filter[NetFlow, NetFlow] {
	return Where(func(flow NetFlow) bool {
		return flow.SrcPort == port || flow.DstPort == port
	})
}

// FilterByIPRange filters flows by IP address range
func FilterByIPRange(startIP, endIP uint32) Filter[NetFlow, NetFlow] {
	return Where(func(flow NetFlow) bool {
		return (flow.SrcIP >= startIP && flow.SrcIP <= endIP) ||
		       (flow.DstIP >= startIP && flow.DstIP <= endIP)
	})
}

// GroupByEndpoints groups flows by source and destination IP
func GroupByEndpoints() Filter[NetFlow, Record] {
	return func(input Stream[NetFlow]) Stream[Record] {
		// Collect all flows first
		flows, err := Collect(input)
		if err != nil {
			return func() (Record, error) { return Record{}, err }
		}
		
		// Group by source-destination pairs
		groups := make(map[string][]NetFlow)
		for _, flow := range flows {
			key := fmt.Sprintf("%d-%d", flow.SrcIP, flow.DstIP)
			groups[key] = append(groups[key], flow)
		}
		
		// Convert to records
		var results []Record
		for _, groupFlows := range groups {
			if len(groupFlows) == 0 {
				continue
			}
			
			first := groupFlows[0]
			
			var totalBytes, totalPackets uint64
			protocols := make(map[uint8]int)
			
			for _, flow := range groupFlows {
				totalBytes += flow.Bytes
				totalPackets += flow.Packets
				protocols[flow.Protocol]++
			}
			
			result := R(
				"src_ip", first.SrcIP,
				"dst_ip", first.DstIP,
				"total_bytes", totalBytes,
				"total_packets", totalPackets,
				"flow_count", len(groupFlows),
				"protocols", protocols,
			)
			
			results = append(results, result)
		}
		
		return FromSlice(results)
	}
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

// intToIP converts uint32 to IP string
func intToIP(ip uint32) string {
	return fmt.Sprintf("%d.%d.%d.%d",
		byte(ip>>24), byte(ip>>16), byte(ip>>8), byte(ip))
}

// protocolToString converts protocol number to string
func protocolToString(proto uint8) string {
	switch proto {
	case 1:
		return "ICMP"
	case 6:
		return "TCP"
	case 17:
		return "UDP"
	default:
		return fmt.Sprintf("Proto-%d", proto)
	}
}

// ============================================================================
// NETWORK STREAM CONSTRUCTORS
// ============================================================================

// FromNetFlows creates a stream from NetFlow slice
func FromNetFlows(flows []NetFlow) Stream[NetFlow] {
	return FromSlice(flows)
}

// GenerateTestFlows creates synthetic network flows for testing
func GenerateTestFlows(count int) []NetFlow {
	flows := make([]NetFlow, count)
	
	// Create some "heavy hitters" for realistic distribution
	heavyHitters := []uint32{
		0xC0A80101, // 192.168.1.1
		0xC0A80102, // 192.168.1.2
		0xC0A80103, // 192.168.1.3
	}
	
	for i := 0; i < count; i++ {
		var srcIP, dstIP uint32
		
		// 30% of traffic from heavy hitters
		if i%3 == 0 {
			srcIP = heavyHitters[i%len(heavyHitters)]
			dstIP = uint32(0xC0A80000 + (i % 256))
		} else {
			srcIP = uint32(0xC0A80000 + (i % 256))
			dstIP = uint32(0xC0A80100 + (i % 256))
		}
		
		flows[i] = NetFlow{
			SrcIP:     srcIP,
			DstIP:     dstIP,
			SrcPort:   uint16(32768 + (i % 32768)),
			DstPort:   uint16(80 + (i % 1000)),
			Protocol:  6, // TCP
			Bytes:     uint64(64 + (i % 1400)),
			Packets:   uint64(1 + (i % 10)),
			StartTime: time.Now().Add(-time.Duration(i) * time.Second),
		}
	}
	
	return flows
}