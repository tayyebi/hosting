package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// ════════════════════════════════════════════════════════════════
// Workspace Networking API — Private Connectivity: VPN Gateway
// networking/private/vpn
// ════════════════════════════════════════════════════════════════

type VPNGateway struct {
	ID         string
	Name       string
	Type       string
	Status     string
	Region     string
	PublicIP   string
	RemoteIP   string
	Tunnels    int
	Throughput string
	Encryption string
	IKEVersion string
	Uptime     string
}

type vpnGatewayStoreType struct {
	mu       sync.RWMutex
	gateways map[string][]VPNGateway
}

var vpnGatewayStore = &vpnGatewayStoreType{
	gateways: map[string][]VPNGateway{},
}

func (s *vpnGatewayStoreType) list(slug string) []VPNGateway {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.gateways[slug]
	out := make([]VPNGateway, len(src))
	copy(out, src)
	return out
}

func (s *vpnGatewayStoreType) getByID(slug, id string) (VPNGateway, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, g := range s.gateways[slug] {
		if g.ID == id {
			return g, true
		}
	}
	return VPNGateway{}, false
}

func (s *vpnGatewayStoreType) add(slug string, g VPNGateway) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.gateways[slug] = append(s.gateways[slug], g)
}

func (s *vpnGatewayStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.gateways[slug]
	for i, g := range list {
		if g.ID == id {
			s.gateways[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

type apiVPNGateway struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Type       string `json:"type"`
	Status     string `json:"status"`
	Region     string `json:"region"`
	PublicIP   string `json:"public_ip"`
	RemoteIP   string `json:"remote_ip"`
	Tunnels    int    `json:"tunnels"`
	Throughput string `json:"throughput"`
	Encryption string `json:"encryption"`
	IKEVersion string `json:"ike_version"`
	Uptime     string `json:"uptime"`
}

type apiVPNGatewayList struct {
	Gateways []apiVPNGateway `json:"gateways"`
	Total    int             `json:"total"`
}

func toAPIVPNGateway(g VPNGateway) apiVPNGateway {
	return apiVPNGateway{
		ID: g.ID, Name: g.Name, Type: g.Type,
		Status: g.Status, Region: g.Region,
		PublicIP: g.PublicIP, RemoteIP: g.RemoteIP,
		Tunnels: g.Tunnels, Throughput: g.Throughput,
		Encryption: g.Encryption, IKEVersion: g.IKEVersion,
		Uptime: g.Uptime,
	}
}

func filterVPNGateway(all []VPNGateway, search, status string) []VPNGateway {
	var out []VPNGateway
	q := strings.ToLower(search)
	for _, g := range all {
		if status != "" && g.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(g.Name + " " + g.Type + " " + g.Region + " " + g.PublicIP)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, g)
	}
	return out
}

func apiWSNetPrivVPNListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	gws := filterVPNGateway(vpnGatewayStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiVPNGateway, len(gws))
	for i, g := range gws {
		out[i] = toAPIVPNGateway(g)
	}
	apiOK(w, apiVPNGatewayList{Gateways: out, Total: len(out)})
}

type apiCreateVPNGatewayRequest struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	Region     string `json:"region"`
	RemoteIP   string `json:"remote_ip"`
	Encryption string `json:"encryption"`
	IKEVersion string `json:"ike_version"`
}

func apiWSNetPrivVPNCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateVPNGatewayRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		apiError(w, http.StatusBadRequest, "Name is required.")
		return
	}
	typ := strings.TrimSpace(req.Type)
	if typ == "" {
		typ = "site-to-site"
	}
	region := strings.TrimSpace(req.Region)
	if region == "" {
		region = "eu-central-1"
	}
	encryption := strings.TrimSpace(req.Encryption)
	if encryption == "" {
		encryption = "AES-256-GCM"
	}
	ikeVer := strings.TrimSpace(req.IKEVersion)
	if ikeVer == "" {
		ikeVer = "IKEv2"
	}
	remoteIP := strings.TrimSpace(req.RemoteIP)
	if remoteIP == "" {
		remoteIP = "—"
	}
	gw := VPNGateway{
		ID: newID(), Name: name, Type: typ,
		Status: "provisioning", Region: region,
		PublicIP: "pending", RemoteIP: remoteIP,
		Tunnels: 0, Throughput: "0 Mbps",
		Encryption: encryption, IKEVersion: ikeVer,
		Uptime: "—",
	}
	vpnGatewayStore.add(slug, gw)
	apiCreated(w, toAPIVPNGateway(gw))
}

func apiWSNetPrivVPNGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	gw, found := vpnGatewayStore.getByID(slug, r.PathValue("gid"))
	if !found {
		apiError(w, http.StatusNotFound, "VPN gateway not found.")
		return
	}
	apiOK(w, toAPIVPNGateway(gw))
}

func apiWSNetPrivVPNDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := vpnGatewayStore.getByID(slug, r.PathValue("gid")); !found {
		apiError(w, http.StatusNotFound, "VPN gateway not found.")
		return
	}
	vpnGatewayStore.delete(slug, r.PathValue("gid"))
	apiOK(w, map[string]any{"ok": true, "message": "VPN gateway deleted."})
}

