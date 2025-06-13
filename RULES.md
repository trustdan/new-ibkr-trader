# RULES.md - Session Start Reference

## 🎯 Primary Directive
Build an automated vertical spread options trading system for Interactive Brokers following vibe coding principles - maintaining flow state while creating robust, production-ready code.

## 📋 Quick Reference Checklist

### At Session Start:
1. ✅ Read CLAUDE.md for project-specific instructions
2. ✅ Check TodoRead for current task status
3. ✅ Review ROADMAP.md current phase
4. ✅ Scan IDEAS.md for inspiration
5. ✅ Note any recent commits in flow_journal/

## 🌊 Vibe Coding Principles

### Core Values:
- **Flow State Preservation**: Minimize interruptions, batch similar tasks
- **Living Documentation**: Update docs as code evolves, not after
- **Intuitive Organization**: Structure that feels natural, not imposed
- **Creative Exploration**: Maintain experiments/ folder for safe exploration
- **Momentum Tracking**: Use TodoWrite proactively to track progress

### Development Rhythm:
1. **Start with Intent**: What's the vibe? What excites you about this session?
2. **Match Energy to Task**: Complex work when fresh, refactoring when tired
3. **Document in Flow**: Quick notes, voice memos, screenshots - polish later
4. **Commit the Story**: Descriptive commits that tell the development narrative
5. **End with Reflection**: Update flow journal, capture insights

## 🏗️ Project Structure Guidelines

### Directory Organization:
```
/src/              # Clean, production code only
/experiments/      # Playground for ideas
/docs/            # Living documentation
/.vibe/           # Flow logs, snippets, inspiration
/flow_journal/    # Session notes and breakthroughs
```

### File Naming Vibes:
- Descriptive over abbreviated: `scanner_optimization/` not `opt/`
- Progress indicators: `v1_basic/`, `v2_enhanced/`
- Emotional context where helpful: `tricky_rate_limiting/`

## ⚡ IBKR Integration Constraints

### Must Remember:
- TWS connection on 127.0.0.1:7497 (paper) or 7496 (live)
- Rate limit: 45 req/sec safe threshold (50 max)
- Daily TWS restart handling required
- Manual authentication only (no headless)
- EReader thread required for async messages
- Memory allocation: 4GB for TWS

### Error Codes to Handle:
- Error 100: Pacing violation
- Error 502: TWS not running
- Error 1100: Connectivity lost
- Error 507: Bad message/Socket EOF

## 🎭 Session Management

### TodoWrite Usage:
- Create todos for complex multi-step tasks
- Mark in_progress BEFORE starting work
- Complete immediately after finishing
- One in_progress task at a time
- Skip for trivial single-step tasks

### Documentation Flow:
- ADRs for significant technical decisions
- Update CHANGELOG.md after feature completion
- Keep IDEAS.md for future possibilities
- Flow journal for daily insights

## 🚀 Code Quality Standards

### Before Committing:
1. Run linting and type checking
2. Ensure tests pass
3. Update relevant documentation
4. Write descriptive commit message
5. Check for hardcoded values or secrets

### API Best Practices:
- Cache contract details to reduce lookups
- Batch similar requests
- Implement exponential backoff for retries
- Use whatIfOrder() before placing trades
- Monitor order status via callbacks

## 🖥️ Windows Environment Breakpoint

### STOP and Switch to Windows When:
- **TWS Connection Required**: Any actual TWS API connection testing
- **Windows GUI Development**: Building the Svelte/Electron Windows app
- **Windows-specific Features**: Registry access, Windows notifications, etc.
- **Production Testing**: Final integration testing on target platform

### Current Development Status:
- **Currently on**: Kali Linux (WSL2)
- **Can continue on Linux**: Documentation, architecture, Docker setup, Go scanner
- **Must switch to Windows**: TWS testing, GUI development, production builds

### Breakpoint Protocol:
1. 🛑 **STOP** when reaching Windows-required components
2. 📝 Document exactly where you stopped in flow journal
3. 💾 Commit all work with clear handoff notes
4. 🪟 Continue development on Windows machine with TWS installed
5. 🔄 Sync back to Linux for non-Windows work

## 🎨 Creative Guidelines

### Maintain The Vibe:
- If stuck, switch to experiments/ folder
- Use voice notes for complex thoughts
- Take breaks to preserve flow state
- Trust intuition on code organization
- Celebrate small wins in flow journal

### Balance Structure & Flexibility:
**Keep Rigid:**
- Core project structure
- API rate limiting
- Error handling
- Security practices

**Keep Flexible:**
- Experiment organization
- Documentation style
- Development order
- Creative exploration

## 🔄 Session End Protocol

1. Update any in-progress todos
2. Commit work with story-telling message
3. Quick flow journal entry (5 min max)
4. Note any breakthrough ideas in IDEAS.md
5. Set intention for next session

## 📝 Quick Commands Reference

```bash
# Check current todos
TodoRead

# Update task progress
TodoWrite

# Run tests (when implemented)
docker-compose run python-tests
docker-compose run go-tests

# Check API connection
docker-compose run python-check-connection
```

## 🎯 Success Metrics
- Sub-second scanner response
- 99.9% order execution reliability
- Zero pacing violations
- Smooth daily restart handling
- Intuitive 3-click GUI navigation

---

*Remember: The best code is written in flow state. Trust the vibe, document the journey, and build something delightful.*