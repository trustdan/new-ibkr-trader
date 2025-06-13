# Flow Journal - Day 0 - Planning Phase

## Morning Intention
- Energy level: 9/10
- Focus area: Understanding TWS API constraints and creating a rock-solid plan
- Vibe: Excited but analytical - need to get the architecture right

## Session Highlights
### Breakthroughs
- **Major realization**: ib-insync handles ALL the hard stuff! No need to reinvent threading, rate limiting, or reconnection
- **Event-driven is the way**: Fighting async patterns = pain. Embracing them = flow
- **The One Rule**: Never block the event loop - this changes everything about our Python service design

### Challenges
- Initial plan was too traditional (request/response thinking)
- Missed the importance of market data line limits
- Didn't fully appreciate ib-insync's elegance at first

### Code Snippets
```python
# This pattern will be EVERYWHERE
async def do_something():
    # Never use time.sleep()!
    await ib.sleep(1)  # Maintains event loop
    
    # Let ib-insync handle the complexity
    trade = await ib.placeOrderAsync(contract, order)
    # Events will fire automatically!
```

## API Learnings
- TWS requires manual login (no headless) - must document clearly
- Daily restart at 11:45 PM EST is non-negotiable - Watchdog handles it
- Rate limiting: 45 req/sec is safe (not 50) - but ib-insync throttles automatically
- Market data lines vary by subscription - must track usage
- Error 1100 = connection lost, Watchdog auto-recovers
- Order IDs are sacred - must track nextValidId

## Tomorrow's Flow
- Set up project structure with async-first mindset
- Create Docker environment that respects The One Rule
- Start experiments/ folder with ib-insync examples
- Test Watchdog recovery scenarios early

## Vibe Check
- Flow state achieved: Yes! The plan clicked into place
- Best working environment: Multiple monitors with docs open
- Commits made: 
  - "good vibes" - Initial positive energy
  - "first" - Starting the journey

---

## Quick Notes
- Remember: Don't fight the framework
- Monitoring from day 1 is crucial
- The GUI needs to show system health prominently
- Backpressure between Go and Python is CRITICAL
- Event handlers should be lightweight - no heavy processing
- Consider voice notes for complex async patterns
- Test with paper trading immediately, not later