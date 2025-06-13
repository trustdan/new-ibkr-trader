#!/usr/bin/env python
"""
Development helper script for common tasks
"""
import click
import subprocess
import os
from pathlib import Path
from datetime import datetime

@click.group()
def cli():
    """IBKR Spread Automation Development Helper"""
    pass

@cli.command()
def vibe():
    """Check the current vibe and show flow journal"""
    click.echo("🌊 Checking the vibe...")
    
    # Show manifesto if exists
    manifesto_path = Path('.vibe/manifesto.md')
    if manifesto_path.exists():
        click.echo("\n📜 Vibe Manifesto:")
        click.echo("-" * 40)
        with open(manifesto_path) as f:
            click.echo(f.read())
    
    # Show latest flow journal
    flow_dir = Path('flow_journal')
    if flow_dir.exists():
        journals = sorted(flow_dir.glob('*.md'), key=lambda x: x.stat().st_mtime, reverse=True)
        if journals:
            latest = journals[0]
            click.echo(f"\n📝 Latest Flow Journal: {latest.name}")
            click.echo("-" * 40)
            with open(latest) as f:
                # Show last 20 lines
                lines = f.readlines()
                click.echo(''.join(lines[-20:]))

@cli.command()
@click.option('--title', prompt='Session title', help='Title for this flow session')
def flow_start(title):
    """Start a new flow journal entry"""
    date_str = datetime.now().strftime('%Y-%m-%d')
    time_str = datetime.now().strftime('%H:%M')
    
    journal_dir = Path('flow_journal')
    journal_dir.mkdir(exist_ok=True)
    
    filename = f"{date_str}-{title.lower().replace(' ', '-')}.md"
    filepath = journal_dir / filename
    
    template = f"""# Flow Journal - {date_str} - {title}

## 🌅 Morning Intention
- Energy level: [1-10]
- Focus area: {title}
- Vibe: [Current mood/energy]
- Started: {time_str}

## 🚀 Session Highlights

### Breakthroughs
- 

### Challenges
- 

### Code Snippets
```python
# Interesting patterns discovered
```

## 📚 API Learnings
- 

## 🎯 Progress Check
- [ ] Maintained flow state
- [ ] Updated documentation
- [ ] Committed with story
- [ ] No pacing violations
- [ ] Tests passing

## 🌊 Tomorrow's Flow
- 

## 🎨 Vibe Check
- Flow state achieved: [Yes/No/Partial]
- Best working music: 
- Environment notes: 
- Overall satisfaction: [1-10]
"""
    
    with open(filepath, 'w') as f:
        f.write(template)
    
    click.echo(f"✨ Flow journal created: {filepath}")
    click.echo("🎯 Go forth and code with intention!")

@cli.command()
def experiment():
    """Create a new experiment folder with timestamp"""
    timestamp = datetime.now().strftime('%Y%m%d_%H%M%S')
    exp_name = click.prompt('Experiment name')
    
    exp_dir = Path('experiments') / f"{timestamp}_{exp_name.lower().replace(' ', '_')}"
    exp_dir.mkdir(parents=True, exist_ok=True)
    
    # Create README
    readme_path = exp_dir / 'README.md'
    with open(readme_path, 'w') as f:
        f.write(f"""# Experiment: {exp_name}
Created: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}

## Purpose
[What are we testing/exploring?]

## Hypothesis
[What do we expect to learn?]

## Results
[What did we discover?]

## Next Steps
[How does this inform the main project?]
""")
    
    # Create test script
    test_script = exp_dir / 'test.py'
    with open(test_script, 'w') as f:
        f.write("""#!/usr/bin/env python
\"\"\"
Experiment: {exp_name}
\"\"\"
import asyncio
from ib_insync import IB, util

async def experiment():
    \"\"\"Run experiment\"\"\"
    ib = IB()
    try:
        await ib.connectAsync('localhost', 7497, clientId=900)
        
        # Experiment code here
        
    finally:
        ib.disconnect()

if __name__ == '__main__':
    util.run(experiment())
""".format(exp_name=exp_name))
    
    click.echo(f"🧪 Experiment created: {exp_dir}")
    click.echo("💡 Start experimenting!")

@cli.command()
@click.argument('message')
def commit(message):
    """Make a vibe-conscious git commit"""
    # Add emoji based on keywords
    emoji_map = {
        'fix': '🐛',
        'feat': '✨',
        'docs': '📚',
        'refactor': '♻️',
        'test': '🧪',
        'perf': '⚡',
        'style': '🎨',
        'build': '🏗️',
        'chore': '🔧',
        'async': '⚡',
        'connection': '🔌',
        'scanner': '🔍',
        'gui': '🖥️',
        'vibe': '🌊'
    }
    
    # Find appropriate emoji
    emoji = '🚀'  # default
    message_lower = message.lower()
    for keyword, emoji_char in emoji_map.items():
        if keyword in message_lower:
            emoji = emoji_char
            break
    
    # Create commit message
    commit_msg = f"{emoji} {message}"
    
    # Show what will be committed
    click.echo("📝 Git status:")
    subprocess.run(['git', 'status', '--short'])
    
    if click.confirm(f"\nCommit with message: '{commit_msg}'?"):
        subprocess.run(['git', 'add', '-A'])
        subprocess.run(['git', 'commit', '-m', commit_msg])
        click.echo("✅ Committed successfully!")
    else:
        click.echo("❌ Commit cancelled")

@cli.command()
def metrics():
    """Show development metrics"""
    click.echo("📊 Development Metrics")
    click.echo("=" * 40)
    
    # Count files
    py_files = len(list(Path('.').rglob('*.py')))
    go_files = len(list(Path('.').rglob('*.go')))
    js_files = len(list(Path('.').rglob('*.js'))) + len(list(Path('.').rglob('*.svelte')))
    
    click.echo(f"\n📁 File Count:")
    click.echo(f"  Python files: {py_files}")
    click.echo(f"  Go files: {go_files}")
    click.echo(f"  JS/Svelte files: {js_files}")
    
    # Count TODOs
    todo_count = 0
    for ext in ['*.py', '*.go', '*.js', '*.svelte', '*.md']:
        for file in Path('.').rglob(ext):
            try:
                with open(file, 'r', encoding='utf-8') as f:
                    content = f.read()
                    todo_count += content.count('TODO')
            except:
                pass
    
    click.echo(f"\n📋 TODOs found: {todo_count}")
    
    # Show recent commits
    click.echo("\n📝 Recent commits:")
    result = subprocess.run(['git', 'log', '--oneline', '-5'], 
                          capture_output=True, text=True)
    click.echo(result.stdout)

if __name__ == '__main__':
    cli()