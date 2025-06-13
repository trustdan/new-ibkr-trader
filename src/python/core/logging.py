"""
Vibe-aware logging configuration
Maintains flow while providing essential debugging info
"""
import os
import sys
import logging
from datetime import datetime
from pythonjsonlogger import jsonlogger


def setup_logging(name: str = "ibkr-service") -> logging.Logger:
    """
    Setup structured logging that maintains developer flow
    
    - JSON format for production parsing
    - Human-friendly for development
    - Async-safe configuration
    """
    logger = logging.getLogger(name)
    
    # Clear existing handlers
    logger.handlers.clear()
    
    # Set log level from environment
    log_level = os.getenv("LOG_LEVEL", "INFO").upper()
    logger.setLevel(getattr(logging, log_level))
    
    # Create handler
    handler = logging.StreamHandler(sys.stdout)
    
    # Use JSON format in production, human-friendly in development
    if os.getenv("ENV", "development") == "production":
        # JSON format for structured logging
        formatter = jsonlogger.JsonFormatter(
            "%(timestamp)s %(level)s %(name)s %(message)s",
            rename_fields={"timestamp": "@timestamp", "level": "severity"}
        )
    else:
        # Human-friendly format with emojis for vibe
        class VibeFormatter(logging.Formatter):
            """Custom formatter with vibe-friendly output"""
            
            EMOJIS = {
                logging.DEBUG: "🔍",
                logging.INFO: "📝",
                logging.WARNING: "⚠️",
                logging.ERROR: "❌",
                logging.CRITICAL: "🚨"
            }
            
            def format(self, record):
                emoji = self.EMOJIS.get(record.levelno, "📌")
                timestamp = datetime.fromtimestamp(record.created).strftime("%H:%M:%S")
                
                # Add color codes for terminal
                if record.levelno >= logging.ERROR:
                    color_start = "\033[91m"  # Red
                elif record.levelno >= logging.WARNING:
                    color_start = "\033[93m"  # Yellow
                else:
                    color_start = "\033[0m"   # Default
                color_end = "\033[0m"
                
                return f"{color_start}{timestamp} {emoji} [{record.name}] {record.getMessage()}{color_end}"
        
        formatter = VibeFormatter()
    
    handler.setFormatter(formatter)
    logger.addHandler(handler)
    
    # Prevent propagation to avoid duplicate logs
    logger.propagate = False
    
    # Log startup message
    logger.info(f"Logging initialized at {log_level} level")
    
    return logger


def get_logger(name: str) -> logging.Logger:
    """Get a logger instance with consistent configuration"""
    return logging.getLogger(name)