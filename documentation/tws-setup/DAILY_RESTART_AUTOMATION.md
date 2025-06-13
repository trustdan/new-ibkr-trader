# TWS Daily Restart Automation Guide
## Handling the Inevitable with Grace

> TWS will restart daily. This is not optional. Plan for it or suffer.

### Restart Schedule (Eastern Time)

```python
# TWS enforced restart windows
RESTART_SCHEDULE = {
    'weekday': {
        'primary': ('23:45', '00:05'),     # 11:45 PM - 12:05 AM
        'secondary': ('05:45', '06:00'),   # 5:45 AM - 6:00 AM (maintenance)
    },
    'sunday': {
        'primary': ('23:45', '00:05'),     # 11:45 PM - 12:05 AM
        'maintenance': ('10:00', '14:00'),  # 10:00 AM - 2:00 PM (extended)
    }
}
```

### Automated Restart Handler Implementation

```python
import asyncio
import datetime
from enum import Enum
from typing import Optional, Callable

class TWS RestartState(Enum):
    NORMAL = "normal"
    PRE_RESTART = "pre_restart"
    RESTARTING = "restarting"
    RECONNECTING = "reconnecting"
    VALIDATING = "validating"

class TWS RestartManager:
    def __init__(self, tws_client, state_manager):
        self.client = tws_client
        self.state_manager = state_manager
        self.state = TWS RestartState.NORMAL
        self.restart_callbacks = []
        
    async def start_monitoring(self):
        """Start the restart monitoring loop"""
        asyncio.create_task(self._monitor_loop())
        
    async def _monitor_loop(self):
        """Main monitoring loop"""
        while True:
            try:
                current_time = datetime.datetime.now()
                
                # Check if we're approaching restart window
                if self._is_pre_restart_window(current_time):
                    await self._handle_pre_restart()
                    
                # Check if we're in restart window
                elif self._is_in_restart_window(current_time):
                    await self._handle_restart_window()
                    
                # Normal operation
                else:
                    self.state = TWS RestartState.NORMAL
                    
            except Exception as e:
                logger.error(f"Restart monitor error: {e}")
                
            await asyncio.sleep(30)  # Check every 30 seconds
    
    def _is_pre_restart_window(self, current_time: datetime.datetime) -> bool:
        """Check if we're 5 minutes before restart"""
        for window_start, _ in self._get_restart_windows():
            start_time = self._parse_time(window_start)
            pre_window = start_time - datetime.timedelta(minutes=5)
            
            if pre_window <= current_time.time() <= start_time:
                return True
        return False
    
    async def _handle_pre_restart(self):
        """Prepare for restart"""
        if self.state != TWS RestartState.PRE_RESTART:
            logger.warning("Entering pre-restart phase")
            self.state = TWS RestartState.PRE_RESTART
            
            # 1. Cancel all pending orders
            await self._cancel_all_pending_orders()
            
            # 2. Close non-essential connections
            await self._close_non_essential_connections()
            
            # 3. Save current state
            await self._save_trading_state()
            
            # 4. Notify callbacks
            await self._notify_callbacks('pre_restart')
    
    async def _handle_restart_window(self):
        """Handle actual restart"""
        if self.state != TWS RestartState.RESTARTING:
            logger.warning("TWS restart window active")
            self.state = TWS RestartState.RESTARTING
            
            # 1. Disconnect cleanly
            await self.client.disconnect()
            
            # 2. Wait for TWS to restart
            await self._wait_for_tws_restart()
            
            # 3. Reconnect with validation
            await self._reconnect_and_validate()
            
            # 4. Restore state
            await self._restore_trading_state()
            
            # 5. Resume operations
            self.state = TWS RestartState.NORMAL
            await self._notify_callbacks('restart_complete')
```

### State Preservation System

