# rsync
- <https://rsync.samba.org>
- Uses TCP
- Connection is established via single initialisation message (exchange version number & supported hash methods) (as well as TCP)
- Three message types (Command, Query and Data)
- Uses bytes for message contents
- Splits files into chunks, and computes checksums for each
- Implements pipelining to reduce latency
- only 1 round trip is required for a file sync
- Transfer A->B
    1. B splits file into non-overlapping fixed-sized blocks
    2. B calculates checksums for each block (both rolling and md4)
    3. Sends checksums to A
    4. A searches through blocks to find matching checksums
    5. A sends B instructions on how to construct copy of A. Each instruction will be either reference of a block in B or raw data
- uses a "delta-transfer algorithm"