// ════════════════════════════════════════════════════════════════
// Workspace Networking API — Private Connectivity: Direct Connect
// networking/private/direct
// ════════════════════════════════════════════════════════════════

type DirectConnect struct {
	ID         string
	Name       string
	Status     string
	Region     string
	Location   string
	Port       string
	Bandwidth  string
	VLAN       int
	BGP_ASN    string
	Redundancy string
	Latency    string
	MACsec     bool
}

type directConnectStoreType struct {
	mu          sync.RWMutex
	connections map[string][]DirectConnect
}

var directConnectStore = &directConnectStoreType{
	connections: map[string][]DirectConnect{},
}

func (s *directConnectStoreType) list(slug string) []DirectConnect {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.connections[slug]
	out := make([]DirectConnect, len(src))
	copy(out, src)
	return out
}

func (s *directConnectStoreType) getByID(slug, id string) (DirectConnect, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, c := range s.connections[slug] {
		if c.ID == id {
			return c, true
		}
	}
	return DirectConnect{}, false
}

func (s *directConnectStoreType) add(slug string, c DirectConnect) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.connections[slug] = append(s.connections[slug], c)
}

func (s *directConnectStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.connections[slug]
	for i, c := range list {
		if c.ID == id {
			s.connections[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

type apiDirectConnect struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Status     string `json:"status"`
	Region     string `json:"region"`
	Location   string `json:"location"`
	Port       string `json:"port"`
	Bandwidth  string `json:"bandwidth"`
	VLAN       int    `json:"vlan"`
	BGP_ASN    string `json:"bgp_asn"`
	Redundancy string `json:"redundancy"`
	Latency    string `json:"latency"`
	MACsec     bool   `json:"macsec"`
}

type apiDirectConnectList struct {
	Connections []apiDirectConnect `json:"connections"`
	Total       int                `json:"total"`
}

func toAPIDirectConnect(c DirectConnect) apiDirectConnect {
	return apiDirectConnect{
		ID: c.ID, Name: c.Name, Status: c.Status,
		Region: c.Region, Location: c.Location,
		Port: c.Port, Bandwidth: c.Bandwidth,
		VLAN: c.VLAN, BGP_ASN: c.BGP_ASN,
		Redundancy: c.Redundancy, Latency: c.Latency,
		MACsec: c.MACsec,
	}
}

func filterDirectConnect(all []DirectConnect, search, status string) []DirectConnect {
	var out []DirectConnect
	q := strings.ToLower(search)
	for _, c := range all {
		if status != "" && c.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(c.Name + " " + c.Region + " " + c.Location + " " + c.Port)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, c)
	}
	return out
}

func apiWSNetPrivDirectListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	conns := filterDirectConnect(directConnectStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiDirectConnect, len(conns))
	for i, c := range conns {
		out[i] = toAPIDirectConnect(c)
	}
	apiOK(w, apiDirectConnectList{Connections: out, Total: len(out)})
}

type apiCreateDirectConnectRequest struct {
	Name       string `json:"name"`
	Region     string `json:"region"`
	Location   string `json:"location"`
	Port       string `json:"port"`
	Bandwidth  string `json:"bandwidth"`
	BGP_ASN    string `json:"bgp_asn"`
	Redundancy string `json:"redundancy"`
	MACsec     bool   `json:"macsec"`
}

func apiWSNetPrivDirectCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateDirectConnectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		apiError(w, http.StatusBadRequest, "Name is required.")
		return
	}
	location := strings.TrimSpace(req.Location)
	if location == "" {
		apiError(w, http.StatusBadRequest, "Location is required.")
		return
	}
	region := strings.TrimSpace(req.Region)
	if region == "" {
		region = "eu-central-1"
	}
	port := strings.TrimSpace(req.Port)
	if port == "" {
		port = "1 GbE"
	}
	bandwidth := strings.TrimSpace(req.Bandwidth)
	if bandwidth == "" {
		bandwidth = "1 Gbps"
	}
	bgpASN := strings.TrimSpace(req.BGP_ASN)
	if bgpASN == "" {
		bgpASN = "65000"
	}
	redundancy := strings.TrimSpace(req.Redundancy)
	if redundancy == "" {
		redundancy = "single"
	}
	conn := DirectConnect{
		ID: newID(), Name: name, Status: "pending-approval",
		Region: region, Location: location,
		Port: port, Bandwidth: bandwidth,
		VLAN: 0, BGP_ASN: bgpASN,
		Redundancy: redundancy, Latency: "—",
		MACsec: req.MACsec,
	}
	directConnectStore.add(slug, conn)
	apiCreated(w, toAPIDirectConnect(conn))
}

func apiWSNetPrivDirectGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	conn, found := directConnectStore.getByID(slug, r.PathValue("cid"))
	if !found {
		apiError(w, http.StatusNotFound, "Direct connect not found.")
		return
	}
	apiOK(w, toAPIDirectConnect(conn))
}

func apiWSNetPrivDirectDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := directConnectStore.getByID(slug, r.PathValue("cid")); !found {
		apiError(w, http.StatusNotFound, "Direct connect not found.")
		return
	}
	directConnectStore.delete(slug, r.PathValue("cid"))
	apiOK(w, map[string]any{"ok": true, "message": "Direct connect deleted."})
}

// ════════════════════════════════════════════════════════════════
// Workspace Networking API — Private Connectivity: SD-WAN
// networking/private/sdwan
// ════════════════════════════════════════════════════════════════

type SDWAN struct {
	ID         string
	Name       string
	Status     string
	Topology   string
	Sites      int
	Hubs       int
	Bandwidth  string
	Protocol   string
	Encryption string
	QoS        bool
	Uptime     string
	Region     string
}

type sdwanStoreType struct {
	mu       sync.RWMutex
	networks map[string][]SDWAN
}

var sdwanStore = &sdwanStoreType{
	networks: map[string][]SDWAN{},
}

func (s *sdwanStoreType) list(slug string) []SDWAN {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.networks[slug]
	out := make([]SDWAN, len(src))
	copy(out, src)
	return out
}

func (s *sdwanStoreType) getByID(slug, id string) (SDWAN, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, n := range s.networks[slug] {
		if n.ID == id {
			return n, true
		}
	}
	return SDWAN{}, false
}

func (s *sdwanStoreType) add(slug string, n SDWAN) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.networks[slug] = append(s.networks[slug], n)
}

func (s *sdwanStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.networks[slug]
	for i, n := range list {
		if n.ID == id {
			s.networks[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

type apiSDWAN struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Status     string `json:"status"`
	Topology   string `json:"topology"`
	Sites      int    `json:"sites"`
	Hubs       int    `json:"hubs"`
	Bandwidth  string `json:"bandwidth"`
	Protocol   string `json:"protocol"`
	Encryption string `json:"encryption"`
	QoS        bool   `json:"qos"`
	Uptime     string `json:"uptime"`
	Region     string `json:"region"`
}

type apiSDWANList struct {
	Networks []apiSDWAN `json:"networks"`
	Total    int        `json:"total"`
}

func toAPISDWAN(n SDWAN) apiSDWAN {
	return apiSDWAN{
		ID: n.ID, Name: n.Name, Status: n.Status,
		Topology: n.Topology, Sites: n.Sites, Hubs: n.Hubs,
		Bandwidth: n.Bandwidth, Protocol: n.Protocol,
		Encryption: n.Encryption, QoS: n.QoS,
		Uptime: n.Uptime, Region: n.Region,
	}
}

func filterSDWAN(all []SDWAN, search, status string) []SDWAN {
	var out []SDWAN
	q := strings.ToLower(search)
	for _, n := range all {
		if status != "" && n.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(n.Name + " " + n.Topology + " " + n.Protocol + " " + n.Region)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, n)
	}
	return out
}

func apiWSNetPrivSDWANListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	nets := filterSDWAN(sdwanStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiSDWAN, len(nets))
	for i, n := range nets {
		out[i] = toAPISDWAN(n)
	}
	apiOK(w, apiSDWANList{Networks: out, Total: len(out)})
}

type apiCreateSDWANRequest struct {
	Name       string `json:"name"`
	Topology   string `json:"topology"`
	Protocol   string `json:"protocol"`
	Encryption string `json:"encryption"`
	Region     string `json:"region"`
	QoS        bool   `json:"qos"`
}

func apiWSNetPrivSDWANCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateSDWANRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		apiError(w, http.StatusBadRequest, "Name is required.")
		return
	}
	topology := strings.TrimSpace(req.Topology)
	if topology == "" {
		topology = "hub-and-spoke"
	}
	protocol := strings.TrimSpace(req.Protocol)
	if protocol == "" {
		protocol = "IPsec + VXLAN"
	}
	encryption := strings.TrimSpace(req.Encryption)
	if encryption == "" {
		encryption = "AES-256-GCM"
	}
	region := strings.TrimSpace(req.Region)
	if region == "" {
		region = "global"
	}
	net := SDWAN{
		ID: newID(), Name: name, Status: "provisioning",
		Topology: topology, Sites: 0, Hubs: 0,
		Bandwidth: "0 Mbps", Protocol: protocol,
		Encryption: encryption, QoS: req.QoS,
		Uptime: "—", Region: region,
	}
	sdwanStore.add(slug, net)
	apiCreated(w, toAPISDWAN(net))
}

func apiWSNetPrivSDWANGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	net, found := sdwanStore.getByID(slug, r.PathValue("nid"))
	if !found {
		apiError(w, http.StatusNotFound, "SD-WAN network not found.")
		return
	}
	apiOK(w, toAPISDWAN(net))
}

func apiWSNetPrivSDWANDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := sdwanStore.getByID(slug, r.PathValue("nid")); !found {
		apiError(w, http.StatusNotFound, "SD-WAN network not found.")
		return
	}
	sdwanStore.delete(slug, r.PathValue("nid"))
	apiOK(w, map[string]any{"ok": true, "message": "SD-WAN network deleted."})
}
