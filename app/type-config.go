package app

import (
	"kscan/lib/IP"
	"kscan/lib/chinese"
	"kscan/lib/misc"
	"kscan/lib/params"
	"kscan/lib/slog"
	"kscan/lib/urlparse"
	"os"
	"regexp"
	"runtime"
	"strings"
	"time"
)

type Config struct {
	HostTarget, UrlTarget       []string
	Port                        []int
	Output                      *os.File
	Proxy, Host, Path, Encoding string
	OSEncoding, NewLine         string
	OutputJson                  string
	Threads, Rarity             int
	Timeout                     time.Duration
	ScanPing, Check, NoColor    bool
	Spy                         string
	//hydra
	Hydra, HydraUpdate             bool
	HydraPass, HydraUser, HydraMod []string
	HydraPortArr                   []int
	HydraProtocolArr               []string
	HydraMap                       map[int]string
	//fofa
	Fofa, FofaField []string
	FofaFixKeywored string
	FofaSize        int
	Scan            bool
}

var Setting = New()

func (c *Config) WriteLine(s string) {
	if c.OSEncoding == "utf-8" {
		s = chinese.ToUTF8(s)
	} else {
		s = chinese.ToGBK(s)
	}
	s = s + c.NewLine
	_, _ = c.Output.WriteString(s)
}

func (c *Config) Load(p *params.OsArgs) {
	c.Spy = p.Spy()
	if p.Spy() == "None" {
		c.loadTarget(p.Target(), false)
	}
	c.loadPort(p.Port())
	c.loadPort(p.Top())
	c.loadOutput(p.Output())
	c.loadTimeout(p.Timeout())
	c.ScanPing = c.loadScanPing(p.ScanPing())
	c.Check = p.Check()
	c.Path = p.Path()
	c.Proxy = p.Proxy()
	c.Host = p.Host()
	c.Threads = p.Threads()
	c.Encoding = p.Encoding()
	c.Rarity = p.Rarity()
	c.OutputJson = p.OutputJson()
	c.NoColor = p.NoColor()
	//hydra模块
	c.Hydra = p.Hydra()
	c.HydraUpdate = p.HydraUpdate()
	c.HydraUser = c.makeHydraUser(p.HydraUser())
	c.HydraPass = c.makeHydraPass(p.HydraPass())
	c.HydraMap = c.makeHydraMap()
	c.HydraPortArr = c.makeHydraPortArr(c.HydraMap)
	c.HydraProtocolArr = c.makeHydraProtocolArr(c.HydraMap)
	c.loadHydraMod(p.HydraMod())
	//fofa模块
	c.Fofa = c.loadFofa(p.Fofa())
	c.FofaSize = p.FofaSize()
	c.FofaField = c.loadFofaField(p.FofaField())
	c.FofaFixKeywored = p.FofaFixKeyword()
	c.Scan = p.Scan()

}

func (c *Config) loadTarget(expr string, recursion bool) {
	if expr == "" {
		return
	}
	if strings.Contains(expr, ",") {
		for _, s := range strings.Split(expr, ",") {
			c.loadTarget(s, true)
		}
		return
	}
	//判断target字符串是否为文件
	if regexp.MustCompile("^file:.+$").MatchString(expr) {
		expr = strings.Replace(expr, "file:", "", 1)
		err := misc.ReadLine(expr, c.loadTarget)
		if err != nil {
			if recursion == true {
				slog.Debug(expr + err.Error())
			} else {
				slog.Error(expr + err.Error())
			}
		}
		c.HostTarget = misc.RemoveDuplicateElement(c.HostTarget)
		c.UrlTarget = misc.RemoveDuplicateElement(c.UrlTarget)
		return
	}
	//判断target字符串是否为类IP/MASK
	if ok := IP.FormatCheck(expr); ok {
		c.HostTarget = append(c.HostTarget, IP.ExprToList(expr)...)
		return
	}
	//判断target字符串是否为类URL
	if url, err := urlparse.Load(expr); err != nil {
		if recursion == true {
			slog.Debug(expr + err.Error())
		} else {
			slog.Error(expr + err.Error())
		}
	} else {
		if url.Scheme != "" {
			c.UrlTarget = append(c.UrlTarget, expr)
			c.HostTarget = append(c.HostTarget, url.Netloc)
			return
		} else {
			c.HostTarget = append(c.HostTarget, url.Netloc)
			return
		}
	}
}
func (c *Config) loadTimeout(i int) {
	c.Timeout = time.Duration(i) * time.Second
}

