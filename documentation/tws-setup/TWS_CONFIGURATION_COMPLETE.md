# TWS Configuration Complete Guide
## Critical Setup Requirements for Production Trading

> ⚠️ **CRITICAL**: This guide contains battle-tested configuration requirements. Ignoring these settings WILL cause production failures.

## Table of Contents
1. [Memory Configuration](#memory-configuration)
2. [Socket Configuration](#socket-configuration)
3. [Port Configuration](#port-configuration)
4. [API Settings](#api-settings)
5. [Daily Restart Handling](#daily-restart-handling)
6. [Common Pitfalls & Solutions](#common-pitfalls--solutions)
7. [Production Checklist](#production-checklist)

---

## Memory Configuration

### Minimum Requirements
```
TWS Minimum: 4GB RAM
Recommended: 8GB RAM
Production: 16GB RAM (for heavy scanning)
```

### JVM Memory Settings
1. Navigate to: `Configure > Settings > Memory`
2. Set the following:
   ```
   Initial heap size: 2048 MB
   Maximum heap size: 4096 MB
   ```

### Windows Memory Configuration
```batch
# Add to TWS startup script
set JAVA_OPTS=-Xms2048m -Xmx4096m -XX:+UseG1GC
```

> ⚠️ **WARNING**: Insufficient memory causes:
> - Random disconnections
> - Slow market data updates
> - Order rejection during high volume
> - Complete TWS freezes

---

## Socket Configuration

### Critical Socket Settings
```
File: tws.xml (in TWS settings directory)

<socket_client>
    <reconnect_count>100</reconnect_count>
    <reconnect_interval>10</reconnect_interval>
    <socket_timeout>30</socket_timeout>
    <keep_alive>true</keep_alive>
    <tcp_no_delay>true</tcp_no_delay>
</socket_client>
```

### API Socket Configuration
1. Navigate to: `Configure > Settings > API > Settings`
2. Enable:
   - [✓] Enable ActiveX and Socket Clients
   - [✓] Socket Port: 7497 (paper) / 7496 (live)
   - [✓] Allow connections from localhost only (security)
   - [✓] Create API message log file

### Socket Buffer Settings
```python
# In your Python client
self.client.socket.setsockopt(socket.SOL_SOCKET, socket.SO_RCVBUF, 65536)
self.client.socket.setsockopt(socket.SOL_SOCKET, socket.SO_SNDBUF, 65536)
self.client.socket.setsockopt(socket.IPPROTO_TCP, socket.TCP_NODELAY, 1)
```

> ⚠️ **CRITICAL**: Default socket buffers are too small for options scanning!

---

## Port Configuration

### Standard Port Assignments
```
Paper Trading: 7497
Live Trading: 7496
Gateway (if used): 4001 (live) / 4002 (paper)
```

### Firewall Configuration
```powershell
# Windows Firewall Rules (Run as Administrator)
netsh advfirewall firewall add rule name="TWS API Paper" dir=in action=allow protocol=TCP localport=7497
netsh advfirewall firewall add rule name="TWS API Live" dir=in action=allow protocol=TCP localport=7496
netsh advfirewall firewall add rule name="TWS API Out" dir=out action=allow program="C:\Jts\tws.exe"
```

### Docker Port Mapping
```yaml
# docker-compose.yml
services:
  python-service:
    ports:
      - "127.0.0.1:7497:7497"  # Secure localhost-only binding
    environment:
      - TWS_PORT=7497
      - TWS_HOST=host.docker.internal  # For Windows Docker
```

---

## API Settings

### Essential API Configuration
1. **Master Client ID**:
   - Reserved ID: 0 (for order management)
   - Scanner IDs: 1-10
   - Market Data IDs: 11-50
   - Order IDs: 51-100

2. **Request Throttling**:
   ```
   Max requests/second: 50 (documented limit: 60)
   Concurrent requests: 100
   Market data lines: Depends on subscription
   ```

3. **API Precautions**:
   ```
   Configure > Settings > API > Precautions
   
   [✓] Bypass Order Precautions for API Orders
   [ ] Bypass Bond warning for bonds
   [✓] Bypass negative yield confirmation
   [ ] Bypass Called Bond warning
   [✓ Bypass "same action pair trade" warning
   ```

### Order ID Management
```python
# CRITICAL: Always track next valid ID
class OrderIdManager:
    def __init__(self, initial_id: int):
        self._next_id = initial_id
        self._id_lock = threading.Lock()
    
    def get_next_id(self) -> int:
        with self._id_lock:
            current_id = self._next_id
            self._next_id += 1
            return current_id
```

> ⚠️ **NEVER** reuse order IDs - TWS will reject or cause undefined behavior!

---

## Daily Restart Handling

### Automatic Restart Schedule
TWS forces daily restarts. Plan for it:

```python
# Restart windows (EST/EDT)
RESTART_WINDOWS = [
    (datetime.time(23, 45), datetime.time(0, 15)),  # Midnight
    (datetime.time(5, 45), datetime.time(6, 15)),   # Morning maintenance
]

def is_in_restart_window() -> bool:
    now = datetime.datetime.now().time()
    for start, end in RESTART_WINDOWS:
        if start <= now <= end:
            return True
    return False
```

### Graceful Restart Handler
```python
async def handle_tws_restart(self):
    """Handle daily TWS restart gracefully"""
    logger.warning("TWS restart detected - entering safe mode")
    
    # 1. Cancel all pending orders
    await self.cancel_all_orders()
    
    # 2. Save state
    await self.save_trading_state()
    
    # 3. Disconnect cleanly
    await self.disconnect()
    
    # 4. Wait for restart window
    await self.wait_for_restart_window()
    
    # 5. Reconnect with exponential backoff
    await self.reconnect_with_backoff()
    
    # 6. Restore state
    await self.restore_trading_state()
```

### Windows Task Scheduler Setup
```xml
<!-- TWS Auto-restart task -->
<Task>
    <Triggers>
        <CalendarTrigger>
            <StartBoundary>2024-01-01T06:30:00</StartBoundary>
            <Repetition>
                <Interval>P1D</Interval>
            </Repetition>
        </CalendarTrigger>
    </Triggers>
    <Actions>
        <Exec>
            <Command>C:\Scripts\restart_tws.bat</Command>
        </Exec>
    </Actions>
</Task>
```

---

## Common Pitfalls & Solutions

### 1. "Socket Connection Broken" Errors
**Cause**: Network interruption or TWS overload
**Solution**:
```python
# Implement automatic reconnection
@retry(
    stop=stop_after_attempt(10),
    wait=wait_exponential(multiplier=1, min=4, max=60),
    retry=retry_if_exception_type(ConnectionError)
)
async def connect_with_retry(self):
    await self.connect()
```

### 2. "No Security Definition Found"
**Cause**: Requesting data before contract details are loaded
**Solution**:
```python
# Always verify contract first
contract_details = await self.req_contract_details(contract)
if not contract_details:
    raise ValueError(f"Invalid contract: {contract}")
```

### 3. "Max Rate of Messages Exceeded"
**Cause**: Exceeding 50 messages/second limit
**Solution**: Implement request queuing with rate limiting

### 4. "Invalid Order ID"
**Cause**: Reusing or skipping order IDs
**Solution**: Always use sequential IDs from `nextValidId`

### 5. Memory Leaks During Long Sessions
**Cause**: Not clearing market data subscriptions
**Solution**:
```python
# Implement subscription manager
class SubscriptionManager:
    def __init__(self, max_subscriptions=100):
        self._active = {}
        self._max = max_subscriptions
    
    async def cleanup_old_subscriptions(self):
        if len(self._active) > self._max * 0.9:
            # Remove oldest 20%
            await self._remove_oldest(int(self._max * 0.2))
```

---

## Production Checklist

### Pre-Launch Verification
- [ ] TWS memory set to minimum 4GB
- [ ] Socket buffers increased
- [ ] Firewall rules configured
- [ ] API precautions disabled for automation
- [ ] Order ID manager initialized
- [ ] Restart handler implemented
- [ ] Error recovery in place
- [ ] Monitoring configured

### Daily Operations
- [ ] Check TWS memory usage (should be <75%)
- [ ] Verify socket connections active
- [ ] Monitor order ID sequence
- [ ] Check for restart warnings
- [ ] Review error logs
- [ ] Validate market data flow

### Weekly Maintenance
- [ ] Clear TWS cache
- [ ] Review and archive logs
- [ ] Update configuration backups
- [ ] Test failover procedures
- [ ] Verify all integrations

---

## Critical Commands Reference

### Check TWS Status
```bash
# Windows
netstat -an | findstr "7497"
tasklist | findstr "tws.exe"

# Memory usage
wmic process where name="tws.exe" get WorkingSetSize,PeakWorkingSetSize
```

### Force Restart TWS
```batch
@echo off
echo Stopping TWS...
taskkill /F /IM tws.exe
timeout /t 5
echo Starting TWS...
start "" "C:\Jts\tws.exe" -J-Xmx4096m
```

### Monitor Connections
```python
# In your monitoring service
async def check_tws_health():
    checks = {
        'connected': self.client.isConnected(),
        'memory_mb': get_tws_memory_usage() / 1024 / 1024,
        'socket_alive': await self.ping_socket(),
        'order_id_valid': self.next_order_id > 0,
        'subscriptions': len(self.active_subscriptions)
    }
    return checks
```

---

> 📌 **Remember**: TWS is powerful but quirky. Respect its limitations, implement proper error handling, and ALWAYS test in paper trading first!