# 🌊 The IBKR Spread Automation Manifesto

## Our Development Philosophy

This project embraces **vibe coding** - a development approach that prioritizes flow state, developer happiness, and code that feels good to write and maintain.

## Core Principles

### 1. Flow State is Sacred 🧘
- **Batch similar tasks** to maintain context
- **Minimize interruptions** - gather requirements before diving deep
- **Trust your instincts** - if something feels wrong, it probably is
- **Celebrate small wins** - every working function is a victory

### 2. Living Documentation 📝
- **Document as you discover**, not after you forget
- **Code tells what, comments tell why**
- **Flow journals capture the journey**, not just the destination
- **Examples > explanations** - show, don't just tell

### 3. Async-First Architecture ⚡
- **Events over polling** - let the system tell us what's happening
- **Never block the event loop** - keep everything responsive
- **Embrace callbacks** - they're not scary when done right
- **Let ib-insync handle the complexity** - don't reinvent the wheel

### 4. Fail Gracefully, Recover Automatically 🛡️
- **Expect disconnections** - TWS restarts daily, plan for it
- **Rate limits are real** - respect them or face the consequences
- **Watchdogs save lives** - automatic recovery is non-negotiable
- **Log everything interesting** - future you will thank present you

### 5. Developer Experience Matters 🎨
- **Beautiful code is maintainable code**
- **Meaningful variable names** > clever abbreviations
- **Consistent patterns** across the codebase
- **If it's not fun to work on, refactor it**

## Technical Philosophy

### Event-Driven Everything
```python
# Not this:
while True:
    check_for_updates()
    time.sleep(1)

# But this:
ib.pendingTickersEvent += handle_updates
```

### Composition Over Inheritance
Small, focused components that work together beautifully.

### Metrics From Day One
You can't optimize what you don't measure. Prometheus + Grafana = ❤️

### Testing in Layers
1. Experiments folder for wild ideas
2. Unit tests for critical logic
3. Integration tests for workflows
4. Manual testing for the "feel"

## The Vibe Test

Before committing code, ask yourself:
1. **Does this spark joy?** Would I be happy maintaining this in 6 months?
2. **Is it obvious?** Could another developer understand it quickly?
3. **Does it flow?** Is the logic natural and easy to follow?
4. **Is it resilient?** Will it handle the unexpected gracefully?

## Our Promise

We promise to:
- **Respect the trader's capital** - no careless mistakes
- **Respect the API** - no pacing violations or resource abuse
- **Respect future developers** - leave code better than we found it
- **Respect the process** - trust in vibe coding principles

## The Journey

This isn't just about building a trading system. It's about crafting something we're proud of, learning deeply about markets and async programming, and enjoying every step of the journey.

When in doubt:
- Check the vibe ✨
- Trust the process 🌊
- Ship working code 🚀
- Document the adventure 📖

---

*"Code with intention, trade with precision, and always maintain the vibe."*

## Quick Vibe Checks

Feeling stuck? Try these:
1. **Take a walk** - solutions come when you stop forcing them
2. **Switch to experiments/** - play without pressure
3. **Update the flow journal** - writing clarifies thinking
4. **Review this manifesto** - remember why we're here
5. **Celebrate progress** - you're building something amazing!

Remember: The best code is written in flow state. Protect it, nurture it, and let it guide you. 🌊✨