package nibe

import (
	//"log"

	"github.com/fatih/color"
)

// Controller request package
// [ 0x5C ][ 0x00 ][ ADDR ][ CMD ][ LEN ][ DATA... ][ CHK ]
//
// Known addresses
// 0x14 - RCU10
// 0x19 - RMU40
// 0x20 - Modbus40
// 0x22 - Sam40
//
// CMD
// 0x68 - Cyclic log data (sent every other second)
//        body:  [ 16bit reg1 ][ 16bit value1 ][ 16bit reg2 ][ 16bit value2 ]
//               [ 16bit reg3 ][ 16bit value3 ][ 16bit reg4 ][ 16bit value4 ]
//               [ 16bit reg5 ][ 16bit value5 ][ 16bit reg6 ][ 16bit value6 ]
//               [ 16bit reg7 ][ 16bit value7 ][ 16bit reg8 ][ 16bit value8 ]
//               [ 16bit reg9 ][ 16bit value9 ][ 16bit reg10 ][ 16bit value10 ]
//        response: 0x06 (ack)
// 0x69 - Read token (controller asks Modbus40 if it wants to read a parameter)
//        (no body)
//        response: [ 0xC0 ][ 0x69 ][ LEN ][ 16bit reg ][ CHK ] (request read a register)
//           OR
//        response: [ 0x06 ] (ack)
//
// 0x6A - Read response
//        [ 16bit value ]
//        response: 0x06 (ack)
//
// 0x6B - Write token (controller asks Modbus40 if it wants to write a parameter)
//        (no body)
//        response: [ 0xC0 ][ 0x6B ][ LEN ][ 16bit reg ][ 16bit new value ][ CHK ] (request write a register)
//           OR
//        response: [ 0x06 ] (ack)
//
// 0x6C - Write response
//        [ 16bit value ]
//
// 0x6D - Connected controller information
//        [ 4byte unknown ][ ?byte controllername as string ]
//        response: 0x06 (ack)
//
// 0xEE - Device information request (controller asks Modbus40 what address it has)
//        (no body)
//        response: [ 0xC0][ 0xEE ][ LEN ][ 16bit modbus address ][ CHK ] (the address is then shown in the status display)
//

func decodeMessage(data []byte) (*Message, []byte) {
	for i, val := range data {
		if val == 0x5c && data[i+1] == 0x00 && len(data) > i+5 {
			junk := data[:i]
			addr := data[i+2]
			cmd := data[i+3]
			length := int(data[i+4])

			if len(data) < i+6+length {
				return nil, data
			}

			content := data[i+5 : i+5+length]
			chk := data[i+5+length]

			chk2 := byte(0)
			for _, v := range data[i+2 : i+5+length] {
				chk2 ^= v
			}

			if chk == chk2 {
				if len(junk) > 0 {
					sendLog := color.New(color.FgCyan)
					sendLog.Printf("Junk - %x\n", junk)
					//if (len(junk)>19) {
					//var parameter struct {
					//	Value uint16
					//}
					//r := bytes.NewReader(junk[3:19])
					//for {
					//	err := binary.Read(r, binary.LittleEndian, &parameter);
					//	if (err != nil) {
					//		break
					//	}

					//	log.Printf(" - %f - %d", 45 -(float64(parameter.Value) / 1023 * 33), parameter.Value);
					//}
					//}

				}
				//log.Printf(" <- %s (%x) - cmd %x len %d data %x\n", devices[addr], addr, cmd, length, content)

				msg := &Message{
					Addr: addr,
					Cmd:  cmd,
					Data: content,
				}

				return msg, data[i+6+length:]
			} else {
				//log.Printf(" <- %s (%x) - cmd %x len %d data %x | %x != %x\n", devices[addr], addr, cmd, length, content, chk, chk2)

				return nil, data[i+6+length:]
			}
		}

		if val == 0xC0 && len(data) > i+4 { // Response from ex sam40
			cmd := data[i+1]
			length := int(data[i+2])

			if len(data) < i+4+length {
				return nil, data
			}

			content := data[i+3 : i+3+length]
			//chk := data[i+3+length]

			chk2 := byte(0)
			for _, v := range data[i : i+3+length] {
				chk2 ^= v
			}

			//log.Printf(" <- RESPONSE %x len %d data %x | %x ?= %x\n", cmd, length, content, chk, chk2)

			msg := &Message{
				Cmd:  cmd,
				Data: content,
			}
			return msg, data[i+4+length:]
		}

		if val == 0x06 { // ACK
			//log.Printf(" <- ACK\n")
			msg := &Message{
				Cmd: val,
			}
			return msg, data[i+1:]
		}
	}

	return nil, data
}

func encodeMessage(cmd byte, addr uint16, data []byte) []byte {
	msg := append(
		append(
			[]byte{0xC0},
			cmd,
			byte(len(data)+2),
			byte(addr),
			byte(addr>>8),
		),
		data...,
	)
	chk := byte(0)
	for _, v := range msg {
		chk ^= v
	}
	msg = append([]byte{}, append(msg, chk)...)

	return msg
}
