# Experiments Folder 🧪

This is our playground - a safe space to try wild ideas, test API patterns, and break things without fear.

## Purpose

- **Test TWS API behaviors** - How does it really work?
- **Try new patterns** - Event handling, async flows, etc.
- **Performance experiments** - How fast can we go?
- **Proof of concepts** - Will this crazy idea work?
- **Learning exercises** - Understanding by doing

## Rules

1. **Nothing here goes to production** (without proper migration)
2. **Break things freely** - That's what this is for
3. **Document insights** - What did you learn?
4. **Share discoveries** - Update IDEAS.md with findings
5. **Keep it fun** - If it's not exciting, why do it?

## Structure

```
experiments/
├── README.md           # You are here
├── sandbox/           # Quick one-off tests
├── async-patterns/    # Testing async/await approaches
├── tws-behaviors/     # Understanding TWS quirks
├── performance/       # Speed tests and optimizations
├── monitoring/        # Metrics experiments
└── archived/          # Old experiments (still valuable!)
```

## Current Experiments

### sandbox/
Quick tests that don't need their own folder

### connection-test/
Testing various connection patterns with TWS

## How to Experiment

1. Create a new folder or use sandbox/
2. Write messy code - perfectionism kills creativity
3. Document what you're trying in a README
4. Run it, break it, learn from it
5. If it works, consider graduating it to src/
6. If it doesn't, document why (equally valuable!)

## Useful Patterns Found So Far

### Async Connection Test
```python
# This pattern works well for testing connections
from ib_insync import IB, util

async def test_connection():
    ib = IB()
    await ib.connectAsync('localhost', 7497)
    print(f"Connected: {ib.isConnected()}")
    ib.disconnect()

util.run(test_connection())
```

### Event Monitoring
```python
# See ALL events for debugging
def log_all_events(ib):
    for attr in dir(ib):
        if attr.endswith('Event'):
            event = getattr(ib, attr)
            event += lambda *args: print(f"{attr}: {args}")
```

## Failed Experiments (Learning Opportunities)

### ❌ Threading Instead of Async
- **What**: Tried using threads for parallel requests
- **Result**: Fought with ib-insync's event loop
- **Learning**: Embrace async all the way

### ❌ Polling for Updates  
- **What**: Regular polling for position updates
- **Result**: Inefficient and missed updates
- **Learning**: Events are always better

## Remember

> "An experiment is never a failure. It always demonstrates something." - Buckminster Fuller

The best features often start as "what if..." questions in this folder.

Happy experimenting! 🚀