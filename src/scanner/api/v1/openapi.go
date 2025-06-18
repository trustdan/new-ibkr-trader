package v1

import (
	"encoding/json"
	"net/http"
)

// openAPISpec returns the OpenAPI 3.0 specification
func (api *API) openAPISpec(w http.ResponseWriter, r *http.Request) {
	spec := map[string]interface{}{
		"openapi": "3.0.0",
		"info": map[string]interface{}{
			"title":       "IBKR Scanner API",
			"description": "High-performance options scanner for Interactive Brokers",
			"version":     "1.0.0",
			"contact": map[string]string{
				"name":  "IBKR Trader",
				"email": "support@ibkr-trader.com",
			},
		},
		"servers": []map[string]string{
			{
				"url":         "http://localhost:8080/api/v1",
				"description": "Local development server",
			},
			{
				"url":         "https://api.ibkr-trader.com/v1",
				"description": "Production server",
			},
		},
		"paths": getAPIPaths(),
		"components": map[string]interface{}{
			"schemas":   getAPISchemas(),
			"responses": getAPIResponses(),
			"parameters": getAPIParameters(),
			"securitySchemes": map[string]interface{}{
				"ApiKeyAuth": map[string]string{
					"type": "apiKey",
					"in":   "header",
					"name": "X-API-Key",
				},
			},
		},
		"tags": []map[string]string{
			{"name": "Scanner", "description": "Options scanning operations"},
			{"name": "Filters", "description": "Filter management"},
			{"name": "Analytics", "description": "Analytics and reporting"},
			{"name": "History", "description": "Historical data access"},
			{"name": "System", "description": "System information and health"},
		},
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(spec)
}

func getAPIPaths() map[string]interface{} {
	return map[string]interface{}{
		"/health": map[string]interface{}{
			"get": map[string]interface{}{
				"tags":        []string{"System"},
				"summary":     "Health check",
				"description": "Returns the health status of the service",
				"operationId": "getHealth",
				"responses": map[string]interface{}{
					"200": map[string]interface{}{
						"description": "Service is healthy",
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"$ref": "#/components/schemas/HealthResponse",
								},
							},
						},
					},
				},
			},
		},
		"/scan/{symbol}": map[string]interface{}{
			"get": map[string]interface{}{
				"tags":        []string{"Scanner"},
				"summary":     "Scan single symbol",
				"description": "Scans options for a single symbol with optional filter overrides",
				"operationId": "scanSymbol",
				"parameters": []map[string]interface{}{
					{
						"name":        "symbol",
						"in":          "path",
						"required":    true,
						"description": "Stock symbol to scan",
						"schema": map[string]string{
							"type":    "string",
							"example": "AAPL",
						},
					},
					{"$ref": "#/components/parameters/DeltaMin"},
					{"$ref": "#/components/parameters/DeltaMax"},
					{"$ref": "#/components/parameters/DTEMin"},
					{"$ref": "#/components/parameters/DTEMax"},
				},
				"responses": map[string]interface{}{
					"200": map[string]interface{}{
						"description": "Scan completed successfully",
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"$ref": "#/components/schemas/ScanResponse",
								},
							},
						},
					},
					"400": {"$ref": "#/components/responses/BadRequest"},
					"500": {"$ref": "#/components/responses/InternalError"},
				},
			},
		},
		"/scan": map[string]interface{}{
			"post": map[string]interface{}{
				"tags":        []string{"Scanner"},
				"summary":     "Scan multiple symbols",
				"description": "Scans options for multiple symbols with custom filters",
				"operationId": "scanMultiple",
				"requestBody": map[string]interface{}{
					"required": true,
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"$ref": "#/components/schemas/ScanRequest",
							},
						},
					},
				},
				"responses": map[string]interface{}{
					"200": map[string]interface{}{
						"description": "Scan completed successfully",
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"type": "object",
									"properties": map[string]interface{}{
										"data": map[string]interface{}{
											"type": "array",
											"items": map[string]interface{}{
												"$ref": "#/components/schemas/ScanResult",
											},
										},
									},
								},
							},
						},
					},
					"400": {"$ref": "#/components/responses/BadRequest"},
					"500": {"$ref": "#/components/responses/InternalError"},
				},
			},
		},
		"/filters": map[string]interface{}{
			"get": map[string]interface{}{
				"tags":        []string{"Filters"},
				"summary":     "Get current filters",
				"description": "Returns the current filter configuration",
				"operationId": "getFilters",
				"responses": map[string]interface{}{
					"200": map[string]interface{}{
						"description": "Current filter configuration",
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"$ref": "#/components/schemas/FilterConfig",
								},
							},
						},
					},
				},
			},
			"put": map[string]interface{}{
				"tags":        []string{"Filters"},
				"summary":     "Update filters",
				"description": "Updates the scanner filter configuration",
				"operationId": "updateFilters",
				"requestBody": map[string]interface{}{
					"required": true,
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"$ref": "#/components/schemas/FilterConfig",
							},
						},
					},
				},
				"responses": map[string]interface{}{
					"200": map[string]interface{}{
						"description": "Filters updated successfully",
					},
					"400": {"$ref": "#/components/responses/BadRequest"},
				},
			},
		},
		"/ws": map[string]interface{}{
			"get": map[string]interface{}{
				"tags":        []string{"Scanner"},
				"summary":     "WebSocket streaming",
				"description": "Establishes WebSocket connection for real-time updates",
				"operationId": "websocket",
				"responses": map[string]interface{}{
					"101": map[string]interface{}{
						"description": "Switching Protocols",
					},
				},
			},
		},
	}
}

