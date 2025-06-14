# TWS Quick Setup for Integration Testing

## 🎯 Goal
Get TWS running for Day 6 afternoon integration testing within 15-20 minutes.

## 📋 Quick Checklist

### Step 1: TWS Installation (if needed)
- [ ] Go to [Interactive Brokers](https://www.interactivebrokers.com/en/trading/tws.php)
- [ ] Download "TWS Latest" for Windows
- [ ] Run installer (default settings are fine)
- [ ] Create paper trading account if you don't have one

### Step 2: TWS Configuration (Critical!)
- [ ] Launch TWS
- [ ] Log in with paper trading credentials
- [ ] Go to **File → Global Configuration → API → Settings**
- [ ] **✅ Enable "ActiveX and Socket Clients"**
- [ ] **❌ Disable "Read-Only API"**
- [ ] Set **Socket port** to `7497` (paper trading)
- [ ] **Add `127.0.0.1` to Trusted IPs** (if not present)
- [ ] Set **Master API client ID** to `0`
- [ ] Click **OK** and **restart TWS**

### Step 3: Verification
- [ ] TWS is running and logged in
- [ ] Port 7497 is active: `netstat -an | findstr "7497"`
- [ ] Run our test: `python src/python/test_connection.py`

## 🚨 Common Issues

### "Connection refused"
- TWS not running or not logged in
- Socket client not enabled
- Wrong port (7497 for paper, 7496 for live)

### "Read-only API" error
- Must disable "Read-Only API" in configuration
- Restart TWS after changing this setting

### "Authentication failed"
- Check paper trading credentials
- Ensure you're using the correct login URL

## ⚡ Fast Track (5 minutes)
If you already have TWS:

1. **Start TWS** → Log in with paper account
2. **File → Global Configuration → API → Settings**
3. **Enable Socket, Disable Read-Only, Port 7497**
4. **Restart TWS**
5. **Test**: `python src/python/test_connection.py`

## 🎯 Success Criteria
When setup is complete, you should see:
```
🔌 Attempting to connect to TWS...
✅ Connected successfully!
📅 Server time: [current time]
📊 Managed accounts: ['DU#######']
🔍 Connection info:
   - Host: host.docker.internal
   - Port: 7497
   - Client ID: 999
👋 Disconnected from TWS
```

## 📞 Next Steps After Setup
1. Run integration tests: `pytest tests/python/integration/ -m integration -v`
2. Continue with Day 6 afternoon development
3. Test Watchdog functionality
4. Validate trading operations

---

*Quick setup guide for maintaining Day 6 afternoon flow state!* ⚡ 