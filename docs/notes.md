# Notes
## Related Projects
- FTP, SMB, etc
- SyncThing
- Cloud Storage (Google Drive, OneDrive, etc)


## Testing Ideas
- Capture Traffic In Wireshark
- See how app/protocol handles loosing connection, *will need to use a large file*
- See how much extra data is sent, that is not the test file
- How much data is sent during "handshake"
- Type of packets sent
- Record timestamps of each network change/interaction by automated script. So it can be compared with Wireshark output
- How well does it work with different amounts of traffic on network
- Limit bandwidth, using [PfSense](https://docs.netgate.com/pfsense/en/latest/trafficshaper/limiters.html)?
- Amount of inactive traffic

## Tools For Scripts

### Toggle Interface
```
ip link set <interface> down/up
```
- [Source](https://www.2daygeek.com/enable-disable-up-down-nic-network-interface-port-linux/)

### DHCP Get New IP
```
# release ip
dhclient -v -r <interface>

# request new one
dhclient -v <interface>
```
- [Source](https://www.cyberciti.biz/faq/howto-linux-renew-dhcp-client-ip-address/)

### Get Current DateTime
```
date --rfc-3339 ns
# output: 2022-10-18 15:11:27.463650129+01:00
```

- [Source](https://man7.org/linux/man-pages/man1/date.1.html)