func (c *Config) loadPort(v interface{}) {
	switch v.(type) {
	case int:
		if v.(int) == 400 {
			return
		}
		c.Port = TOP_1000[:v.(int)]
	case string:
		if v.(string) == "" {
			return
		}
		c.Port = intParam2IntArr(v.(string))
	}
}

func (c *Config) loadOutput(expr string) {
	if expr == "" {
		return
	}
	f, err := os.OpenFile(expr, os.O_CREATE+os.O_RDWR, 0764)
	if err != nil {
		slog.Error(err.Error())
	} else {
		c.Output = f
	}
}

func (c *Config) makeHydraUser(expr string) []string {
	//判断对象是否为文件
	if regexp.MustCompile("^file:.+$").MatchString(expr) {
		path := strings.Replace(expr, "file:", "", 1)
		return misc.ReadLineAll(path)
	}
	//判断对象是否为多个
	if strArr := strings.ReplaceAll(expr, "\\,", "[DouHao]"); strings.Count(strArr, ",") > 0 {
		var userArr []string
		for _, str := range strings.Split(strArr, ",") {
			userArr = append(userArr, strings.ReplaceAll(str, "[DouHao]", ","))
		}
		return userArr
	}
	//对象为单个且不为空时直接返回
	if expr != "" {
		return []string{expr}
	}
	return []string{}
}

func (c *Config) makeHydraPass(expr string) []string {
	//判断对象是否为文件
	if regexp.MustCompile("^file:.+$").MatchString(expr) {
		path := strings.Replace(expr, "file:", "", 1)
		return misc.ReadLineAll(path)
	}
	//判断对象是否为多个
	if strArr := strings.ReplaceAll(expr, "\\,", "[DouHao]"); strings.Count(strArr, ",") > 0 {
		var passArr []string
		for _, str := range strings.Split(strArr, ",") {
			passArr = append(passArr, strings.ReplaceAll(str, "[DouHao]", ","))
		}
		return passArr
	}
	//对象为单个且不为空时直接返回
	if expr != "" {
		return []string{expr}
	}
	return []string{}
}

func (c *Config) makeHydraMap() map[int]string {
	return map[int]string{
		22:    "ssh",
		3389:  "rdp",
		3306:  "mysql",
		1433:  "mssql",
		1521:  "oracle",
		5432:  "postgresql",
		27017: "mongodb",
		6379:  "redis",
		//110:   "pop3",
		//995:   "pop3",
		//25:    "smtp",
		//994:   "smtp",
		//143:   "imap",
		//993:   "imap",
		//389:   "ldap",
		//23:   "telnet",
		21:   "ftp",
		2121: "ftp",
		//50000: "db2",
		445: "smb",
	}
}

func (c *Config) makeHydraPortArr(hydraMap map[int]string) []int {
	var intArr []int
	for key := range hydraMap {
		intArr = append(intArr, key)
	}
	return intArr
}

func (c *Config) makeHydraProtocolArr(hydraMap map[int]string) []string {
	var strArr []string
	for _, value := range hydraMap {
		strArr = append(strArr, value)
	}
	return strArr
}

func (c *Config) loadHydraMod(expr string) {
	if expr == "" || expr == "all" {
		return
	}
	var protocolArr []string
	if strings.Contains(expr, ",") {
		for _, protocol := range strings.Split(expr, ",") {
			if misc.IsInStrArr(c.HydraProtocolArr, protocol) {
				protocolArr = append(protocolArr, protocol)
			}
		}
	} else {
		if misc.IsInStrArr(c.HydraProtocolArr, expr) {
			protocolArr = append(protocolArr, expr)
		}
	}
	c.HydraMod = protocolArr

	hydraMap := make(map[int]string)
	for port, protocol := range c.HydraMap {
		if misc.IsInStrArr(c.HydraMod, protocol) {
			hydraMap[port] = protocol
		}
	}
	c.HydraPortArr = c.makeHydraPortArr(hydraMap)
	c.HydraProtocolArr = c.makeHydraProtocolArr(hydraMap)
}

