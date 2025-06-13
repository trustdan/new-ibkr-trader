# IBKR Spread Automation - Project Manifesto

## Our Why

We believe trading should feel like conducting a symphony, not wrestling with spreadsheets. This project exists to transform the complex dance of options trading into an elegant, flowing experience where technology amplifies human intuition rather than replacing it.

## Core Values

### 1. Flow State is Sacred
- Code when energized, document when reflective
- Never force creativity - let it emerge
- Tools should disappear into the background
- Interruptions are the enemy of great work

### 2. Living Documentation
- Docs grow with code, not after
- Every decision has a story worth telling
- Comments explain why, not what
- Examples are better than explanations

### 3. Embrace the Event Stream
- The market is alive - our code should be too
- React don't poll, listen don't ask
- Let ib-insync guide our patterns
- Async all the way down

### 4. Human-Centric Automation
- Automate the mechanical, amplify the strategic
- Every click should feel intentional
- Errors should teach, not punish
- The trader remains in control

### 5. Pragmatic Excellence
- Ship working code, iterate to perfection
- Test what matters, mock what doesn't
- Monitor everything, alert on what's critical
- Performance is a feature

## Development Philosophy

### The Vibe Check
Before each session, ask:
- What excites me about today's work?
- What energy level am I bringing?
- Which tasks match my current state?
- How can I maintain flow?

### The One Rule (ib-insync edition)
**Never block the event loop.** Everything else is negotiable.

### Commit Messages as Stories
Each commit tells part of our journey. Make them worth reading.

### Experiments are Expected
The `experiments/` folder is our playground. Break things there, not in production.

## Technical Principles

### Async-First Architecture
- Events drive everything
- Callbacks are our friends
- Await the future, don't block on it
- Let the framework do the heavy lifting

### Monitoring as Meditation
- Dashboards are our window into the system's soul
- Metrics tell stories - listen to them
- Alert fatigue kills attention - be selective
- A healthy system hums

### The Three-Click Rule
Any trading action should be accessible within three clicks. Cognitive load is the enemy of good decisions.

## Success Looks Like

- A trader opens the app and smiles
- Complex strategies execute flawlessly
- The system recovers gracefully from failures
- New developers understand the code instantly
- TWS restarts don't cause panic
- Performance metrics show consistent sub-second responses
- The flow journal has more "breakthroughs" than "blockers"

## Remember

We're not just building a trading system - we're crafting an experience. Every line of code, every design decision, every documentation paragraph should contribute to a sense of flow, control, and confidence.

When in doubt:
1. Choose clarity over cleverness
2. Choose events over polling  
3. Choose monitoring over hoping
4. Choose flow over force

---

*"The best trading system is one that feels like an extension of the trader's mind, not a replacement for it."*

Let's build something remarkable. 🚀