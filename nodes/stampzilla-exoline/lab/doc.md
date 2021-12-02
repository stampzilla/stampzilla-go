EXOLine-TCP documentation
=========================

The document describes the current understanding of the EXOLine-TCP IPv4 protocol.
This information is copied from https://github.com/merbanan/EXOlink/blob/master/Specification.md in case that repo disapears

General description
-------------------

The protocol uses a synchronous client/server approch. The tcp port used is 26486.
The data payload syntax of server answers depends on the sent client query. So you
have to keep track of your query and tcp-connection to process the answer. It is assumed
that EXOLine-TCP is related to EXOLine (serial) and thus design limitations from that protocol
are present in EXOLine-TCP.



Message structure
-----------------

The messages are byte-based and formed like the following:

| Position      | Value         | Meaning|
| ------------- |:-------------:| -----:|
| 0             | 0x3c/0x3d     | Message start value (Client/Server) |
| X             | ...           | Payload (depends on command) |
| Last-1        | XORSUM        | Xorsum of all bytes after start value until XORSUM byte |
| Last          | 0x3e          | Message stop value |


Payload structure
-----------------

The payload consists of the following:

| Position      | Value         | Meaning|
| ------------- |:-------------:| -----:|
| 0             | 0xFF          | PLA |
| 1             | 0x1E          | ELA |
| 2             | 0xC8          | Multi Command marker |
| 3             | 0x04          | Length of bytes in this command block |
| 4             | 0xB6          | CMD, RRP in this case |
| 5             | 0x04          | argument ln in this case |
| 6             | 0x08          | argument cell/60 in this case |
| 7             | 0x00          | argument cell%60 in this case |


### Escape-value

If any start, stop or escape value is present in the Payload or XORSUM it is
replaced by an escape value (**0x1B**) and the bit-wise inverse of the actual value.
This is done by both clients and serves.

As example here is a server response for a temperature reading.

[ 3d 05 00 14 c3 ae 41 *1b* c2 3e ] (escape value in cursive)

The payload is [ 05 00 14 c3 ae 41 ] and its xorsum is 0x3D (server start value).
Thus an escape value and the inverse value (0xC2) is inserted instead.

Encryption
----------

It is possible to use encryption, the crypto is called Saphire and is currently not documented.





### Commands

| Opc | Hex| dec | Interpretation | Data | Anwser |
|-----|----|---|------------------|----------------|-----|
| SLV | 01 | 1 | Set logical var. | DLn Cell Value | Ok! |
| SLP | 2F | 47 | Set logic segment var. | DLn Seg Offs Value | Ok! |
| SXV | 02 | 2 | Set index var. | DLn Cell Value | Ok |
| SXP | B0 | 176 | Set index segment var. | DLn Seg Offs Value | Ok! |
| SRV | 04 | 4 | Set real var. | DLn Cell Value (4) | Ok |
| SRP | 32 | 50 | Set real segment var. | DLn Seg Offs Value (4) | Ok! |
| RLV | 86 | 134 | Read logical var. | DLn Cell | Value |
| RLP | B3 | 179 | Read logic segment var. | DLn Seg Offset | Value |
| RXV | 07 | 7 | Read index var. | DLn Cell | Value |
| RXP | 34 | 52 | Read index segment var. | DLn Seg Offset | Value |
| RRV | 89 | 137 | Read real var. | DLn Cell | Value (4) |
| RRP | B6 | 182 | Read real segment var. | DLn Seg Offset | Value (4) |
| READV | 10 | 16 | Read Vpac page. | DLn DPn | Data (n) |