func getAPISchemas() map[string]interface{} {
	return map[string]interface{}{
		"HealthResponse": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"status": map[string]string{
					"type":    "string",
					"example": "healthy",
				},
				"timestamp": map[string]interface{}{
					"type":    "integer",
					"format":  "int64",
					"example": 1634567890,
				},
				"version": map[string]string{
					"type":    "string",
					"example": "1.0.0",
				},
				"uptime": map[string]interface{}{
					"type":        "number",
					"description": "Uptime in seconds",
					"example":     3600.5,
				},
			},
		},
		"ScanRequest": map[string]interface{}{
			"type":     "object",
			"required": []string{"symbols"},
			"properties": map[string]interface{}{
				"symbols": map[string]interface{}{
					"type":        "array",
					"description": "List of symbols to scan",
					"items": map[string]string{
						"type": "string",
					},
					"example": []string{"AAPL", "MSFT", "GOOGL"},
				},
				"filters": map[string]interface{}{
					"$ref": "#/components/schemas/FilterConfig",
				},
			},
		},
		"ScanResponse": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"data": map[string]interface{}{
					"$ref": "#/components/schemas/ScanResult",
				},
				"status": map[string]string{
					"type":    "string",
					"example": "success",
				},
				"timestamp": map[string]interface{}{
					"type":   "integer",
					"format": "int64",
				},
			},
		},
		"ScanResult": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"symbol": map[string]string{
					"type":    "string",
					"example": "AAPL",
				},
				"scan_time": map[string]string{
					"type":    "string",
					"format":  "date-time",
					"example": "2023-10-15T14:30:00Z",
				},
				"total_contracts": map[string]interface{}{
					"type":    "integer",
					"example": 1500,
				},
				"filtered_contracts": map[string]interface{}{
					"type":    "integer",
					"example": 42,
				},
				"spreads": map[string]interface{}{
					"type": "array",
					"items": map[string]interface{}{
						"$ref": "#/components/schemas/SpreadResult",
					},
				},
			},
		},
		"SpreadResult": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"long_strike": map[string]interface{}{
					"type":    "number",
					"example": 150.0,
				},
				"short_strike": map[string]interface{}{
					"type":    "number",
					"example": 155.0,
				},
				"expiration": map[string]string{
					"type":    "string",
					"format":  "date",
					"example": "2023-11-17",
				},
				"net_credit": map[string]interface{}{
					"type":    "number",
					"example": 2.35,
				},
				"max_profit": map[string]interface{}{
					"type":    "number",
					"example": 235.0,
				},
				"max_loss": map[string]interface{}{
					"type":    "number",
					"example": 265.0,
				},
				"probability_profit": map[string]interface{}{
					"type":    "number",
					"example": 0.72,
				},
				"score": map[string]interface{}{
					"type":    "number",
					"example": 85.5,
				},
			},
		},
		"FilterConfig": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"delta": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"min": map[string]interface{}{"type": "number", "example": 0.20},
						"max": map[string]interface{}{"type": "number", "example": 0.35},
					},
				},
				"dte": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"min": map[string]interface{}{"type": "integer", "example": 30},
						"max": map[string]interface{}{"type": "integer", "example": 60},
					},
				},
				"liquidity": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"min_open_interest": map[string]interface{}{"type": "integer", "example": 100},
						"min_volume":        map[string]interface{}{"type": "integer", "example": 50},
					},
				},
				"spread": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"min_credit":    map[string]interface{}{"type": "number", "example": 0.50},
						"max_width":     map[string]interface{}{"type": "number", "example": 10.0},
						"min_risk_reward": map[string]interface{}{"type": "number", "example": 0.5},
					},
				},
			},
		},
		"ErrorResponse": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"error": map[string]string{
					"type":        "string",
					"description": "Error message",
				},
				"status": map[string]interface{}{
					"type":        "integer",
					"description": "HTTP status code",
				},
				"timestamp": map[string]interface{}{
					"type":   "integer",
					"format": "int64",
				},
			},
		},
	}
}