func (c *Config) loadFofa(expr string) []string {
	//判断对象是否为文件
	if regexp.MustCompile("^file:.+$").MatchString(expr) {
		path := strings.Replace(expr, "file:", "", 1)
		return misc.ReadLineAll(path)
	}
	//判断对象是否为多个
	if strArr := strings.ReplaceAll(expr, "\\,", "[DouHao]"); strings.Count(strArr, ",") > 0 {
		var passArr []string
		for _, str := range strings.Split(strArr, ",") {
			passArr = append(passArr, strings.ReplaceAll(str, "[DouHao]", ","))
		}
		return passArr
	}
	//对象为单个且不为空时直接返回
	if expr != "" {
		return []string{expr}
	}
	return []string{}
}

func (c *Config) loadFofaField(expr string) []string {
	//判断对象是否为多个
	if strArr := strings.ReplaceAll(expr, "\\,", "[DouHao]"); strings.Count(strArr, ",") > 0 {
		var passArr []string
		for _, str := range strings.Split(strArr, ",") {
			passArr = append(passArr, strings.ReplaceAll(str, "[DouHao]", ","))
		}
		return passArr
	}
	//对象为单个且不为空时直接返回
	if expr != "" {
		return []string{expr}
	}
	return []string{}
}

func (c *Config) loadScanPing(ping bool) bool {
	if len(c.Port) < 10 {
		return true
	}
	return ping
}

func New() Config {
	return Config{
		HostTarget: []string{},
		UrlTarget:  []string{},
		Path:       "/",
		Port:       TOP_1000[:400],
		Output:     nil,
		Proxy:      "",
		Host:       "",
		Threads:    500,
		Timeout:    0,
		Encoding:   "utf-8",
		OSEncoding: getOSEncoding(),
		NewLine:    getNewline(),
	}
}

func getNewline() string {
	if runtime.GOOS == "windows" {
		return "\r\n"
	} else {
		return "\n"
	}
}

func getOSEncoding() string {
	if runtime.GOOS == "windows" {
		return "gb2312"
	} else {
		return "utf-8"
	}
}

func intParam2IntArr(v string) []int {
	var res []int
	vArr := strings.Split(v, ",")
	for _, v := range vArr {
		var vvArr []int
		if strings.Contains(v, "-") {
			iArr := strings.Split(v, "-")
			if len(iArr) != 2 {
				slog.Error("参数输入错误！！！")
			} else {
				smallNum := misc.Str2Int(iArr[0])
				bigNum := misc.Str2Int(iArr[1])
				if smallNum >= bigNum {
					slog.Error("参数输入错误！！！")
				}
				vvArr = append(vvArr, misc.Xrange(smallNum, bigNum)...)
			}
		} else {
			vvArr = append(vvArr, misc.Str2Int(v))
		}
		res = append(res, vvArr...)
	}
	return res
}

