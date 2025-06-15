"""
Custom exceptions for IBKR connector operations.

This module defines specific exceptions for different error scenarios
that can occur when interacting with the Interactive Brokers API.
"""


class IBKRError(Exception):
    """Base exception for all IBKR-related errors."""
    pass


class ConnectionError(IBKRError):
    """Raised when connection to TWS/Gateway fails."""
    pass


class AuthenticationError(IBKRError):
    """Raised when authentication with TWS fails."""
    pass


class RateLimitError(IBKRError):
    """Raised when API rate limit is exceeded (Error 100)."""
    def __init__(self, message: str, retry_after: float = None):
        super().__init__(message)
        self.retry_after = retry_after


class MarketDataError(IBKRError):
    """Raised when market data operations fail."""
    pass


class OrderError(IBKRError):
    """Raised when order placement or management fails."""
    pass


class ConfigurationError(IBKRError):
    """Raised when configuration is invalid or missing."""
    pass


class TWSDailyRestartError(IBKRError):
    """Raised during TWS daily restart window."""
    pass


class WatchdogError(IBKRError):
    """Raised when watchdog operations fail."""
    pass


# Error code mapping for TWS-specific errors
TWS_ERROR_MAPPING = {
    100: RateLimitError,
    502: ConnectionError,
    504: ConnectionError,
    1100: ConnectionError,
    1102: ConnectionError,
    2103: MarketDataError,
    2104: MarketDataError,
    2105: MarketDataError,
    2106: MarketDataError,
}