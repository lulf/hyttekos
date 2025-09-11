# Hyttekos Architecture Design

## System Overview

Hyttekos is a cabin management system that provides remote control of heating and hot water systems, along with temperature monitoring. The system follows a client-server architecture with a Go backend and a lightweight HTML/HTMX frontend.

## Architecture Components

### 1. Frontend (Web Client)
- **Technology**: HTML, CSS, HTMX, minimal JavaScript
- **Authentication**: OAuth2 flow with backend
- **Functionality**: 
  - Display heating/hot water status
  - Toggle heating/hot water controls
  - Show current temperature readings
  - Cozy cabin-themed UI design

### 2. Backend (Go Service)
- **Technology**: Go (Golang)
- **Database**: PostgreSQL for all data persistence
- **Authentication**: 
  - OAuth2 integration with Akiles API
  - HTTP Basic Auth for temperature sensor endpoints
- **Functionality**:
  - Proxy requests to Akiles API
  - Store temperature readings and session data
  - Provide REST API for frontend
  - Handle authentication flows

### 3. External Services

#### Akiles API Integration
- **Purpose**: Control heating and hot water gadgets
- **Authentication**: OAuth2 with client credentials
- **Key Endpoints**:
  - Gadget state retrieval
  - Gadget action execution (on/off)
- **Configuration**: Organization and gadget IDs via environment variables

#### Temperature Sensors
- **Input Method**: HTTP Basic Auth protected endpoint
- **Storage**: Persistent storage in backend database
- **Access**: Available through backend API

## System Architecture Diagram

```
┌─────────────────┐    HTTPS/OAuth2    ┌──────────────────┐
│   Web Frontend  │◄──────────────────►│   Go Backend     │
│   (HTMX/HTML)   │                    │                  │
└─────────────────┘                    │  ┌─────────────┐ │
                                       │  │PostgreSQL   │ │
                                       │  │Database     │ │
                                       │  └─────────────┘ │
                                       └──────┬───────────┘
                                              │
                               OAuth2/HTTPS   │   HTTP Basic Auth
                                              │   (Temperature Input)
                          ┌───────────────────┼─────────────────┐
                          │                   │                 │
                          ▼                   ▼                 ▼
               ┌─────────────────┐   ┌─────────────────┐  ┌──────────┐
               │   Akiles API    │   │   Cabin Sensors │  │Temperature│
               │                 │   │                 │  │ Sensors  │
               │ ┌─────────────┐ │   │ ┌─────────────┐ │  │          │
               │ │   Heating   │ │   │ │   Heating   │ │  └──────────┘
               │ │   Gadget    │ │   │ │   Hardware  │ │
               │ └─────────────┘ │   │ └─────────────┘ │
               │                 │   │                 │
               │ ┌─────────────┐ │   │ ┌─────────────┐ │
               │ │ Hot Water   │ │   │ │ Hot Water   │ │
               │ │   Gadget    │ │   │ │  Hardware   │ │
               │ └─────────────┘ │   │ └─────────────┘ │
               └─────────────────┘   └─────────────────┘
```

## Data Flow

### 1. User Interaction Flow
1. User accesses web interface
2. Frontend authenticates via OAuth2 with backend
3. User requests current status or control action
4. Frontend sends HTMX request to backend
5. Backend queries Akiles API or database
6. Backend returns response
7. Frontend updates UI via HTMX

### 2. Temperature Monitoring Flow
1. External temperature sensor POSTs data to backend
2. Backend authenticates via HTTP Basic Auth
3. Backend stores temperature reading in database
4. Frontend periodically requests temperature updates
5. Backend returns latest temperature data

### 3. Control Flow (Heating/Hot Water)
1. User clicks control button in frontend
2. Frontend sends HTMX POST to backend
3. Backend authenticates with Akiles API
4. Backend sends gadget action request to Akiles
5. Akiles API executes command on physical hardware
6. Backend returns success/failure status
7. Frontend updates UI state

## Security Considerations

### Authentication & Authorization
- **Frontend-Backend**: OAuth2 flow with secure token handling
- **Backend-Akiles**: OAuth2 client credentials flow
- **Temperature Input**: HTTP Basic Auth with environment-configured credentials
- **Session Management**: Secure cookie handling with HTTPS

### Data Protection
- **In-Transit**: HTTPS for all communications
- **At-Rest**: Database encryption for temperature data
- **Credentials**: Environment variables for all sensitive configuration

### Access Control
- **Frontend**: Authentication required for all operations
- **Temperature API**: Basic auth protection
- **Admin Functions**: Separate authentication if needed

## Environment Configuration

### Backend Environment Variables
```
# Akiles Integration
AKILES_CLIENT_ID=<oauth2_client_id>
AKILES_CLIENT_SECRET=<oauth2_client_secret>
AKILES_HEATING_GADGET_ID=<heating_gadget_id>
AKILES_HOTWATER_GADGET_ID=<hotwater_gadget_id>

# Temperature Sensor Auth
TEMP_SENSOR_USERNAME=<basic_auth_username>
TEMP_SENSOR_PASSWORD=<basic_auth_password>

# Application Config
DATABASE_URL=<database_connection_string>
PORT=<server_port>
OAUTH_REDIRECT_URL=<oauth_callback_url>
```

## Scalability & Performance

### Database Design
- **PostgreSQL**: Single database for all persistence needs
- **Temperature Readings**: Time-series data with proper indexing and partitioning
- **Session Storage**: Stored in PostgreSQL with automatic cleanup
- **Connection Pooling**: PostgreSQL connection pooling configured for expected load

### Caching Strategy
- **Gadget Status**: Short-term caching to reduce Akiles API calls
- **Temperature Data**: Recent readings cached in memory
- **Static Assets**: HTTP caching headers for frontend resources

### Monitoring & Logging
- **API Metrics**: Request/response times and error rates
- **System Health**: Database connectivity and Akiles API availability
- **Security Events**: Failed authentication attempts and unusual patterns