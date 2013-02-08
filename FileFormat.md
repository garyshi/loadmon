# General #
* No file headers
* Each file contains a list of load messages
* Numbers are in big-endians
* Timestamp: number of seconds from 2000-01-01 00:00:00 UTC
* Message version == 1

# Load Message #
* UINT32: local timestamp
* UINT16: message length
* message body

# Message Body #
* UINT8: message version (1)
* UINT16: monitor interval in seconds
* UINT32: source timestamp
* ProcLoad
  + FLOAT32: total/idle uptime
  + FLOAT32: 3 loadavgs
  + UINT32: number of all/running/iowait/zombie procs
* CPULoad
  + UINT8: number of cpu (cores)
  + for each cpu core:
      - UINT8: user/sys/iowait/idle rate to 0-255
* MemoryLoad (TODO: missing total)
  + UINT32: free/buffers/cached/dirty/active
  + UINT32: swap total/free/cached
* IOLoad
  + UINT8: number of devices ("sd?" in /proc/diskstats)
  + for each disk:
      - UINT8: device name length
      - BYTES: device name
      - UINT32: tps-read, tps-written, kbytes-read, kbytes-written
* NetworkLoad
  + UINT8: number of interface (non-alias devices in /proc/net/dev)
  + for each interface:
      - UINT8: interface name length
      - BYTES: interface name
      - UINT32: pkts-recv, pkts-sent, kbytes-recv, kbytes-sent
