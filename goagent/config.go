package main

import (
	"fmt"
	"github.com/dlintw/goconf"
	"regexp"
	"strings"
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
	GaeWindow           int
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
	PacEnable           bool
	PacIp               string
	PacPort             int
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
	PhpFetchserver      []string
	PhpHosts            []string
	VpsEnable           bool
	VpsListen           string
	VpsFetchserver      []string
	ProxyEnable         bool
	ProxyAutodetect     bool
	ProxyHost           string
	ProxyPort           int
	ProxyUsername       string
	ProxyPasswrod       string
	AutorangeHosts      []string
	AutorangeEndswith   []string
	AutorangeNoendswith []string
	AutorangeMaxsize    int
	AutorangeWaitsize   int
	AutorangeBufsize    int
	AutorangeThreads    int
	FetchmaxLocal       int
	FetchmaxServer      int
	DnsEnable           bool
	DnsListen           string
	DnsServers          []string
	UseragentEnable     bool
	UseragentString     string
	LoveEnable          bool
	LoveTip             []string

	IplistCNames      map[string][]string
	IplistFixed       []string
	IpListMap         map[string]string
	UrlRewriteMap     map[string]string
	CrlfSites         []string
	NocrlfSites       []string
	ForcehttpsSites   []string
	NoforcehttpsSites []string
	FakehttpsSites    []string
	NofakehttpsSites  []string
	WithGAESites      []string
	WithPHPSites      []string
	WithVPSSites      []string
}

func getString(c *goconf.ConfigFile, section string, option string) string {
	value, err := c.GetString(section, option)
	if err != nil {
		panic(err)
	}
	return value
}

func getStrings(c *goconf.ConfigFile, section string, option string) []string {
	value, err := c.GetString(section, option)
	if err != nil {
		panic(err)
	}
	return regexp.MustCompile(`[|,]`).Split(value, -1)
}

func getInt(c *goconf.ConfigFile, section string, option string) int {
	value, err := c.GetInt(section, option)
	if err != nil {
		panic(err)
	}
	return value
}

func getBool(c *goconf.ConfigFile, section string, option string) bool {
	value, err := c.GetBool(section, option)
	if err != nil {
		panic(err)
	}
	return value
}

