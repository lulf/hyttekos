# Hyttekos Data Flow Diagrams

## Overview

This document describes the data flows within the Hyttekos system, showing how information moves between components and external services.

## 1. User Authentication Flow

```
┌─────────────┐    1. GET /          ┌─────────────┐
│   Browser   │───────────────────►│   Backend   │
│             │                     │             │
│             │◄─────────2. Login───│             │
│             │     redirect        │             │
└─────────────┘                     └─────────────┘
       │                                   │
       │ 3. OAuth2 redirect                │
       │                                   │
       ▼                                   ▼
┌─────────────┐                    ┌─────────────┐
│   Akiles    │                    │             │
│   OAuth2    │                    │  Session    │
│  Provider   │                    │  Storage    │
│             │                    │             │
└─────────────┘                    └─────────────┘
       │
       │ 4. User login & consent
       │
       ▼
┌─────────────┐
│   Browser   │
│             │
│ 5. Callback │
│    with code│
└─────────────┘
       │
       │ 6. GET /auth/callback?code=...
       │
       ▼
┌─────────────┐    7. Exchange code  ┌─────────────┐
│   Backend   │────────────────────►│   Akiles    │
│             │      for tokens      │    API      │
│             │◄────────────────────│             │
│             │    8. Access token   └─────────────┘
└─────────────┘
       │
       │ 9. Set session cookie
       │    & redirect to app
       ▼
┌─────────────┐
│   Browser   │
│ (Logged in) │
└─────────────┘
```

## 2. Dashboard Status Retrieval Flow

```
┌─────────────┐  1. GET /htmx/dashboard  ┌─────────────┐
│   Browser   │─────────────────────────►│   Backend   │
│   (HTMX)    │                          │             │
└─────────────┘                          └──────┬──────┘
       ▲                                        │
       │                                        │ 2. Check session
       │                                        ▼
       │                                 ┌─────────────┐
       │                                 │  Session    │
       │                                 │  Storage    │
       │                                 └──────┬──────┘
       │                                        │ 3. Valid session
       │                                        ▼
       │                                 ┌─────────────┐
       │     8. HTML fragments           │   Backend   │
       │                                 │  (Process)  │
       │                                 └──────┬──────┘
       │                                        │
       │    ┌─────────────┐                     │ 4. Parallel requests
       │    │Temperature  │◄────────────────────┤
       │    │ Database    │  5. Latest temp     │
       │    └─────────────┘                     │
       │                                        │
       │                              6. Get gadget states
       │                                        │
       │                                        ▼
       │                                 ┌─────────────┐
       │                                 │   Akiles    │
       │◄────────────────────────────────│    API      │
       │          7. Gadget status       │             │
       │                                 └─────────────┘
       │
       ▼
┌─────────────┐
│   Browser   │
│  (Updated   │
│  Dashboard) │
└─────────────┘
```

## 3. Control Action Flow (Heating/Hot Water Toggle)

```
┌─────────────┐  1. Click toggle button  ┌─────────────┐
│   Browser   │─────────────────────────►│   Backend   │
│   (HTMX)    │    POST /htmx/heating/   │             │
└─────────────┘         toggle           └──────┬──────┘
       ▲                                        │
       │                                        │ 2. Authenticate
       │                                        │    & validate
       │                                        ▼
       │                                 ┌─────────────┐
       │                                 │  Session    │
       │                                 │  Storage    │
       │                                 └──────┬──────┘
       │                                        │ 3. Valid session
       │                                        ▼
       │                                 ┌─────────────┐
       │                                 │   Backend   │
       │                                 │  (Process)  │
       │                                 └──────┬──────┘
       │                                        │
       │                                        │ 4. Send gadget action
       │                                        ▼
       │                                 ┌─────────────┐
       │                                 │   Akiles    │
       │                                 │    API      │
       │                                 └──────┬──────┘
       │                                        │
       │                                        │ 5. Execute action
       │                                        │    on hardware
       │                                        ▼
       │                                 ┌─────────────┐
       │                                 │   Physical  │
       │                                 │   Heating   │
       │                                 │  Hardware   │
       │                                 └──────┬──────┘
       │                                        │
       │                                        │ 6. Status update
       │                                        ▼
       │                                 ┌─────────────┐
       │     9. Updated HTML fragment    │   Akiles    │
       │                                 │    API      │
       │◄────────────────────────────────│             │
       │          8. Success response    └──────┬──────┘
       │                                        │
       │                                        │ 7. Return success
       │                                        ▼
       │                                 ┌─────────────┐
       │                                 │   Backend   │
       │                                 │             │
       └─────────────────────────────────│             │
                                         └─────────────┘
```

## 4. Temperature Data Collection Flow

```
┌─────────────┐  1. HTTP POST           ┌─────────────┐
│Temperature  │    /api/temperature     │   Backend   │
│   Sensor    │────────────────────────►│             │
│             │  Basic Auth + JSON      └──────┬──────┘
└─────────────┘                                │
                                               │ 2. Authenticate
                                               │    Basic Auth
                                               ▼
                                        ┌─────────────┐
                                        │   Backend   │
                                        │   (Auth)    │
                                        └──────┬──────┘
                                               │ 3. Valid credentials
                                               ▼
                                        ┌─────────────┐
                                        │   Backend   │
                                        │ (Validate & │
                                        │   Store)    │
                                        └──────┬──────┘
                                               │
                                               │ 4. INSERT temperature
                                               │    reading with timestamp
                                               ▼
                                        ┌─────────────┐
                                        │PostgreSQL   │
                                        │  Database   │
                                        │             │
                                        └─────────────┘
```

