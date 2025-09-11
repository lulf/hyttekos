# Hyttekos API Specification

## Base URL
```
https://api.hyttekos.local
```

## Authentication

### Frontend Authentication
- **Method**: OAuth2 Authorization Code flow
- **Token Type**: Bearer
- **Headers**: `Authorization: Bearer <access_token>`

### Temperature Sensor Authentication
- **Method**: HTTP Basic Authentication
- **Headers**: `Authorization: Basic <base64(username:password)>`

## API Endpoints

### 1. Authentication Endpoints

#### `GET /auth/login`
Initiates OAuth2 login flow
- **Authentication**: None
- **Response**: Redirect to OAuth2 provider

#### `GET /auth/callback`
OAuth2 callback endpoint
- **Authentication**: None
- **Parameters**: 
  - `code` (query): Authorization code
  - `state` (query): CSRF protection state
- **Response**: Redirect to frontend with session

#### `POST /auth/logout`
Logout and invalidate session
- **Authentication**: Bearer token
- **Response**: `200 OK`

### 2. Cabin Control Endpoints

#### `GET /api/status`
Get current status of all cabin systems
- **Authentication**: Bearer token
- **Response**: `200 OK`
```json
{
  "heating": {
    "enabled": true,
    "last_updated": "2023-12-07T10:30:00Z"
  },
  "hot_water": {
    "enabled": false,
    "last_updated": "2023-12-07T09:15:00Z"
  },
  "temperature": {
    "current": 18.5,
    "timestamp": "2023-12-07T10:35:00Z"
  }
}
```

#### `POST /api/heating`
Control heating system
- **Authentication**: Bearer token
- **Request Body**:
```json
{
  "action": "on|off"
}
```
- **Response**: `200 OK`
```json
{
  "success": true,
  "heating": {
    "enabled": true,
    "last_updated": "2023-12-07T10:35:00Z"
  }
}
```
- **Error Response**: `400 Bad Request`
```json
{
  "error": "Invalid action",
  "message": "Action must be 'on' or 'off'"
}
```

#### `POST /api/hot-water`
Control hot water system
- **Authentication**: Bearer token
- **Request Body**:
```json
{
  "action": "on|off"
}
```
- **Response**: `200 OK`
```json
{
  "success": true,
  "hot_water": {
    "enabled": false,
    "last_updated": "2023-12-07T10:35:00Z"
  }
}
```

### 3. Temperature Endpoints

#### `GET /api/temperature`
Get current temperature reading
- **Authentication**: Bearer token
- **Response**: `200 OK`
```json
{
  "current": 18.5,
  "timestamp": "2023-12-07T10:35:00Z"
}
```

#### `GET /api/temperature/history`
Get temperature history
- **Authentication**: Bearer token
- **Parameters**:
  - `from` (query, optional): ISO8601 timestamp
  - `to` (query, optional): ISO8601 timestamp
  - `limit` (query, optional): Max results (default: 100)
- **Response**: `200 OK`
```json
{
  "readings": [
    {
      "temperature": 18.5,
      "timestamp": "2023-12-07T10:35:00Z"
    },
    {
      "temperature": 18.3,
      "timestamp": "2023-12-07T10:30:00Z"
    }
  ]
}
```

#### `POST /api/temperature`
Submit temperature reading (sensor endpoint)
- **Authentication**: HTTP Basic Auth
- **Request Body**:
```json
{
  "temperature": 18.5,
  "timestamp": "2023-12-07T10:35:00Z"
}
```
- **Response**: `201 Created`
```json
{
  "success": true,
  "message": "Temperature recorded"
}
```

### 4. HTMX-Specific Endpoints

#### `GET /htmx/dashboard`
Get dashboard HTML fragment
- **Authentication**: Bearer token
- **Response**: `200 OK` (HTML)
```html
<div id="dashboard">
  <div class="status-card heating">
    <h3>Heating</h3>
    <span class="status on">On</span>
    <button hx-post="/htmx/heating/toggle" hx-target="#dashboard">Toggle</button>
  </div>
  <!-- ... more cards ... -->
</div>
```

#### `POST /htmx/heating/toggle`
Toggle heating via HTMX
- **Authentication**: Bearer token
- **Response**: `200 OK` (HTML fragment)
```html
<div class="status-card heating">
  <h3>Heating</h3>
  <span class="status off">Off</span>
  <button hx-post="/htmx/heating/toggle" hx-target="#dashboard">Toggle</button>
</div>
```

#### `POST /htmx/hot-water/toggle`
Toggle hot water via HTMX
- **Authentication**: Bearer token
- **Response**: `200 OK` (HTML fragment)

#### `GET /htmx/temperature`
Get temperature display fragment
- **Authentication**: Bearer token
- **Response**: `200 OK` (HTML)
```html
<div class="temperature-display">
  <span class="temp-value">18.5°C</span>
  <span class="temp-time">10:35</span>
</div>
```

## Error Handling

### HTTP Status Codes
- `200 OK`: Successful request
- `201 Created`: Resource created successfully
- `400 Bad Request`: Invalid request parameters
- `401 Unauthorized`: Authentication required
- `403 Forbidden`: Access denied
- `404 Not Found`: Resource not found
- `500 Internal Server Error`: Server error

### Error Response Format
```json
{
  "error": "error_code",
  "message": "Human readable error message",
  "details": {} // Optional additional details
}
```

### Common Error Responses

#### Authentication Errors
```json
{
  "error": "unauthorized",
  "message": "Valid authentication token required"
}
```

#### Akiles API Errors
```json
{
  "error": "external_service_error",
  "message": "Unable to communicate with cabin systems",
  "details": {
    "service": "akiles",
    "status": "timeout"
  }
}
```

#### Validation Errors
```json
{
  "error": "validation_error",
  "message": "Invalid request parameters",
  "details": {
    "field": "action",
    "expected": "on|off"
  }
}
```

## Rate Limiting

### Frontend API Endpoints
- **Rate**: 100 requests per minute per user
- **Headers**: 
  - `X-RateLimit-Limit`: Request limit
  - `X-RateLimit-Remaining`: Remaining requests
  - `X-RateLimit-Reset`: Reset timestamp

### Temperature Sensor Endpoint
- **Rate**: 1 request per minute per sensor
- **Response**: `429 Too Many Requests` if exceeded

## CORS Policy

### Allowed Origins
- Frontend domain (same-origin for production)
- Development localhost (development only)

### Allowed Methods
- `GET`, `POST`, `OPTIONS`

### Allowed Headers
- `Content-Type`
- `Authorization`
- `HX-Request` (for HTMX detection)

## WebSocket Support (Future Enhancement)

### `/ws/status`
Real-time status updates
- **Authentication**: Query parameter token
- **Messages**:
```json
{
  "type": "heating_changed",
  "data": {
    "enabled": true,
    "timestamp": "2023-12-07T10:35:00Z"
  }
}
```

## API Versioning

### Current Version
- **Version**: v1
- **URL Pattern**: `/api/v1/...`

### Version Headers
- `API-Version: v1`
- `Accept: application/json; version=1`

## Testing Endpoints (Development Only)

#### `POST /test/akiles/mock`
Mock Akiles API responses
- **Authentication**: Bearer token
- **Available**: Development environment only