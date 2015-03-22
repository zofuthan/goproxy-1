package main

import (
	"github.com/dlintw/goconf"
)

type CommonConfig struct {
	ListenIp            string
	ListenPort          int
	ListenUsername      string
	ListenPassword      string
	ListenVisible       bool
	ListenDebuginfo     bool
	GaeEnable           bool
	GaeAppids           []string
	GaePassword         string
	GaePath             string
	GaeMode             string
	GaeIpv6             bool
	GaeWindow           uint
	GaeKeepalive        bool
	GaeCachesock        bool
	GaeHeadfirst        bool
	GaeObfuscate        bool
	GaeValidate         bool
	GaeTransport        bool
	GaeOptions          []string
	GaeRegions          []string
	GaeSslversion       string
	GaePagespeed        bool
	WithGAESites        []string
	WithPHPSites        []string
	CrlfSites           []string
	NocrlfSites         []string
	ForcehttpsSites     []string
	NoforcehttpsSites   []string
	FakehttpsSites      []string
	NofakehttpsSites    []string
	UrlRewriteMap       map[string]string
	RuleMap             map[string]string
	IplistCNames        map[string]string
	IplistFixed         []string
	PacEnable           bool
	PacIp               string
	PacPort             uint
	PacFile             string
	PacGfwlist          string
	PacAdblock          string
	PacAdmode           int
	PacExpired          int
	PhpEnable           bool
	PhpListen           string
	PhpPassword         string
	PhpCrlf             bool
	PhpValidate         bool
	PhpKeepalive        bool
	PhpFetchserver      bool
	PhpHosts            []string
	VpsEnable           bool
	VpsListen           string
	VpsFetchserver      []string
	ProxyEnable         bool
	ProxyAutodetect     bool
	ProxyHost           string
	ProxyPort           uint
	ProxyUsername       string
	ProxyPasswrod       string
	AutorangeHosts      []string
	AutorangeEndswith   []string
	AutorangeNoendswith []string
	AutorangeMaxsize    uint
	AutorangeWaitsize   uint
	AutorangeBufsize    uint
	AutorangeThreads    uint
	FetchmaxLocal       uint
	FetchmaxServer      uint
	DnsEnable           bool
	DnsListen           string
	DnsServers          []string
	UseragentEnable     bool
	UseragentString     string
	LoveEnable          bool
	LoveTip             []string
}

func ReadConfigFile(filename string) (*CommonConfig, error) {
	cc := &CommonConfig{}
	c, err := goconf.ReadConfigFile("something.config")
	if err != nil {
		return nil, err
	}
	cc.ListenIp, _ = c.GetString("listen", "ip")
	return cc, nil
}