func getAPIResponses() map[string]interface{} {
	return map[string]interface{}{
		"BadRequest": map[string]interface{}{
			"description": "Bad request",
			"content": map[string]interface{}{
				"application/json": map[string]interface{}{
					"schema": map[string]interface{}{
						"$ref": "#/components/schemas/ErrorResponse",
					},
				},
			},
		},
		"NotFound": map[string]interface{}{
			"description": "Resource not found",
			"content": map[string]interface{}{
				"application/json": map[string]interface{}{
					"schema": map[string]interface{}{
						"$ref": "#/components/schemas/ErrorResponse",
					},
				},
			},
		},
		"InternalError": map[string]interface{}{
			"description": "Internal server error",
			"content": map[string]interface{}{
				"application/json": map[string]interface{}{
					"schema": map[string]interface{}{
						"$ref": "#/components/schemas/ErrorResponse",
					},
				},
			},
		},
	}
}

func getAPIParameters() map[string]interface{} {
	return map[string]interface{}{
		"DeltaMin": map[string]interface{}{
			"name":        "delta_min",
			"in":          "query",
			"description": "Minimum delta value",
			"required":    false,
			"schema": map[string]interface{}{
				"type":    "number",
				"minimum": 0,
				"maximum": 1,
				"example": 0.20,
			},
		},
		"DeltaMax": map[string]interface{}{
			"name":        "delta_max",
			"in":          "query",
			"description": "Maximum delta value",
			"required":    false,
			"schema": map[string]interface{}{
				"type":    "number",
				"minimum": 0,
				"maximum": 1,
				"example": 0.35,
			},
		},
		"DTEMin": map[string]interface{}{
			"name":        "dte_min",
			"in":          "query",
			"description": "Minimum days to expiration",
			"required":    false,
			"schema": map[string]interface{}{
				"type":    "integer",
				"minimum": 0,
				"example": 30,
			},
		},
		"DTEMax": map[string]interface{}{
			"name":        "dte_max",
			"in":          "query",
			"description": "Maximum days to expiration",
			"required":    false,
			"schema": map[string]interface{}{
				"type":    "integer",
				"minimum": 0,
				"example": 60,
			},
		},
	}
}