```python
class TradingStateManager:
    """Preserve trading state across restarts"""
    
    def __init__(self, state_file: str = "trading_state.json"):
        self.state_file = state_file
        self.state = {
            'open_orders': {},
            'positions': {},
            'subscriptions': [],
            'scanner_filters': {},
            'last_order_id': 0,
            'timestamp': None
        }
    
    async def save_state(self, client):
        """Save current trading state"""
        logger.info("Saving trading state before restart")
        
        # Get open orders
        self.state['open_orders'] = await self._get_open_orders(client)
        
        # Get positions
        self.state['positions'] = await self._get_positions(client)
        
        # Get active subscriptions
        self.state['subscriptions'] = client.get_active_subscriptions()
        
        # Save to file
        self.state['timestamp'] = datetime.datetime.now().isoformat()
        
        async with aiofiles.open(self.state_file, 'w') as f:
            await f.write(json.dumps(self.state, indent=2))
            
        logger.info(f"Saved state: {len(self.state['open_orders'])} orders, "
                   f"{len(self.state['positions'])} positions")
    
    async def restore_state(self, client):
        """Restore trading state after restart"""
        if not os.path.exists(self.state_file):
            logger.warning("No state file found")
            return
            
        async with aiofiles.open(self.state_file, 'r') as f:
            self.state = json.loads(await f.read())
            
        age = datetime.datetime.now() - datetime.datetime.fromisoformat(
            self.state['timestamp']
        )
        
        if age.total_seconds() > 3600:  # 1 hour
            logger.warning("State file too old, not restoring")
            return
            
        logger.info(f"Restoring state from {age.total_seconds():.0f} seconds ago")
        
        # Restore subscriptions
        for sub in self.state['subscriptions']:
            await client.restore_subscription(sub)
            
        # Verify positions match
        await self._verify_positions(client)
        
        # Re-monitor open orders
        await self._restore_order_monitoring(client)
```

### Windows Task Scheduler Configuration

```xml
<?xml version="1.0" encoding="UTF-16"?>
<Task version="1.4" xmlns="http://schemas.microsoft.com/windows/2004/02/mit/task">
  <RegistrationInfo>
    <Description>Manages TWS daily restart and recovery</Description>
  </RegistrationInfo>
  <Triggers>
    <!-- Pre-restart trigger - 11:40 PM -->
    <CalendarTrigger>
      <StartBoundary>2024-01-01T23:40:00</StartBoundary>
      <Repetition>
        <Interval>P1D</Interval>
      </Repetition>
    </CalendarTrigger>
    <!-- Post-restart trigger - 12:10 AM -->
    <CalendarTrigger>
      <StartBoundary>2024-01-01T00:10:00</StartBoundary>
      <Repetition>
        <Interval>P1D</Interval>
      </Repetition>
    </CalendarTrigger>
  </Triggers>
  <Actions>
    <Exec>
      <Command>C:\TradingSystem\scripts\handle_tws_restart.bat</Command>
    </Exec>
  </Actions>
  <Settings>
    <Priority>0</Priority>
    <RunOnlyIfNetworkAvailable>true</RunOnlyIfNetworkAvailable>
  </Settings>
</Task>
```

### Restart Handling Script

```batch
@echo off
REM handle_tws_restart.bat

echo [%date% %time%] Starting TWS restart handler >> C:\TradingSystem\logs\restart.log

REM Check if it's pre-restart time
for /f "tokens=1-2 delims=:" %%a in ('time /t') do (
    if %%a==11 if %%b GEQ 40 (
        echo [%date% %time%] Pre-restart phase >> C:\TradingSystem\logs\restart.log
        python C:\TradingSystem\scripts\pre_restart.py
        goto :end
    )
)

REM Check if it's post-restart time
for /f "tokens=1-2 delims=:" %%a in ('time /t') do (
    if %%a==12 if %%b LEQ 15 (
        echo [%date% %time%] Post-restart phase >> C:\TradingSystem\logs\restart.log
        
        REM Kill any hanging TWS process
        taskkill /F /IM tws.exe 2>nul
        
        REM Wait 30 seconds
        timeout /t 30 /nobreak
        
        REM Start TWS with proper memory settings
        start "" "C:\Jts\tws.exe" -J-Xmx4096m
        
        REM Wait for TWS to fully start
        timeout /t 60 /nobreak
        
        REM Start trading system
        python C:\TradingSystem\scripts\post_restart.py
    )
)

:end
echo [%date% %time%] Restart handler complete >> C:\TradingSystem\logs\restart.log
```