func ReadConfigFile(filename string) (*CommonConfig, error) {
	cc := &CommonConfig{}
	c, err := goconf.ReadConfigFile(filename)
	if err != nil {
		return nil, err
	}

	cc.ListenIp = getString(c, "listen", "ip")
	cc.ListenPort = getInt(c, "listen", "port")
	cc.ListenUsername = getString(c, "listen", "username")
	cc.ListenPassword = getString(c, "listen", "password")
	cc.ListenVisible = getBool(c, "listen", "visible")
	cc.ListenDebuginfo = getBool(c, "listen", "debuginfo")

	cc.GaeEnable = getBool(c, "gae", "enable")
	cc.GaeAppids = getStrings(c, "gae", "appid")
	cc.GaePassword = getString(c, "gae", "password")
	cc.GaePath = getString(c, "gae", "path")
	cc.GaeMode = getString(c, "gae", "mode")
	cc.GaeIpv6 = getBool(c, "gae", "ipv6")
	cc.GaeSslversion = getString(c, "gae", "sslversion")
	cc.GaeWindow = getInt(c, "gae", "window")
	cc.GaeCachesock = getBool(c, "gae", "cachesock")
	cc.GaeHeadfirst = getBool(c, "gae", "headfirst")
	cc.GaeKeepalive = getBool(c, "gae", "keepalive")
	cc.GaeObfuscate = getBool(c, "gae", "obfuscate")
	cc.GaeValidate = getBool(c, "gae", "validate")
	cc.GaeTransport = getBool(c, "gae", "transport")
	cc.GaeOptions = getStrings(c, "gae", "options")
	cc.GaeRegions = getStrings(c, "gae", "regions")

	cc.PacEnable = getBool(c, "pac", "enable")
	cc.PacIp = getString(c, "pac", "ip")
	cc.PacPort = getInt(c, "pac", "port")
	cc.PacFile = getString(c, "pac", "file")
	cc.PacAdmode = getInt(c, "pac", "admode")
	cc.PacAdblock = getString(c, "pac", "adblock")
	cc.PacGfwlist = getString(c, "pac", "gfwlist")
	cc.PacExpired = getInt(c, "pac", "expired")

	cc.PhpEnable = getBool(c, "php", "enable")
	cc.PhpListen = getString(c, "php", "listen")
	cc.PhpPassword = getString(c, "php", "password")
	cc.PhpCrlf = getBool(c, "php", "crlf")
	cc.PhpValidate = getBool(c, "php", "validate")
	cc.PhpKeepalive = getBool(c, "php", "keepalive")
	cc.PhpFetchserver = getStrings(c, "php", "fetchserver")
	cc.PhpHosts = getStrings(c, "php", "hosts")

	cc.VpsEnable = getBool(c, "vps", "enable")
	cc.VpsListen = getString(c, "vps", "listen")
	cc.VpsFetchserver = getStrings(c, "vps", "fetchserver")

	cc.ProxyEnable = getBool(c, "proxy", "enable")
	cc.ProxyAutodetect = getBool(c, "proxy", "autodetect")
	cc.ProxyHost = getString(c, "proxy", "host")
	cc.ProxyPort = getInt(c, "proxy", "port")
	cc.ProxyUsername = getString(c, "proxy", "username")
	cc.ProxyPasswrod = getString(c, "proxy", "password")

	cc.AutorangeHosts = getStrings(c, "autorange", "hosts")
	cc.AutorangeEndswith = getStrings(c, "autorange", "endswith")
	cc.AutorangeNoendswith = getStrings(c, "autorange", "noendswith")
	cc.AutorangeThreads = getInt(c, "autorange", "threads")
	cc.AutorangeMaxsize = getInt(c, "autorange", "maxsize")
	cc.AutorangeWaitsize = getInt(c, "autorange", "waitsize")
	cc.AutorangeBufsize = getInt(c, "autorange", "bufsize")

	cc.DnsEnable = getBool(c, "dns", "enable")
	cc.DnsListen = getString(c, "dns", "listen")
	cc.DnsServers = getStrings(c, "dns", "servers")

	cc.UseragentEnable = getBool(c, "useragent", "enable")
	cc.UseragentString = getString(c, "useragent", "string")

	cc.FetchmaxLocal = getInt(c, "fetchmax", "local")
	cc.FetchmaxServer = getInt(c, "fetchmax", "server")

	cc.LoveEnable = getBool(c, "love", "enable")
	cc.LoveTip = getStrings(c, "love", "tip")

	cc.IplistCNames = make(map[string][]string)
	if options, err := c.GetOptions("iplist"); err == nil {
		for _, option := range options {
			cc.IplistCNames[option] = getStrings(c, "iplist", option)
		}
	} else {
		panic(err)
	}

	cc.IplistCNames = make(map[string][]string, 0)
	cc.IplistFixed = make([]string, 0)
	cc.IpListMap = make(map[string]string, 0)
	cc.UrlRewriteMap = make(map[string]string, 0)
	cc.CrlfSites = make([]string, 0)
	cc.NocrlfSites = make([]string, 0)
	cc.ForcehttpsSites = make([]string, 0)
	cc.NoforcehttpsSites = make([]string, 0)
	cc.FakehttpsSites = make([]string, 0)
	cc.NofakehttpsSites = make([]string, 0)
	cc.WithGAESites = make([]string, 0)
	cc.WithPHPSites = make([]string, 0)
	cc.WithVPSSites = make([]string, 0)
	if options, err := c.GetOptions("profile"); err == nil {
		for _, option := range options {
			pattern := option
			rules := getStrings(c, "profile", option)
			for {
				if len(rules) == 0 {
					break
				}
				rule := rules[0]
				rules = rules[1:]
				switch rule {
				case "crlf":
					cc.CrlfSites = append(cc.CrlfSites, pattern)
				case "nocrlf":
					cc.NocrlfSites = append(cc.NocrlfSites, pattern)
				case "forcehttps":
					cc.ForcehttpsSites = append(cc.ForcehttpsSites, pattern)
				case "noforcehttps":
					cc.NoforcehttpsSites = append(cc.NoforcehttpsSites, pattern)
				case "fakehttps":
					cc.FakehttpsSites = append(cc.FakehttpsSites, pattern)
				case "nofakehttps":
					cc.NofakehttpsSites = append(cc.NofakehttpsSites, pattern)
				case "withgae":
					cc.WithGAESites = append(cc.WithGAESites, pattern)
				case "withphp":
					cc.WithPHPSites = append(cc.WithPHPSites, pattern)
				case "withvps":
					cc.WithVPSSites = append(cc.WithVPSSites, pattern)
				case "direct":
					cc.IpListMap[pattern] = pattern
					cc.IplistCNames[pattern] = nil
				default:
					fmt.Printf("rule=%#v, pattern=%#v\n", rule, pattern)
					if _, ok := cc.IplistCNames[rule]; ok {
						cc.IpListMap[pattern] = rule
					} else if strings.Contains(pattern, "\\") {
						cc.UrlRewriteMap[pattern] = rule
					}
				}
			}
		}
	} else {
		panic(err)
	}

	return cc, nil
}
