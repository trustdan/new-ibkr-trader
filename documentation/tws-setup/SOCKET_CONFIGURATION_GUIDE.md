# Socket Configuration Deep Dive
## Optimizing TWS Socket Communication for High-Performance Trading

### Socket Architecture Overview

TWS uses a traditional TCP socket architecture with some quirks:

```
[Your App] <--TCP Socket--> [TWS Client] <--Proprietary--> [IB Servers]
           Port 7497/7496                    SSL/TLS
```

### Critical Socket Parameters

#### 1. TCP_NODELAY (Nagle's Algorithm)
**ALWAYS DISABLE** for trading applications:
```python
# Python
sock.setsockopt(socket.IPPROTO_TCP, socket.TCP_NODELAY, 1)

# Go
conn.SetNoDelay(true)
```
**Impact**: 40ms latency reduction on small messages

#### 2. Socket Buffer Sizes
Default buffers are insufficient for options scanning:

```python
# Python optimal settings
sock.setsockopt(socket.SOL_SOCKET, socket.SO_RCVBUF, 262144)  # 256KB receive
sock.setsockopt(socket.SOL_SOCKET, socket.SO_SNDBUF, 131072)  # 128KB send

# Go optimal settings
conn.SetReadBuffer(262144)
conn.SetWriteBuffer(131072)
```

#### 3. Keep-Alive Settings
Prevent connection drops during quiet periods:

```python
# Enable TCP keepalive
sock.setsockopt(socket.SOL_SOCKET, socket.SO_KEEPALIVE, 1)

# Windows-specific (values in milliseconds)
sock.ioctl(socket.SIO_KEEPALIVE_VALS, (1, 10000, 3000))
# Keepalive after 10 seconds, retry every 3 seconds
```

### Message Framing & Protocol

TWS uses a custom protocol with length-prefixed messages:

```
[4 bytes length][message data]
```

#### Reading Messages Correctly
```python
async def read_message(reader: asyncio.StreamReader) -> bytes:
    # Read length prefix
    length_bytes = await reader.readexactly(4)
    length = int.from_bytes(length_bytes, 'big')
    
    # Read exact message length
    message = await reader.readexactly(length)
    return message
```

### Connection Pooling Strategy

For high-frequency operations, use multiple connections:

```python
class ConnectionPool:
    def __init__(self, size: int = 5):
        self.connections = []
        self.client_ids = list(range(1, size + 1))
        
    async def initialize(self):
        for client_id in self.client_ids:
            conn = await self.create_connection(client_id)
            self.connections.append(conn)
    
    def get_connection(self, purpose: str) -> Connection:
        # Round-robin or purpose-based selection
        if purpose == "scanner":
            return self.connections[0]
        elif purpose == "market_data":
            return self.connections[1]
        elif purpose == "orders":
            return self.connections[2]
```

### Error Recovery Patterns

#### 1. Exponential Backoff Reconnection
```python
async def reconnect_with_backoff(self):
    backoff = 1
    max_backoff = 60
    
    while not self.connected:
        try:
            await self.connect()
            logger.info("Reconnected successfully")
            break
        except Exception as e:
            logger.error(f"Reconnection failed: {e}")
            await asyncio.sleep(backoff)
            backoff = min(backoff * 2, max_backoff)
```

#### 2. Socket Health Monitoring
```python
class SocketMonitor:
    def __init__(self, socket):
        self.socket = socket
        self.last_activity = time.time()
        self.ping_interval = 30
        
    async def monitor_loop(self):
        while True:
            await asyncio.sleep(self.ping_interval)
            if time.time() - self.last_activity > self.ping_interval:
                await self.send_ping()
    
    async def send_ping(self):
        # TWS doesn't have explicit ping, use lightweight request
        await self.request_current_time()
```

### Platform-Specific Optimizations

#### Windows
```python
# Disable delayed ACK
import ctypes
from ctypes import wintypes

# Windows TCP_NODELAY is more aggressive
sock.setsockopt(socket.IPPROTO_TCP, 12, 1)  # TCP_NODELAY variant
```

#### Linux
```python
# Enable TCP_QUICKACK for lower latency
sock.setsockopt(socket.IPPROTO_TCP, 12, 1)  # TCP_QUICKACK

# Increase priority
sock.setsockopt(socket.SOL_SOCKET, socket.SO_PRIORITY, 6)
```

### Diagnostic Tools

#### Socket Statistics
```python
def get_socket_stats(sock):
    """Get detailed socket statistics"""
    stats = {
        'recv_buffer': sock.getsockopt(socket.SOL_SOCKET, socket.SO_RCVBUF),
        'send_buffer': sock.getsockopt(socket.SOL_SOCKET, socket.SO_SNDBUF),
        'nodelay': sock.getsockopt(socket.IPPROTO_TCP, socket.TCP_NODELAY),
        'keepalive': sock.getsockopt(socket.SOL_SOCKET, socket.SO_KEEPALIVE),
    }
    return stats
```

#### Connection Testing Script
```python
async def test_socket_performance():
    """Test socket latency and throughput"""
    latencies = []
    
    for _ in range(100):
        start = time.time()
        await client.reqCurrentTime()
        latency = (time.time() - start) * 1000
        latencies.append(latency)
    
    print(f"Average latency: {sum(latencies)/len(latencies):.2f}ms")
    print(f"Min latency: {min(latencies):.2f}ms")
    print(f"Max latency: {max(latencies):.2f}ms")
```

### Common Socket Issues

1. **"Socket is already connected"**
   - Cause: Trying to reconnect without proper cleanup
   - Fix: Always close() before reconnecting

2. **"Connection reset by peer"**
   - Cause: TWS timeout or overload
   - Fix: Implement keepalive and monitor activity

3. **"Buffer overflow"**
   - Cause: Receiving data faster than processing
   - Fix: Increase buffer sizes and use async processing

4. **High latency spikes**
   - Cause: Nagle's algorithm or delayed ACK
   - Fix: Set TCP_NODELAY and platform-specific options

### Production Socket Checklist

- [ ] TCP_NODELAY enabled
- [ ] Buffer sizes optimized (256KB/128KB)
- [ ] Keep-alive configured
- [ ] Connection pool implemented
- [ ] Monitoring in place
- [ ] Error recovery tested
- [ ] Platform optimizations applied
- [ ] Diagnostic tools ready