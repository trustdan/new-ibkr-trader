#!/usr/bin/env python
"""
Environment validation script to ensure all dependencies and configurations are correct
"""
import sys
import os
import subprocess
import json
from pathlib import Path

def check_python_version():
    """Ensure Python 3.11+ is installed"""
    version = sys.version_info
    if version.major == 3 and version.minor >= 11:
        print("✅ Python version: {}.{}.{}".format(*version[:3]))
        return True
    else:
        print("❌ Python 3.11+ required, found: {}.{}.{}".format(*version[:3]))
        return False

def check_required_packages():
    """Check if required packages are available"""
    required = [
        'ib_insync',
        'asyncio',
        'aiohttp',
        'prometheus_client',
        'docker',
        'pytest',
        'pytest-asyncio'
    ]
    
    missing = []
    for package in required:
        try:
            __import__(package)
            print(f"✅ Package '{package}' is installed")
        except ImportError:
            print(f"❌ Package '{package}' is NOT installed")
            missing.append(package)
    
    return missing

def check_docker():
    """Check if Docker is running"""
    try:
        result = subprocess.run(['docker', 'info'], 
                              capture_output=True, 
                              text=True,
                              timeout=5)
        if result.returncode == 0:
            print("✅ Docker is running")
            return True
        else:
            print("❌ Docker is not running")
            return False
    except (subprocess.TimeoutExpired, FileNotFoundError):
        print("❌ Docker not found or not responding")
        return False

def check_tws_ports():
    """Check if TWS ports are available"""
    import socket
    
    ports = {
        7497: "Paper Trading",
        7496: "Live Trading"
    }
    
    for port, desc in ports.items():
        sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        sock.settimeout(1)
        try:
            result = sock.connect_ex(('127.0.0.1', port))
            if result == 0:
                print(f"✅ Port {port} ({desc}) is OPEN - TWS may be running")
            else:
                print(f"⚠️  Port {port} ({desc}) is CLOSED - TWS not accessible")
        except:
            print(f"❌ Error checking port {port}")
        finally:
            sock.close()

def check_project_structure():
    """Verify project directories exist"""
    required_dirs = [
        'src/python',
        'src/go',
        'src/gui',
        'docker',
        'docs',
        'tests',
        'monitoring',
        'experiments',
        '.vibe',
        'flow_journal',
        'ADR'
    ]
    
    print("\n📁 Checking project structure:")
    for dir_path in required_dirs:
        path = Path(dir_path)
        if path.exists():
            print(f"✅ {dir_path}")
        else:
            print(f"❌ {dir_path} - Missing")

def create_env_report():
    """Create environment validation report"""
    report = {
        'timestamp': subprocess.run(['date'], capture_output=True, text=True).stdout.strip(),
        'python_version': f"{sys.version_info.major}.{sys.version_info.minor}.{sys.version_info.micro}",
        'platform': sys.platform,
        'docker_available': check_docker(),
        'project_root': os.getcwd()
    }
    
    report_path = Path('.vibe/env_report.json')
    report_path.parent.mkdir(exist_ok=True)
    
    with open(report_path, 'w') as f:
        json.dump(report, f, indent=2)
    
    print(f"\n📄 Environment report saved to: {report_path}")

def main():
    print("🔍 IBKR Spread Automation - Environment Check")
    print("=" * 50)
    
    # Check Python version
    print("\n🐍 Python Environment:")
    if not check_python_version():
        print("Please upgrade to Python 3.11+")
        sys.exit(1)
    
    # Check packages
    print("\n📦 Required Packages:")
    missing = check_required_packages()
    if missing:
        print(f"\nInstall missing packages with:")
        print(f"pip install {' '.join(missing)}")
    
    # Check Docker
    print("\n🐳 Docker Status:")
    check_docker()
    
    # Check TWS ports
    print("\n🔌 TWS Port Status:")
    check_tws_ports()
    
    # Check project structure
    check_project_structure()
    
    # Create report
    create_env_report()
    
    print("\n✨ Environment check complete!")
    
    if missing:
        print("\n⚠️  Some issues found - please address them before proceeding")
        sys.exit(1)
    else:
        print("\n✅ Environment ready for development!")

if __name__ == '__main__':
    main()