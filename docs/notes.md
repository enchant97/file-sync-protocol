# Notes
## Main Deadlines
- Poster Presentation (06/01/2023)
- Project (05/05/2023)


## Goals
- Transfer files between client and server over a network
- Be reliable, must be able to handle connection loss by recovering lost data
  - Handle incomplete transfers without resending the whole file (send only what's missing)
- Minimal idle transfers, when nothing is happening minimal data is sent


## Related Projects
- FTP, SMB, rsync, etc
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

### Test Data
Planned data, that will be used to test against.

- Git repository objects or source code (~MBs, ~200 files)
- Collection of photos 32MP (~1GB, ~100 files)
- 4K Video (~5GB, 1 files)


## Tools For Scripts
### RAM Disk
```sh
mkdir /mnt/tmp

mount -t tmpfs -o size=4G tmp-disk /mnt/tmp

umount /mnt/tmp && rm -f /mnt/tmp
```

### SMB
- <https://unix.stackexchange.com/a/536145>

### Rsync
- <https://www.man7.org/linux/man-pages/man1/rsync.1.html>

### Toggle Interface
```sh
ip link set <interface> down/up
```
- [Source](https://www.2daygeek.com/enable-disable-up-down-nic-network-interface-port-linux/)

### DHCP Get New IP
```sh
# release ip
dhclient -v -r <interface>

# request new one
dhclient -v <interface>
```
- [Source](https://www.cyberciti.biz/faq/howto-linux-renew-dhcp-client-ip-address/)

### Get Current DateTime
```sh
date --rfc-3339 ns
# output: 2022-10-18 15:11:27.463650129+01:00
```

- [Source](https://man7.org/linux/man-pages/man1/date.1.html)

### Emulate Network Stuff
This can be done using tc & netem which is part of iproute2.

> Does not seem to work on my arch system, although does on ubuntu?

- add constant delay
- add varying delay
- add packet loss
- duplicate packets
- corrupt packets
- reorder packets

#### Sample Simulations
```sh
tc qdisc add dev <interface> root netem delay 250ms

tc qdisc add dev <interface> root netem delay 100ms 10ms

tc qdisc add dev <interface> root netem loss 1%

tc qdisc add dev <interface> root netem corrupt 0.2%

tc qdisc add dev <interface> root netem delay 10ms reorder 25% 50%

# if modifying use:
tc qdist change ...

# many more that are found with:
man netem
```

#### Remove All Simulations
```sh
tc qdisc del dev <interface> root netem
```

- [Source](https://srtlab.github.io/srt-cookbook/how-to-articles/using-netem-to-emulate-networks.html)

### Generate Random Data
```sh
# "urandom" as it's faster than /dev/random (does not need to be secure)
# "count" can adjust size in bytes

dd if=/dev/urandom of=random-data.bin bs=1 count=1024
```

- [Source](https://stackoverflow.com/a/1462909/8075455)
