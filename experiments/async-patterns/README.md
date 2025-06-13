# 🧪 Async Patterns Experiments

This directory is our playground for testing ib-insync patterns before integrating them into production code. Feel free to break things here!

## Purpose
- Test async/await patterns with ib-insync
- Experiment with event handling approaches
- Try different connection strategies
- Benchmark performance ideas
- Explore edge cases safely

## Structure
```
async-patterns/
├── connection_tests/     # Different ways to connect
├── event_experiments/    # Event handler patterns
├── rate_limit_tests/     # Testing API limits
├── market_data_tests/    # Subscription experiments
└── order_tests/          # Order execution trials
```

## Guidelines
1. **Break things freely** - This is the safe space
2. **Document discoveries** - Add notes about what worked/didn't
3. **Keep it messy** - Perfect code isn't the goal here
4. **Share insights** - Update flow journal with learnings

## Quick Test Template
```python
# experiment_name.py
"""
What I'm testing:
Expected outcome:
Actual result:
Learnings:
"""

import asyncio
from ib_insync import IB, util

async def experiment():
    # Your experimental code here
    pass

if __name__ == '__main__':
    util.run(experiment())
```

## Current Experiments
- [ ] Test watchdog reconnection patterns
- [ ] Measure event handler performance
- [ ] Try different subscription strategies
- [ ] Explore order preview workflows

Remember: If it works here, refactor it for production. If it doesn't, learn why!