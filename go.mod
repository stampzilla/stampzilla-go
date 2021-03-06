module github.com/stampzilla/stampzilla-go/v2

go 1.15

require (
	github.com/RangelReale/osin v1.0.1
	github.com/StackExchange/wmi v0.0.0-20190523213315-cbe66965904d // indirect
	github.com/cloudfoundry/gosigar v1.1.0
	github.com/coreos/go-systemd/v22 v22.1.0
	github.com/cpuguy83/go-md2man/v2 v2.0.0 // indirect
	github.com/davecgh/go-spew v1.1.1
	github.com/dustin/go-humanize v1.0.0
	github.com/faiface/beep v1.0.1
	github.com/fatih/camelcase v1.0.0 // indirect
	github.com/fatih/color v1.9.0
	github.com/fatih/structs v1.1.0 // indirect
	github.com/gin-contrib/cors v1.3.0
	github.com/gin-contrib/gzip v0.0.1
	github.com/gin-contrib/sessions v0.0.3
	github.com/gin-gonic/gin v1.6.3
	github.com/go-ole/go-ole v1.2.4 // indirect
	github.com/go-playground/validator/v10 v10.4.1 // indirect
	github.com/goburrow/modbus v0.1.0
	github.com/goburrow/serial v0.1.0 // indirect
	github.com/golang/protobuf v1.4.3 // indirect
	github.com/google/cel-go v0.6.0
	github.com/google/go-github v17.0.0+incompatible
	github.com/google/go-querystring v1.0.0 // indirect
	github.com/google/uuid v1.1.2
	github.com/gorilla/websocket v1.4.2
	github.com/influxdata/influxdb1-client v0.0.0-20190809212627-fc22c7df067e
	github.com/itchyny/volume-go v0.0.0
	github.com/jonaz/astrotime v0.0.0-20150127084258-5d2b676e5047 // indirect
	github.com/jonaz/cron v0.0.0-20190121203350-e9ab53dd31db
	github.com/jonaz/ginlogrus v0.0.0-20191118094232-2f4da50f5dd6
	github.com/jonaz/goenocean v0.0.0-20190218201525-96fde8f44745
	github.com/jonaz/gograce v0.0.0-20170710084154-582d7afa93c3
	github.com/json-iterator/go v1.1.10 // indirect
	github.com/koding/multiconfig v0.0.0-20171124222453-69c27309b2d7
	github.com/llgcode/draw2d v0.0.0-20190810100245-79e59b6b8fbc
	github.com/micro/mdns v0.3.0
	github.com/olahol/melody v0.0.0-20180227134253-7bd65910e5ab
	github.com/onrik/logrus v0.4.1
	github.com/onsi/ginkgo v1.10.2 // indirect
	github.com/onsi/gomega v1.7.0 // indirect
	github.com/pborman/uuid v1.2.0 // indirect
	github.com/pkg/errors v0.8.1
	github.com/posener/wstest v0.0.0-20180217133618-28272a7ea048
	github.com/rakyll/statik v0.1.7
	github.com/shirou/gopsutil v2.19.9+incompatible
	github.com/shirou/w32 v0.0.0-20160930032740-bb4de0191aa4 // indirect
	github.com/sirupsen/logrus v1.7.0
	github.com/stamp/hid v0.0.0-20190105143849-bc55d7d13ce1
	github.com/stamp/mdns v0.0.0-20190125083120-df204ac59ca1
	github.com/stampzilla/gocast v0.0.0-20201206204400-b3419c6ce229
	github.com/stampzilla/gozwave v0.0.0-20190221183844-576012815e01
	github.com/stretchr/testify v1.6.1
	github.com/tarm/goserial v0.0.0-20151007205400-b3440c3c6355 // indirect
	github.com/tarm/serial v0.0.0-20180830185346-98f6abe2eb07
	github.com/ugorji/go v1.1.13 // indirect
	github.com/urfave/cli v1.22.1
	github.com/vapourismo/knx-go v0.0.0-20200924173532-6ed2bf1e51e6
	golang.org/x/crypto v0.0.0-20201016220609-9e8e0b390897
	golang.org/x/sys v0.0.0-20201204225414-ed752295db88 // indirect
	google.golang.org/genproto v0.0.0-20200526211855-cb27e3aa2013
	google.golang.org/protobuf v1.25.0 // indirect
	gopkg.in/yaml.v2 v2.3.0 // indirect
	gopkg.in/yaml.v3 v3.0.0-20200615113413-eeeca48fe776 // indirect
	software.sslmate.com/src/go-pkcs12 v0.0.0-20200619203921-c9ed90bd32dc
)

//replace github.com/stampzilla/gocast => /home/jonaz/go/src/github.com/stampzilla/gocast
