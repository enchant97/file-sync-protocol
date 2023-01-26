# Prototype 2
This prototype I will alter the first prototype to improve handling of when lots of packets go missing during a file transfer. It will only include the features listed below:

- Very basic error correction
- Dummy Handshake
- Send one real file from client to server
  - Handle resend of missing file packets
  - Grouped packets
- Customisable MTU size for testing
- Dummy connection close


## Usage
Same as proto-1


## Discovered Issues
TBD.


## Structure
Same as proto-1


## Client File Push

```mermaid
sequenceDiagram
    Note over Client,Server: Init Connection
    Client->>+Server: SYN
    Server-->>-Client: ACK
    Note over Client,Server: Req Push
    Client ->>+ Server: REQ
    Server -->>- Client: ACK
    Note over Client,Server: Send Groups
    loop
        Note over Client,Server: Send Next Group
        break No More Groups
            Note over Server: EOF
            Client->>+ Server: REQ
            Server -->>- Client: ACK
        end
        Note over Server: Group ID
        Client->>+ Server: REQ
        Server -->>- Client: ACK
        loop Send Chunks
            break No More Chunks
                Client ->> Server: REQ
            end
            Note over Client,Server: Send N Chunks
            loop Until ACK
                loop Send Next Chunk
                    break No More Chunks
                        Note over Server: Expected Chunk ID's
                        Client ->> Server: REQ
                    end
                    Client ->> Server: PSH
                end
                alt Received All
                    Server -->> Client: ACK
                else Resend Chunk(s)
                    Note over Client: Missing Chunk ID's
                    Server -->> Client: REQ
                end
            end
        end
    end
    Note over Client,Server: Close Connection
    Client->>+ Server: FIN
    Server -->>- Client: ACK

```
