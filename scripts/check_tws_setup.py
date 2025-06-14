#!/usr/bin/env python3
"""
TWS Setup Verification Script

This script helps verify that TWS is properly configured for API access.
Run this before attempting integration tests.

Usage:
    python scripts/check_tws_setup.py
"""

import socket
import sys
import os
import platform
from pathlib import Path


def check_windows_environment():
    """Check if running on Windows."""
    print("🖥️  Environment Check")
    print("-" * 30)
    
    system = platform.system()
    print(f"Operating System: {system}")
    
    if system != "Windows":
        print("⚠️  Warning: TWS API is optimized for Windows")
        print("   Integration tests should be run on Windows")
    else:
        print("✅ Running on Windows")
    
    return system == "Windows"


def check_port_availability(port: int, host: str = "127.0.0.1") -> bool:
    """Check if a port is available (TWS listening)."""
    try:
        with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as sock:
            sock.settimeout(2)
            result = sock.connect_ex((host, port))
            return result == 0
    except Exception:
        return False


def check_tws_ports():
    """Check TWS port availability."""
    print("\n🔌 TWS Port Check")
    print("-" * 30)
    
    ports = {
        7497: "Paper Trading",
        7496: "Live Trading",
        4001: "IB Gateway Paper",
        4000: "IB Gateway Live"
    }
    
    available_ports = []
    
    for port, description in ports.items():
        if check_port_availability(port):
            print(f"✅ Port {port} ({description}) - AVAILABLE")
            available_ports.append(port)
        else:
            print(f"❌ Port {port} ({description}) - Not available")
    
    if not available_ports:
        print("\n🚨 No TWS ports are available!")
        print("   Please start TWS or IB Gateway")
        return False
    
    print(f"\n📊 Found {len(available_ports)} available TWS port(s)")
    return True


def check_python_environment():
    """Check Python environment for required packages."""
    print("\n🐍 Python Environment Check")
    print("-" * 30)
    
    print(f"Python Version: {sys.version}")
    
    required_packages = [
        "ib_insync",
        "asyncio",
        "pytest"
    ]
    
    missing_packages = []
    
    for package in required_packages:
        try:
            __import__(package)
            print(f"✅ {package} - Available")
        except ImportError:
            print(f"❌ {package} - Missing")
            missing_packages.append(package)
    
    if missing_packages:
        print(f"\n🚨 Missing packages: {', '.join(missing_packages)}")
        print("   Install with: pip install -r requirements.txt")
        return False
    
    return True


def check_project_structure():
    """Check project structure is correct."""
    print("\n📁 Project Structure Check")
    print("-" * 30)
    
    required_paths = [
        "src/python/ibkr_connector/connection.py",
        "src/python/config/settings.py",
        "tests/python/integration/test_connection_integration.py",
        "requirements.txt"
    ]
    
    missing_paths = []
    
    for path in required_paths:
        if Path(path).exists():
            print(f"✅ {path}")
        else:
            print(f"❌ {path} - Missing")
            missing_paths.append(path)
    
    if missing_paths:
        print(f"\n🚨 Missing files: {len(missing_paths)}")
        return False
    
    return True


def provide_tws_setup_guidance():
    """Provide TWS setup guidance."""
    print("\n📋 TWS Configuration Checklist")
    print("-" * 30)
    
    checklist = [
        "1. Start TWS (Trader Workstation)",
        "2. Go to File → Global Configuration → API → Settings",
        "3. ✅ Enable 'ActiveX and Socket Clients'",
        "4. ❌ Disable 'Read-Only API'",
        "5. Set 'Socket port' to 7497 (paper) or 7496 (live)",
        "6. Add '127.0.0.1' to 'Trusted IPs' (if not already there)",
        "7. Set 'Master API client ID' to 0",
        "8. Click 'OK' and restart TWS",
        "9. Log in to your paper trading account",
        "10. Keep TWS running during integration tests"
    ]
    
    for item in checklist:
        print(item)
    
    print("\n🎯 Key Points:")
    print("• Use Client ID 1-999 for API connections (0 reserved for manual)")
    print("• Paper trading is safer for testing")
    print("• TWS must stay running during tests")
    print("• Check TWS logs if connection fails")


def main():
    """Main verification routine."""
    print("🚀 TWS Integration Setup Verification")
    print("=" * 50)
    
    all_good = True
    
    # Check environment
    if not check_windows_environment():
        all_good = False
    
    # Check Python setup
    if not check_python_environment():
        all_good = False
    
    # Check project structure
    if not check_project_structure():
        all_good = False
    
    # Check TWS ports
    tws_available = check_tws_ports()
    if not tws_available:
        all_good = False
        provide_tws_setup_guidance()
    
    # Final assessment
    print("\n" + "=" * 50)
    if all_good and tws_available:
        print("🎉 Ready for Integration Testing!")
        print("\nNext steps:")
        print("1. Run basic connection test:")
        print("   python src/python/test_connection.py")
        print("2. Run integration tests:")
        print("   pytest tests/python/integration/ -m integration -v")
    else:
        print("⚠️  Setup Issues Detected")
        print("\nPlease resolve the issues above before running integration tests.")
        if not tws_available:
            print("\n💡 TWS Setup is critical - see checklist above")
    
    return all_good and tws_available


if __name__ == "__main__":
    success = main()
    sys.exit(0 if success else 1) 