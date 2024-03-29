package edp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var startSeq = []byte{0x45, 0x2, 0x0, 0x3e, 0x0, 0x0, 0x0, 0xe8, 0x3, 0x0, 0x0, 0x1, 0x0, 0x0, 0x0, 0x2, 0x0, 0x95, 0xa1, 0x33, 0x0, 0x45, 0x32}

func TestDecodeZoneOpen(t *testing.T) {
	// [#1000|21155703112020|ZO|8|Kök IR¦ZONE¦1¦Larm||0]
	d := []byte{0x45, 0x2, 0x0, 0x3e, 0x0, 0x0, 0x0, 0xe8, 0x3, 0x0, 0x0, 0x1, 0x0, 0x0, 0x0, 0x2, 0x0, 0x95, 0xa1, 0x33, 0x0, 0x45, 0x32, 0x5b, 0x23, 0x31, 0x30, 0x30, 0x30, 0x7c, 0x32, 0x31, 0x31, 0x35, 0x35, 0x37, 0x30, 0x33, 0x31, 0x31, 0x32, 0x30, 0x32, 0x30, 0x7c, 0x5a, 0x4f, 0x7c, 0x38, 0x7c, 0x4b, 0xf6, 0x6b, 0x20, 0x49, 0x52, 0xa6, 0x5a, 0x4f, 0x4e, 0x45, 0xa6, 0x31, 0xa6, 0x4c, 0x61, 0x72, 0x6d, 0x7c, 0x7c, 0x30, 0x5d}

	pkg, err := Decode(d)

	assert.NoError(t, err)
	assert.Contains(t, pkg.Time.String(), "2020-11-03 21:15:57")
	assert.Equal(t, 1, pkg.Area)
	assert.Equal(t, 1000, pkg.SystemID)
	assert.Equal(t, "Kök IR", pkg.Name)
	assert.Equal(t, "8", pkg.ID)
	assert.Equal(t, "ZO", pkg.Class)

	dev, dev2 := GenerateDevice(pkg)
	assert.Nil(t, dev2)
	assert.Equal(t, "Zone Kök IR", dev.Name)
	assert.Equal(t, ".zone.8", dev.ID.String())
}

func TestDecodeCloseArea(t *testing.T) {
	// E2[#1000|19525404112020|CG|1|Larm¦Jonas¦1||0] arm area
	// E2[#1000|19531104112020|OG|1|Larm¦Jonas¦1||0] disarm area
	//   [#1000|21154904112020|NL|1|Larm¦Jonas¦1||0] perimiter armed
	d := []byte{
		0x45, 0x02, 0x00, 0x77, 0x0e, 0x00, 0x00, 0xe8, 0x03, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x02,
		0x00, 0xd4, 0x01, 0x2d, 0x00, 0x45, 0x32, 0x5b, 0x23, 0x31, 0x30, 0x30, 0x30, 0x7c, 0x31, 0x39,
		0x35, 0x32, 0x35, 0x34, 0x30, 0x34, 0x31, 0x31, 0x32, 0x30, 0x32, 0x30, 0x7c, 0x43, 0x47, 0x7c,
		0x31, 0x7c, 0x4c, 0x61, 0x72, 0x6d, 0xa6, 0x4a, 0x6f, 0x6e, 0x61, 0x73, 0xa6, 0x31, 0x7c, 0x7c,
		0x30, 0x5d,
	}

	pkg, err := Decode(d)

	assert.NoError(t, err)
	assert.Contains(t, pkg.Time.String(), "2020-11-04 19:52:54")
	assert.Equal(t, "Jonas", pkg.UserName)
	assert.Equal(t, 1000, pkg.SystemID)
	assert.Equal(t, "Larm", pkg.Name)
	assert.Equal(t, "1", pkg.ID)
	assert.Equal(t, "CG", pkg.Class)

	dev, dev2 := GenerateDevice(pkg)
	assert.Nil(t, dev2)
	assert.Equal(t, "Area Larm", dev.Name)
	assert.Equal(t, ".area.1", dev.ID.String())
}

func TestFireAlarm(t *testing.T) {
	str := []byte("[#1000|07442202062022|FA|1|Brandvarnare\xa6ZONE\xa62\xa6Brandlarm||0]")

	pkg, err := Decode(append(startSeq, str...))
	assert.NoError(t, err)

	dev1, dev2 := GenerateDevice(pkg)

	assert.Equal(t, "zone.1", dev1.ID.ID)
	assert.Equal(t, "area.2", dev2.ID.ID)
	assert.Equal(t, true, dev2.State["fire"])
	assert.Equal(t, true, dev1.State["fire"])
}
func TestFireAlarmRestore(t *testing.T) {
	str := []byte("[#1000|07442202062022|FR|1|Brandvarnare\xa6ZONE\xa62\xa6Brandlarm||0]")

	pkg, err := Decode(append(startSeq, str...))
	assert.NoError(t, err)

	dev1, dev2 := GenerateDevice(pkg)

	assert.Equal(t, "zone.1", dev1.ID.ID)
	assert.Equal(t, "area.2", dev2.ID.ID)
	assert.Equal(t, false, dev2.State["fire"])
	assert.Equal(t, false, dev1.State["fire"])
}
func TestModemFail(t *testing.T) {
	str := []byte("[#1000|19480006062022|YS|1|Telelinjefel\xa61||0]")

	pkg, err := Decode(append(startSeq, str...))
	assert.NoError(t, err)

	dev1, dev2 := GenerateDevice(pkg)

	assert.Nil(t, dev2)
	assert.Equal(t, "modem.1", dev1.ID.ID)
	assert.Equal(t, "Modem", dev1.Name)
	assert.Equal(t, true, dev1.State["error"])
	// spew.Dump(dev1)
}
func TestModemFailRestore(t *testing.T) {
	str := []byte("[#1000|21023306062022|YK|1|Telelinjefel \xe5terst\xe4llt\xa61||0]")

	pkg, err := Decode(append(startSeq, str...))
	assert.NoError(t, err)

	dev1, dev2 := GenerateDevice(pkg)

	assert.Nil(t, dev2)
	assert.Equal(t, "modem.1", dev1.ID.ID)
	assert.Equal(t, "Modem", dev1.Name)
	assert.Equal(t, false, dev1.State["error"])
	// spew.Dump(dev1)
}