### Graceful Shutdown Procedures

```python
class GracefulShutdown:
    """Handle graceful shutdown before restart"""
    
    def __init__(self, client, timeout: int = 120):
        self.client = client
        self.timeout = timeout
        self.shutdown_tasks = []
        
    async def initiate_shutdown(self):
        """Start graceful shutdown sequence"""
        logger.warning("Initiating graceful shutdown for TWS restart")
        
        # Create shutdown tasks
        self.shutdown_tasks = [
            self._cancel_orders(),
            self._close_scanners(),
            self._save_state(),
            self._notify_users(),
            self._close_positions_if_needed()
        ]
        
        # Execute with timeout
        try:
            await asyncio.wait_for(
                asyncio.gather(*self.shutdown_tasks),
                timeout=self.timeout
            )
            logger.info("Graceful shutdown completed")
        except asyncio.TimeoutError:
            logger.error("Shutdown timeout - forcing disconnect")
            await self.client.disconnect()
    
    async def _cancel_orders(self):
        """Cancel all open orders"""
        open_orders = await self.client.reqOpenOrders()
        
        for order in open_orders:
            try:
                await self.client.cancelOrder(order.orderId)
                logger.info(f"Cancelled order {order.orderId}")
            except Exception as e:
                logger.error(f"Failed to cancel order {order.orderId}: {e}")
    
    async def _close_scanners(self):
        """Close all active scanners"""
        for scanner_id in list(self.client.active_scanners):
            await self.client.cancelScannerSubscription(scanner_id)
```

### Post-Restart Validation

```python
class PostRestartValidator:
    """Validate system state after restart"""
    
    async def validate_system(self, client) -> bool:
        """Run all validation checks"""
        checks = [
            self._check_connection(),
            self._check_account_data(),
            self._check_order_id_sequence(),
            self._check_market_data_flow(),
            self._check_historical_data_access()
        ]
        
        results = await asyncio.gather(*checks, return_exceptions=True)
        
        failed = [i for i, r in enumerate(results) if isinstance(r, Exception)]
        if failed:
            logger.error(f"Validation failed for checks: {failed}")
            return False
            
        logger.info("All post-restart validations passed")
        return True
    
    async def _check_connection(self):
        """Verify connection is stable"""
        for _ in range(5):
            if not client.isConnected():
                raise Exception("Connection not stable")
            await asyncio.sleep(1)
        return True
```

### Monitoring Dashboard Integration

```python
# Prometheus metrics for restart monitoring
restart_metrics = {
    'tws_restart_count': Counter('tws_restart_count', 'Number of TWS restarts'),
    'tws_restart_duration': Histogram('tws_restart_duration_seconds', 
                                     'Time taken for restart recovery'),
    'tws_pre_restart_orders_cancelled': Counter('tws_pre_restart_orders_cancelled',
                                               'Orders cancelled before restart'),
    'tws_state_restoration_success': Counter('tws_state_restoration_success',
                                           'Successful state restorations')
}
```

### Best Practices Checklist

- [ ] Implement pre-restart warning system (5 minutes before)
- [ ] Save all trading state before restart
- [ ] Cancel all open orders gracefully
- [ ] Implement exponential backoff for reconnection
- [ ] Validate system state after restart
- [ ] Monitor restart patterns for anomalies
- [ ] Test restart handling in paper trading
- [ ] Document any custom restart times
- [ ] Alert operators of restart events
- [ ] Log all restart activities for audit