var TOP_1000 = []int{21, 22, 23, 25, 53, 69, 80, 81, 88, 89, 110, 135, 161, 445, 139,
	143, 389, 443, 512, 513, 514, 873, 1433, 1521, 2181, 3306, 3389, 3690, 4848, 5000,
	5001, 5432, 5632, 5900, 5901, 5902, 6379, 7000, 7001, 7002, 8000, 8001, 8007, 8008, 8009,
	8069, 8080, 8081, 8088, 8089, 8090, 8091, 9060, 9090, 9091, 9200, 9300, 10000, 11211, 27017,
	27018, 50000, 1080, 888, 1158, 2100, 2424, 2601, 2604, 3128, 5984, 7080, 8010, 8082, 8083,
	8084, 8085, 8086, 8087, 8222, 8443, 8686, 8888, 9000, 9001, 9002, 9003, 9004, 9005, 9006,
	9007, 9008, 9009, 9010, 9043, 9080, 9081, 9418, 9999, 50030, 50060, 50070, 82, 83, 84,
	85, 86, 87, 7003, 7004, 7005, 7006, 7007, 7008, 7009, 7010, 7070, 7071, 7072, 7073,
	7074, 7075, 7076, 7077, 7078, 7079, 8002, 8003, 8004, 8005, 8006, 8200, 90, 801, 8011,
	8100, 8012, 8070, 99, 7777, 8028, 808, 38888, 8181, 800, 18080, 8099, 8899, 8360, 8300,
	8800, 8180, 3505, 8053, 1000, 8989, 28017, 49166, 3000, 41516, 880, 8484, 6677, 8016, 7200,
	9085, 5555, 8280, 1980, 8161, 7890, 8060, 6080, 8880, 8020, 889, 8881, 38501, 1010, 93,
	6666, 100, 6789, 7060, 8018, 8022, 3050, 8787, 2000, 10001, 8013, 6888, 8040, 10021, 2011,
	6006, 4000, 8055, 4430, 1723, 6060, 7788, 8066, 9898, 6001, 8801, 10040, 9998, 803, 6688,
	10080, 8050, 7011, 40310, 18090, 802, 10003, 8014, 2080, 7288, 8044, 9992, 8889, 5644, 8886,
	9500, 58031, 9020, 8015, 8887, 8021, 8700, 91, 9900, 9191, 3312, 8186, 8735, 8380, 1234,
	38080, 9088, 9988, 2110, 21245, 3333, 2046, 9061, 2375, 9011, 8061, 8093, 9876, 8030, 8282,
	60465, 2222, 98, 1100, 18081, 70, 8383, 5155, 92, 8188, 2517, 8062, 11324, 2008, 9231,
	999, 28214, 16080, 8092, 8987, 8038, 809, 2010, 8983, 7700, 3535, 7921, 9093, 11080, 6778,
	805, 9083, 8073, 10002, 114, 2012, 701, 8810, 8400, 9099, 8098, 8808, 20000, 8065, 8822,
	15000, 9901, 11158, 1107, 28099, 12345, 2006, 9527, 51106, 688, 25006, 8045, 8023, 8029, 9997,
	7048, 8580, 8585, 2001, 8035, 10088, 20022, 4001, 2013, 20808, 8095, 106, 3580, 7742, 8119,
	6868, 32766, 50075, 7272, 3380, 3220, 7801, 5256, 5255, 10086, 1300, 5200, 8096, 6198, 6889,
	3503, 6088, 9991, 806, 5050, 8183, 8688, 1001, 58080, 1182, 9025, 8112, 7776, 7321, 235,
	8077, 8500, 11347, 7081, 8877, 8480, 9182, 58000, 8026, 11001, 10089, 5888, 8196, 8078, 9995,
	2014, 5656, 8019, 5003, 8481, 6002, 9889, 9015, 8866, 8182, 8057, 8399, 10010, 8308, 511,
	12881, 4016, 8042, 1039, 28080, 5678, 7500, 8051, 18801, 15018, 15888, 38443, 8123, 8144, 94,
	9070, 1800, 9112, 8990, 3456, 2051, 9098, 444, 9131, 97, 7100, 7711, 7180, 11000, 8037,
	6988, 122, 8885, 14007, 8184, 7012, 8079, 9888, 9301, 59999, 49705, 1979, 8900, 5080, 5013,
	1550, 8844, 4850, 206, 5156, 8813, 3030, 1790, 8802, 9012, 5544, 3721, 8980, 10009, 8043,
	8390, 7943, 8381, 8056, 7111, 1500, 7088, 5881, 9437, 5655, 8102, 6000, 65486, 4443, 10025,
	8024, 8333, 8666, 103, 8, 9666, 8999, 9111, 8071, 9092, 522, 11381, 20806, 8041, 1085,
	8864, 7900, 1700, 8036, 8032, 8033, 8111, 60022, 955, 3080, 8788, 7443, 8192, 6969, 9909,
	5002, 9990, 188, 8910, 9022, 10004, 866, 8582, 4300, 9101, 6879, 8891, 4567, 4440, 10051,
	10068, 50080, 8341, 30001, 6890, 8168, 8955, 16788, 8190, 18060, 7041, 42424, 8848, 15693, 2521,
	19010, 18103, 6010, 8898, 9910, 9190, 9082, 8260, 8445, 1680, 8890, 8649, 30082, 3013, 30000,
	2480, 7202, 9704, 5233, 8991, 11366, 7888, 8780, 7129, 6600, 9443, 47088, 7791, 18888, 50045,
	15672, 9089, 2585, 60, 9494, 31945, 2060, 8610, 8860, 58060, 6118, 2348, 8097, 38000, 18880,
	13382, 6611, 8064, 7101, 5081, 7380, 7942, 10016, 8027, 2093, 403, 9014, 8133, 6886, 95,
	8058, 9201, 6443, 5966, 27000, 7017, 6680, 8401, 9036, 8988, 8806, 6180, 421, 423, 57880,
	7778, 18881, 812, 15004, 9110, 8213, 8868, 1213, 8193, 8956, 1108, 778, 65000, 7020, 1122,
	9031, 17000, 8039, 8600, 50090, 1863, 8191, 65, 6587, 8136, 9507, 132, 200, 2070, 308,
	5811, 3465, 8680, 7999, 7084, 18082, 3938, 18001, 9595, 442, 4433, 7171, 9084, 7567, 811,
	1128, 6003, 2125, 6090, 10007, 7022, 1949, 6565, 65001, 1301, 19244, 10087, 8025, 5098, 21080,
	1200, 15801, 1005, 22343, 7086, 8601, 6259, 7102, 10333, 211, 10082, 18085, 180, 40000, 7021,
	7702, 66, 38086, 666, 6603, 1212, 65493, 96, 9053, 7031, 23454, 30088, 6226, 8660, 6170,
	8972, 9981, 48080, 9086, 10118, 40069, 28780, 20153, 20021, 20151, 58898, 10066, 1818, 9914, 55351,
	8343, 18000, 6546, 3880, 8902, 22222, 19045, 5561, 7979, 5203, 8879, 50240, 49960, 2007, 1722,
	8913, 8912, 9504, 8103, 8567, 1666, 8720, 8197, 3012, 8220, 9039, 5898, 925, 38517, 8382,
	6842, 8895, 2808, 447, 3600, 3606, 9095, 45177, 19101, 171, 133, 8189, 7108, 10154, 47078,
	6800, 8122, 381, 1443, 15580, 23352, 3443, 1180, 268, 2382, 43651, 10099, 65533, 7018, 60010,
	60101, 6699, 2005, 18002, 2009, 59777, 591, 1933, 9013, 8477, 9696, 9030, 2015, 7925, 6510,
	18803, 280, 5601, 2901, 2301, 5201, 302, 610, 8031, 5552, 8809, 6869, 9212, 17095, 20001,
	8781, 25024, 5280, 7909, 17003, 1088, 7117, 20052, 1900, 10038, 30551, 9980, 9180, 59009, 28280,
	7028, 61999, 7915, 8384, 9918, 9919, 55858, 7215, 77, 9845, 20140, 8288, 7856, 1982, 1123,
	17777, 8839, 208, 2886, 877, 6101, 5100, 804, 983, 5600, 8402, 5887, 8322, 770, 13333,
	7330, 3216, 31188, 47583, 8710, 22580, 1042, 2020, 34440, 20, 7703, 65055, 8997, 6543, 6388,
	8283, 7201, 4040, 61081, 12001, 3588, 7123, 2490, 4389, 1313, 19080, 9050, 6920, 299, 20046,
	8892, 9302, 7899, 30058, 7094, 6801, 321, 1356, 12333, 11362, 11372, 6602, 7709, 45149, 3668,
	517, 9912, 9096, 8130, 7050, 7713, 40080, 8104, 13988, 18264, 8799, 55070, 23458, 8176, 9517,
	9541, 9542, 9512, 8905, 11660, 1025, 44445, 44401, 17173, 436, 560, 733, 968, 602, 3133,
	3398, 16580, 8488, 8901, 8512, 10443, 9113, 9119, 6606, 22080, 5560, 7, 5757, 1600, 8250,
	10024, 10200, 333, 73, 7547, 8054, 6372, 223, 3737, 9800, 9019, 8067, 45692, 15400, 15698,
	9038, 37006, 2086, 1002, 9188, 8094, 8201, 8202, 30030, 2663, 9105, 10017, 4503, 1104, 8893,
	40001, 27779, 3010, 7083, 5010, 5501, 309, 1389, 10070, 10069, 10056, 3094, 10057, 10078, 10050,
	10060, 10098, 4180, 10777, 270, 6365, 9801, 1046, 7140, 1004, 9198, 8465, 8548, 108, 30015,
	8153, 1020, 50100, 8391, 34899, 7090, 6100, 8777, 8298, 8281, 7023, 3377, 9100,
}
