package hueemulator

import (
	"bytes"
	"log"
	"net"
	"strings"
	"text/template"

	"github.com/gin-gonic/gin"
)

const (
	upnp_multicast_address = "239.255.255.250:1900"
	upnp_uri               = "/upnp/setup.xml"
)

//OPT: "http://schemas.upnp.org/upnp/1/0/"; ns=01
var responseTemplateText = `HTTP/1.1 200 OK
CACHE-CONTROL: max-age=86400
EXT:
LOCATION: http://{{.}}
SERVER: FreeRTOS/6.0.5, UPnP/1.0, IpBridge/0.1                                                                                                  
hue-bridgeid: 001E06FFFE123456 
ST: urn:schemas-upnp-org:device:basic:1
USN: uuid:Socket-1_0-221438K0100073::urn:schemas-upnp-org:device:basic:1

`

//var setupTemplateText = `<?xml version="1.0"?>
//<root xmlns="urn:schemas-upnp-org:device-1-0">
//<specVersion>
//<major>1</major>
//<minor>0</minor>
//</specVersion>
//<URLBase>http://{{.}}/</URLBase>
//<device>
//<deviceType>urn:schemas-upnp-org:device:Basic:1</deviceType>
//<friendlyName>huejack</friendlyName>
//<manufacturer>Royal Philips Electronics</manufacturer>
//<modelName>Philips hue bridge 2012</modelName>
//<modelNumber>929000226503</modelNumber>
//<UDN>uuid:f6543a06-800d-48ba-8d8f-bc2949eddc33</UDN>
//</device>
//</root>`

var setupTemplateText = `
<?xml version="1.0"?>
<root xmlns="urn:schemas-upnp-org:device-1-0">
<specVersion>
<major>1</major>
<minor>0</minor>
</specVersion>
<URLBase>http://{{.}}/</URLBase>
<device>
<deviceType>urn:schemas-upnp-org:device:Basic:1</deviceType>
<friendlyName>Amazon-Echo-HA-Bridge (192.168.42.102)</friendlyName>
<manufacturer>Royal Philips Electronics</manufacturer>
<manufacturerURL>http://www.armzilla..com</manufacturerURL>
<modelDescription>Hue Emulator for Amazon Echo bridge</modelDescription>
<modelName>Philips hue bridge 2012</modelName>
<modelNumber>929000226503</modelNumber>
<modelURL>http://www.armzilla.com/amazon-echo-ha-bridge</modelURL>
<serialNumber>01189998819991197253</serialNumber>
<UDN>uuid:88f6698f-2c83-4393-bd03-cd54a9f8595</UDN>
<serviceList>
<service>
<serviceType>(null)</serviceType>
<serviceId>(null)</serviceId>
<controlURL>(null)</controlURL>
<eventSubURL>(null)</eventSubURL>
<SCPDURL>(null)</SCPDURL>
</service>
</serviceList>
<presentationURL>index.html</presentationURL>
<iconList>
<icon>
<mimetype>image/png</mimetype>
<height>48</height>
<width>48</width>
<depth>24</depth>
<url>hue_logo_0.png</url>
</icon>
<icon>
<mimetype>image/png</mimetype>
<height>120</height>
<width>120</width>
<depth>24</depth>
<url>hue_logo_3.png</url>
</icon>
</iconList>
</device>
</root>`

type upnpData struct {
	Addr string
	Uri  string
}

var setupTemplate *template.Template

func upnpTemplateInit() {
	var err error
	setupTemplate, err = template.New("").Parse(setupTemplateText)
	if err != nil {
		log.Fatalln("upnpTemplateInit:", err)
	}
}

func upnpSetup(addr string) gin.HandlerFunc {
	return func(c *gin.Context) {
		//w.Header().Set("Content-Type", "application/xml")
		err := setupTemplate.Execute(c.Writer, addr)
		if err != nil {
			log.Fatalln("[WEB] upnpSetup:", err)
		}
	}
}

func upnpResponder(hostAddr string, endpoint string) {
	responseTemplate, err := template.New("").Parse(responseTemplateText)

	log.Println("[UPNP] listening...")
	addr, err := net.ResolveUDPAddr("udp", upnp_multicast_address)
	if err != nil {
		log.Fatal(err)
	}
	l, err := net.ListenMulticastUDP("udp", nil, addr)
	l.SetReadBuffer(1024)

	for {
		b := make([]byte, 1024)
		n, src, err := l.ReadFromUDP(b)
		if err != nil {
			log.Fatal("[UPNP] ReadFromUDP failed:", err)
		}

		if strings.Contains(string(b[:n]), "MAN: \"ssdp:discover\"") {
			c, err := net.DialUDP("udp", nil, src)
			if err != nil {
				log.Fatal("[UPNP] DialUDP failed:", err)
			}

			log.Println("[UPNP] discovery request from", src)

			// For whatever reason I can't execute the template using c as the reader,
			// you HAVE to put it in a buffer first
			// possible timing issue?
			// don't believe me? try it
			b := &bytes.Buffer{}
			err = responseTemplate.Execute(b, hostAddr+endpoint)
			if err != nil {
				log.Fatal("[UPNP] execute template failed:", err)
			}
			c.Write(b.Bytes())
		}
	}
}