## 5. Real-time Temperature Display Flow

```
┌─────────────┐                       ┌─────────────┐
│   Browser   │  1. Periodic HTMX     │   Backend   │
│   (HTMX)    │     request every     │             │
│             │     30 seconds        │             │
└─────────────┘  GET /htmx/temp       └──────┬──────┘
       ▲                                     │
       │                                     │ 2. Query latest
       │                                     │    temperature
       │                                     ▼
       │                              ┌─────────────┐
       │                              │Temperature  │
       │                              │  Database   │
       │                              │             │
       │                              └──────┬──────┘
       │                                     │
       │                                     │ 3. Latest reading
       │                                     ▼
       │                              ┌─────────────┐
       │    5. HTML temperature       │   Backend   │
       │       display fragment       │  (Format)   │
       │◄─────────────────────────────│             │
       │    4. Render template        └─────────────┘
       │
       ▼
┌─────────────┐
│   Browser   │
│ (Updated    │
│ Temperature)│
└─────────────┘
```

## 6. Error Handling Flow

```
┌─────────────┐  1. User action         ┌─────────────┐
│   Browser   │────────────────────────►│   Backend   │
│             │                         │             │
└─────────────┘                         └──────┬──────┘
       ▲                                       │
       │                                       │ 2. Process request
       │                                       ▼
       │                                ┌─────────────┐
       │                                │   Akiles    │
       │                                │    API      │
       │                                └──────┬──────┘
       │                                       │
       │                                       │ 3. API Error
       │                                       │    (timeout/500)
       │                                       ▼
       │                                ┌─────────────┐
       │     6. Error HTML fragment     │   Backend   │
       │                                │ (Error      │
       │◄───────────────────────────────│ Handling)   │
       │   5. Log error & format        └─────────────┘
       │      user-friendly response           │
       │                                       │ 4. Log error details
       │                                       ▼
       │                                ┌─────────────┐
       │                                │   Logger/   │
       │                                │ Monitoring  │
       │                                │             │
       └────────────────────────────────│             │
                                        └─────────────┘
```

## 7. Session Management Flow

```
┌─────────────┐                        ┌─────────────┐
│   Browser   │   1. Request with      │   Backend   │
│             │      session cookie    │             │
└─────────────┘                        └──────┬──────┘
       ▲                                      │
       │                                      │ 2. Validate session
       │                                      ▼
       │                               ┌─────────────┐
       │                               │   Session   │
       │                               │   Store     │
       │                               │ (Redis/DB)  │
       │                               └──────┬──────┘
       │                                      │
       │                                      │ 3. Session status
       │                                      ▼
       │                               ┌─────────────┐
       │                               │   Backend   │
       │                               │ (Decision)  │
       │                               └──────┬──────┘
       │                                      │
       │                      ┌───────────────┼───────────────┐
       │                      │ Valid         │ Invalid       │
       │                      ▼               ▼               ▼
       │               ┌─────────────┐ ┌─────────────┐ ┌─────────────┐
       │               │  Continue   │ │   Refresh   │ │  Redirect   │
       │               │  Request    │ │   Token     │ │  to Login   │
       │               │  Processing │ │             │ │             │
       │               └─────────────┘ └─────────────┘ └─────────────┘
       │                      │               │               │
       │                      │               │ 4. Get new    │
       │                      │               │    tokens     │
       │                      │               ▼               │
       │                      │        ┌─────────────┐        │
       │                      │        │   Akiles    │        │
       │                      │        │  OAuth2     │        │
       │                      │        │             │        │
       │                      │        └─────────────┘        │
       │                      │               │               │
       │     5. Response      │               │               │
       │◄─────────────────────┼───────────────┼───────────────┘
                              │               │
                              ▼               ▼
                       ┌─────────────┐ ┌─────────────┐
                       │ Successful  │ │  Login      │
                       │ Response    │ │  Redirect   │
                       └─────────────┘ └─────────────┘
```

## Data Persistence Patterns

### Temperature Data Schema
```sql
CREATE TABLE temperature_readings (
    id SERIAL PRIMARY KEY,
    temperature DECIMAL(4,2) NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_temperature_timestamp ON temperature_readings(timestamp DESC);
```

### Session Data Schema
```sql
CREATE TABLE sessions (
    id VARCHAR(255) PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    access_token TEXT NOT NULL,
    refresh_token TEXT NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);
```

## Performance Considerations

### Caching Strategy
- **Gadget Status**: Cache for 30 seconds to reduce Akiles API calls
- **Temperature Readings**: Cache latest reading for 60 seconds
- **Session Validation**: Cache valid sessions for 5 minutes

### Database Optimization
- **Temperature History**: Partition by month for large datasets
- **Index Strategy**: Time-based indexes for efficient querying
- **Connection Pooling**: Limit concurrent database connections