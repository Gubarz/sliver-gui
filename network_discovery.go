package main

import (
	"context"
	"fmt"
	"net/netip"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/bishopfox/sliver/protobuf/clientpb"
	"github.com/kballard/go-shellquote"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

const (
	discoveryTimeout = 10 * time.Minute
	maxSweepHosts    = 256
)

var (
	ipPattern          = regexp.MustCompile(`\b(?:\d{1,3}\.){3}\d{1,3}\b`)
	macPattern         = regexp.MustCompile(`(?i)\b[0-9a-f]{2}(?:[:-][0-9a-f]{2}){5}\b`)
	markerPattern      = regexp.MustCompile(`(?m)^DISCOVERY\|([^\s|]+)\|([^\r\n|]*)\|([0-9]*)`)
	arpParenPattern    = regexp.MustCompile(`\(((?:\d{1,3}\.){3}\d{1,3})\)`)
	windowsPingPattern = regexp.MustCompile(`(?im)^\s*Reply from\s+((?:\d{1,3}\.){3}\d{1,3})\s*:.*?\bTTL[= ](\d+)\b`)
	unixPingPattern    = regexp.MustCompile(`(?im)^\s*\d+\s+bytes from\s+(?:[^\s(]+\s+\()?((?:\d{1,3}\.){3}\d{1,3})\)?[:\s].*?\bttl[= ](\d+)\b`)
)

type NetworkDiscovery struct {
	AgentID  string `json:"agentID"`
	IP       string `json:"ip"`
	MAC      string `json:"mac"`
	Hostname string `json:"hostname"`
	Vendor   string `json:"vendor"`
	OSHint   string `json:"osHint"`
	TTL      int    `json:"ttl"`
	Method   string `json:"method"`
	LastSeen int64  `json:"lastSeen"`
}

func (a *App) GetNetworkDiscoveries() []NetworkDiscovery {
	a.discoveryMu.RLock()
	defer a.discoveryMu.RUnlock()

	var devices []NetworkDiscovery
	for _, byIP := range a.discoveries {
		for _, device := range byIP {
			devices = append(devices, device)
		}
	}
	sort.Slice(devices, func(i, j int) bool {
		if devices[i].AgentID == devices[j].AgentID {
			return devices[i].IP < devices[j].IP
		}
		return devices[i].AgentID < devices[j].AgentID
	})
	return devices
}

func (a *App) DiscoverNetwork(agentID, method, cidr string) ([]NetworkDiscovery, error) {
	if !a.connected {
		return nil, fmt.Errorf("not connected")
	}
	session, beacon, err := a.findTarget(agentID)
	if err != nil {
		return nil, err
	}

	osName := targetOS(session, beacon)
	method = strings.ToLower(strings.TrimSpace(method))
	var commands []string
	switch method {
	case "arp":
		commands = []string{neighborCommand(osName)}
	case "sweep":
		hosts, err := sweepHosts(cidr)
		if err != nil {
			return nil, err
		}
		commands = []string{sweepCommand(osName, hosts), neighborCommand(osName)}
	default:
		return nil, fmt.Errorf("unsupported discovery method: %s", method)
	}

	ctx, cancel := context.WithTimeout(context.Background(), discoveryTimeout)
	defer cancel()

	var found []NetworkDiscovery
	for _, command := range commands {
		output, taskID, err := a.runAutomationConsoleLine(agentID, command)
		if err != nil {
			return nil, err
		}
		if beacon != nil {
			output, _, err = a.awaitBeaconTask(ctx, agentID, output, taskID)
			if err != nil {
				return nil, err
			}
		}
		found = mergeDiscoveryResults(found, parseDiscoveryOutput(agentID, method, output))
	}

	a.storeDiscoveries(found)
	return a.GetNetworkDiscoveries(), nil
}

func (a *App) ClearNetworkDiscoveries(agentID string) {
	a.discoveryMu.Lock()
	if strings.TrimSpace(agentID) == "" {
		clear(a.discoveries)
	} else {
		delete(a.discoveries, agentID)
	}
	a.discoveryMu.Unlock()
	a.emitDiscoveryUpdate()
}

func (a *App) RemoveNetworkDiscoveries(agentID string, ips []string) {
	a.discoveryMu.Lock()
	if byIP := a.discoveries[agentID]; byIP != nil {
		for _, ip := range ips {
			delete(byIP, strings.TrimSpace(ip))
		}
		if len(byIP) == 0 {
			delete(a.discoveries, agentID)
		}
	}
	a.discoveryMu.Unlock()
	a.emitDiscoveryUpdate()
}

func targetOS(session *clientpb.Session, beacon *clientpb.Beacon) string {
	if session != nil {
		return strings.ToLower(session.OS)
	}
	if beacon != nil {
		return strings.ToLower(beacon.OS)
	}
	return ""
}

func neighborCommand(osName string) string {
	switch osName {
	case "windows":
		return shellquote.Join("execute", "-o", "--", "arp.exe", "-a")
	case "darwin":
		return shellquote.Join("execute", "-o", "--", "/usr/sbin/arp", "-an")
	default:
		script := `ip neigh 2>/dev/null || arp -an 2>/dev/null`
		return shellquote.Join("execute", "-o", "--", "/bin/sh", "-c", script)
	}
}

func sweepCommand(osName string, hosts []string) string {
	switch osName {
	case "windows":
		quoted := make([]string, 0, len(hosts))
		for _, host := range hosts {
			quoted = append(quoted, "'"+host+"'")
		}
		script := fmt.Sprintf(
			`$p=New-Object System.Net.NetworkInformation.Ping; %s | ForEach-Object { try { $r=$p.Send($_,500); if ($r.Status -eq 'Success') { Write-Output ('DISCOVERY|' + $_ + '||' + $r.Options.Ttl) } } catch {} }`,
			strings.Join(quoted, ","),
		)
		return shellquote.Join(
			"execute", "-o", "--", "powershell.exe",
			"-NoProfile", "-NonInteractive", "-Command", script,
		)
	case "darwin":
		return shellquote.Join("execute", "-o", "--", "/bin/sh", "-c", unixSweepScript(hosts, "-c 1 -W 1000"))
	default:
		return shellquote.Join("execute", "-o", "--", "/bin/sh", "-c", unixSweepScript(hosts, "-c 1 -W 1"))
	}
}

func unixSweepScript(hosts []string, pingArgs string) string {
	return fmt.Sprintf(
		`i=0; for ip in %s; do (out=$(ping %s "$ip" 2>/dev/null) && ttl=$(printf '%%s\n' "$out" | sed -n 's/.*ttl[= ]\([0-9][0-9]*\).*/\1/p' | head -n 1) && printf 'DISCOVERY|%%s||%%s\n' "$ip" "$ttl") & i=$((i+1)); if [ $((i %% 32)) -eq 0 ]; then wait; fi; done; wait`,
		strings.Join(hosts, " "),
		pingArgs,
	)
}

func sweepHosts(value string) ([]string, error) {
	prefix, err := netip.ParsePrefix(strings.TrimSpace(value))
	if err != nil || !prefix.Addr().Is4() {
		return nil, fmt.Errorf("enter a valid IPv4 CIDR, for example 192.168.1.0/24")
	}
	prefix = prefix.Masked()
	if prefix.Bits() < 24 {
		return nil, fmt.Errorf("sweeps are limited to /24 networks or smaller")
	}

	var hosts []string
	for address := prefix.Addr(); prefix.Contains(address); address = address.Next() {
		hosts = append(hosts, address.String())
		if len(hosts) > maxSweepHosts {
			return nil, fmt.Errorf("sweep exceeds the %d host limit", maxSweepHosts)
		}
	}
	if len(hosts) > 2 {
		hosts = hosts[1 : len(hosts)-1]
	}
	return hosts, nil
}

func parseDiscoveryOutput(agentID, method, output string) []NetworkDiscovery {
	now := time.Now().UnixMilli()
	byIP := make(map[string]NetworkDiscovery)

	for _, match := range markerPattern.FindAllStringSubmatch(output, -1) {
		if ip, err := netip.ParseAddr(match[1]); err == nil && discoveryHostIP(ip) {
			ttl := parseTTL(match[3])
			byIP[ip.String()] = NetworkDiscovery{
				AgentID: agentID, IP: ip.String(), Hostname: strings.TrimSpace(match[2]),
				OSHint: osHintFromTTL(ttl), TTL: ttl, Method: method, LastSeen: now,
			}
		}
	}

	for _, line := range strings.Split(output, "\n") {
		ipText := ""
		parenMatch := arpParenPattern.FindStringSubmatch(line)
		mac := macPattern.FindString(line)
		if len(parenMatch) == 2 {
			if mac == "" {
				continue
			}
			ipText = parenMatch[1]
		} else if mac != "" {
			match := ipPattern.FindString(line)
			ipText = match
		}
		ip, err := netip.ParseAddr(ipText)
		if err != nil || !discoveryHostIP(ip) {
			continue
		}
		device := byIP[ip.String()]
		device.AgentID = agentID
		device.IP = ip.String()
		device.Method = method
		device.LastSeen = now
		if mac != "" {
			device.MAC = normalizeMAC(mac)
			if !hostMAC(device.MAC) {
				continue
			}
			device.Vendor = vendorFromMAC(device.MAC)
		}
		byIP[device.IP] = device
	}

	devices := make([]NetworkDiscovery, 0, len(byIP))
	for _, device := range byIP {
		devices = append(devices, device)
	}
	return devices
}

func parsePingDiscoveryOutput(agentID, output string) []NetworkDiscovery {
	now := time.Now().UnixMilli()
	byIP := make(map[string]NetworkDiscovery)
	for _, pattern := range []*regexp.Regexp{windowsPingPattern, unixPingPattern} {
		for _, match := range pattern.FindAllStringSubmatch(output, -1) {
			ip, err := netip.ParseAddr(match[1])
			if err != nil || !discoveryHostIP(ip) {
				continue
			}
			ttl := parseTTL(match[2])
			byIP[ip.String()] = NetworkDiscovery{
				AgentID:  agentID,
				IP:       ip.String(),
				OSHint:   osHintFromTTL(ttl),
				TTL:      ttl,
				Method:   "ping",
				LastSeen: now,
			}
		}
	}

	devices := make([]NetworkDiscovery, 0, len(byIP))
	for _, device := range byIP {
		devices = append(devices, device)
	}
	return devices
}

func mergeDiscoveryResults(current, next []NetworkDiscovery) []NetworkDiscovery {
	byKey := make(map[string]NetworkDiscovery, len(current)+len(next))
	for _, device := range append(current, next...) {
		key := device.AgentID + "|" + device.IP
		existing := byKey[key]
		if device.MAC == "" {
			device.MAC = existing.MAC
		}
		if device.Hostname == "" {
			device.Hostname = existing.Hostname
		}
		if device.Vendor == "" {
			device.Vendor = existing.Vendor
		}
		if device.OSHint == "" {
			device.OSHint = existing.OSHint
			device.TTL = existing.TTL
		}
		byKey[key] = device
	}
	merged := make([]NetworkDiscovery, 0, len(byKey))
	for _, device := range byKey {
		merged = append(merged, device)
	}
	return merged
}

func (a *App) storeDiscoveries(devices []NetworkDiscovery) {
	a.discoveryMu.Lock()
	for _, device := range devices {
		if a.discoveries[device.AgentID] == nil {
			a.discoveries[device.AgentID] = make(map[string]NetworkDiscovery)
		}
		existing := a.discoveries[device.AgentID][device.IP]
		if device.MAC == "" {
			device.MAC = existing.MAC
		}
		if device.Hostname == "" {
			device.Hostname = existing.Hostname
		}
		if device.Vendor == "" {
			device.Vendor = existing.Vendor
		}
		if device.OSHint == "" {
			device.OSHint = existing.OSHint
			device.TTL = existing.TTL
		}
		a.discoveries[device.AgentID][device.IP] = device
	}
	a.discoveryMu.Unlock()
	a.emitDiscoveryUpdate()
}

func parseTTL(value string) int {
	var ttl int
	_, _ = fmt.Sscanf(strings.TrimSpace(value), "%d", &ttl)
	return ttl
}

func discoveryHostIP(ip netip.Addr) bool {
	return ip.Is4() &&
		!ip.IsUnspecified() &&
		!ip.IsMulticast() &&
		ip.String() != "255.255.255.255"
}

func osHintFromTTL(ttl int) string {
	switch {
	case ttl <= 0:
		return ""
	case ttl <= 64:
		return "Unix-like"
	case ttl <= 128:
		return "Windows-like"
	default:
		return "Network appliance"
	}
}

func normalizeMAC(value string) string {
	return strings.ToLower(strings.ReplaceAll(value, "-", ":"))
}

func hostMAC(value string) bool {
	parts := strings.Split(value, ":")
	if len(parts) != 6 || value == "ff:ff:ff:ff:ff:ff" {
		return false
	}
	var first uint
	if _, err := fmt.Sscanf(parts[0], "%02x", &first); err != nil {
		return false
	}
	return first&1 == 0
}

func vendorFromMAC(value string) string {
	return lookupOUI(value)
}

func (a *App) emitDiscoveryUpdate() {
	if a.ctx != nil {
		runtime.EventsEmit(a.ctx, "network-discovery-updated", a.GetNetworkDiscoveries())
	}
}
