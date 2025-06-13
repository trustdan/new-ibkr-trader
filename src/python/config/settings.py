"""
Configuration settings for IBKR connector.

This module provides configuration management for the IBKR connection,
including environment-based settings and validation.
"""

import os
from dataclasses import dataclass, field
from datetime import time
from typing import Optional


@dataclass
class ConnectionConfig:
    """Configuration for IBKR API connection."""
    host: str = "127.0.0.1"
    port: int = 7497  # Default to paper trading port
    client_id: int = 1
    timeout: float = 30.0
    account: Optional[str] = None
    
    @classmethod
    def from_env(cls) -> 'ConnectionConfig':
        """Create configuration from environment variables."""
        return cls(
            host=os.getenv('IBKR_HOST', cls.host),
            port=int(os.getenv('IBKR_PORT', cls.port)),
            client_id=int(os.getenv('IBKR_CLIENT_ID', cls.client_id)),
            timeout=float(os.getenv('IBKR_TIMEOUT', cls.timeout)),
            account=os.getenv('IBKR_ACCOUNT')
        )


@dataclass
class RateLimitConfig:
    """Configuration for API rate limiting."""
    max_requests_per_second: float = 45.0  # Safety margin below 50
    burst_size: int = 10
    throttle_wait: float = 0.025  # 25ms between requests
    
    @classmethod
    def from_env(cls) -> 'RateLimitConfig':
        """Create configuration from environment variables."""
        return cls(
            max_requests_per_second=float(
                os.getenv('IBKR_MAX_REQ_PER_SEC', cls.max_requests_per_second)
            ),
            burst_size=int(os.getenv('IBKR_BURST_SIZE', cls.burst_size)),
            throttle_wait=float(os.getenv('IBKR_THROTTLE_WAIT', cls.throttle_wait))
        )


@dataclass
class WatchdogConfig:
    """Configuration for connection watchdog."""
    enabled: bool = True
    reconnect_interval: float = 2.0
    max_reconnect_interval: float = 60.0
    backoff_factor: float = 2.0
    daily_restart_time: time = time(23, 55)  # 5 mins before midnight
    health_check_interval: float = 30.0
    
    @classmethod
    def from_env(cls) -> 'WatchdogConfig':
        """Create configuration from environment variables."""
        restart_hour = int(os.getenv('IBKR_RESTART_HOUR', '23'))
        restart_minute = int(os.getenv('IBKR_RESTART_MINUTE', '55'))
        
        return cls(
            enabled=os.getenv('IBKR_WATCHDOG_ENABLED', 'true').lower() == 'true',
            reconnect_interval=float(
                os.getenv('IBKR_RECONNECT_INTERVAL', cls.reconnect_interval)
            ),
            max_reconnect_interval=float(
                os.getenv('IBKR_MAX_RECONNECT_INTERVAL', cls.max_reconnect_interval)
            ),
            backoff_factor=float(
                os.getenv('IBKR_BACKOFF_FACTOR', cls.backoff_factor)
            ),
            daily_restart_time=time(restart_hour, restart_minute),
            health_check_interval=float(
                os.getenv('IBKR_HEALTH_CHECK_INTERVAL', cls.health_check_interval)
            )
        )


@dataclass
class LoggingConfig:
    """Configuration for logging."""
    level: str = "INFO"
    format: str = "%(asctime)s - %(name)s - %(levelname)s - %(message)s"
    log_all_messages: bool = False
    log_to_file: bool = True
    log_dir: str = "logs"
    
    @classmethod
    def from_env(cls) -> 'LoggingConfig':
        """Create configuration from environment variables."""
        return cls(
            level=os.getenv('LOG_LEVEL', cls.level),
            format=os.getenv('LOG_FORMAT', cls.format),
            log_all_messages=os.getenv('LOG_ALL_MESSAGES', 'false').lower() == 'true',
            log_to_file=os.getenv('LOG_TO_FILE', 'true').lower() == 'true',
            log_dir=os.getenv('LOG_DIR', cls.log_dir)
        )


@dataclass
class Config:
    """Main configuration container."""
    connection: ConnectionConfig = field(default_factory=ConnectionConfig)
    rate_limit: RateLimitConfig = field(default_factory=RateLimitConfig)
    watchdog: WatchdogConfig = field(default_factory=WatchdogConfig)
    logging: LoggingConfig = field(default_factory=LoggingConfig)
    
    @classmethod
    def from_env(cls) -> 'Config':
        """Create complete configuration from environment variables."""
        return cls(
            connection=ConnectionConfig.from_env(),
            rate_limit=RateLimitConfig.from_env(),
            watchdog=WatchdogConfig.from_env(),
            logging=LoggingConfig.from_env()
        )
    
    def validate(self) -> None:
        """Validate configuration settings."""
        # Connection validation
        if self.connection.port not in [7496, 7497, 4001, 4002]:
            raise ValueError(f"Invalid port {self.connection.port}. "
                           "Use 7496/7497 for TWS or 4001/4002 for Gateway")
        
        if self.connection.client_id < 0:
            raise ValueError("Client ID must be non-negative")
        
        # Rate limit validation
        if self.rate_limit.max_requests_per_second > 50:
            raise ValueError("Max requests per second cannot exceed 50")
        
        if self.rate_limit.max_requests_per_second <= 0:
            raise ValueError("Max requests per second must be positive")
        
        # Watchdog validation
        if self.watchdog.reconnect_interval <= 0:
            raise ValueError("Reconnect interval must be positive")
        
        if self.watchdog.backoff_factor <= 1:
            raise ValueError("Backoff factor must be greater than 1")


# Default configuration instance
default_config = Config()