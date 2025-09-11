# 🏔️ Hyttekos

A cozy cabin management system for remote control of heating, hot water, and temperature monitoring.

![Hyttekos Dashboard](docs/ui-mockup.svg)

## Features

- 🔥 **Heating Control** - Turn cabin heating on/off remotely
- 💧 **Hot Water Management** - Control hot water system
- 🌡️ **Temperature Monitoring** - Real-time temperature readings
- 📱 **Responsive Design** - Works on mobile and desktop
- 🔐 **Secure Authentication** - OAuth2 integration with Akiles
- 🏠 **Cozy Interface** - Cabin-themed UI for that warm feeling

## Architecture

Hyttekos follows a simple but robust architecture:

- **Backend**: Go with Gin framework
- **Database**: PostgreSQL for temperature data and sessions
- **Frontend**: HTML with HTMX for dynamic updates, minimal JavaScript
- **Authentication**: OAuth2 with Akiles API
- **Deployment**: Docker container deployable to fly.io

See [docs/architecture.md](docs/architecture.md) for detailed architecture information.

## Quick Start

### Prerequisites

- Go 1.21 or later
- Docker and Docker Compose (for local development)
- PostgreSQL (optional, can use Docker)
- Akiles API credentials

### Local Development with Docker

1. **Clone and setup**:
   ```bash
   git clone <repository>
   cd hyttekos
   cp .env.example .env
   # Edit .env with your Akiles API credentials
   ```

2. **Start development environment**:
   ```bash
   make start
   ```

3. **Access the application**:
   - Web Interface: http://localhost:8080
   - Database: postgres://hyttekos:hyttekos@localhost:5432/hyttekos

### Local Development without Docker

1. **Setup PostgreSQL database**:
   ```bash
   make setup-db
   ```

2. **Configure environment**:
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

3. **Install dependencies and run**:
   ```bash
   make deps
   make dev
   ```

### Available Commands

Run `make help` to see all available commands:

```bash
# Development
make dev          # Run application in development mode
make build        # Build the application
make test         # Run tests

# Docker
make docker-up    # Start development environment
make docker-down  # Stop development environment
make docker-logs  # View container logs

# Database
make migrate-up   # Run database migrations
make migrate-down # Rollback database migrations

# Deployment
make deploy       # Deploy to fly.io
```

## Configuration

All configuration is done via environment variables. See `.env.example` for all required variables:

### Required Variables

- `DATABASE_URL` - PostgreSQL connection string
- `AKILES_CLIENT_ID` - Akiles OAuth2 client ID
- `AKILES_CLIENT_SECRET` - Akiles OAuth2 client secret
- `AKILES_HEATING_GADGET_ID` - Heating system gadget ID
- `AKILES_HOTWATER_GADGET_ID` - Hot water system gadget ID
- `TEMP_SENSOR_USERNAME` - Temperature sensor basic auth username
- `TEMP_SENSOR_PASSWORD` - Temperature sensor basic auth password
- `SESSION_SECRET` - Session encryption key

## API Documentation

See [docs/api-specification.md](docs/api-specification.md) for complete API documentation.

### Key Endpoints

- `GET /` - Main dashboard (requires authentication)
- `GET /auth/login` - Initiate OAuth2 login
- `GET /api/status` - Get system status (JSON)
- `POST /api/heating` - Control heating system
- `POST /api/hot-water` - Control hot water system
- `POST /api/temperature` - Submit temperature reading (Basic Auth)

### Temperature Sensor Integration

To submit temperature readings:

```bash
curl -X POST \
  -u "sensor_username:sensor_password" \
  -H "Content-Type: application/json" \
  -d '{"temperature": 18.5}' \
  https://your-app.fly.dev/api/temperature
```

## Deployment to fly.io

1. **Install flyctl**:
   ```bash
   curl -L https://fly.io/install.sh | sh
   ```

2. **Login to fly.io**:
   ```bash
   flyctl auth login
   ```

3. **Deploy**:
   ```bash
   make deploy
   ```

4. **Set secrets**:
   ```bash
   flyctl secrets set AKILES_CLIENT_ID=your_client_id
   flyctl secrets set AKILES_CLIENT_SECRET=your_client_secret
   flyctl secrets set DATABASE_URL=your_aiven_postgres_url
   # ... other secrets
   ```

### Aiven PostgreSQL

The application is configured to work with Aiven PostgreSQL. Set the `DATABASE_URL` secret to your Aiven connection string:

```bash
flyctl secrets set DATABASE_URL="postgres://username:password@host:port/database?sslmode=require"
```

## Temperature Sensors

The system accepts temperature readings via HTTP POST to `/api/temperature` with Basic Authentication. Example Python script for a Raspberry Pi sensor:

```python
import requests
import time
from datetime import datetime

def read_temperature():
    # Your temperature sensor code here
    return 18.5

def submit_temperature(temp):
    url = "https://your-app.fly.dev/api/temperature"
    auth = ("sensor_username", "sensor_password")
    data = {
        "temperature": temp,
        "timestamp": datetime.now().isoformat() + "Z"
    }
    
    try:
        response = requests.post(url, json=data, auth=auth, timeout=10)
        response.raise_for_status()
        print(f"✅ Temperature {temp}°C submitted successfully")
    except requests.RequestException as e:
        print(f"❌ Failed to submit temperature: {e}")

# Submit temperature every minute
while True:
    temp = read_temperature()
    submit_temperature(temp)
    time.sleep(60)
```

## Development

### Project Structure

```
hyttekos/
├── cmd/server/          # Application entry point
├── internal/            # Internal packages
│   ├── akiles/         # Akiles API client
│   ├── config/         # Configuration management
│   ├── database/       # Database setup and migrations
│   ├── handlers/       # HTTP handlers
│   ├── middleware/     # HTTP middleware
│   ├── models/         # Data models
│   └── repository/     # Data access layer
├── web/                # Frontend assets
│   ├── static/css/     # Stylesheets
│   └── templates/      # HTML templates
├── migrations/         # Database migrations
├── docs/              # Design documents
└── deploy.sh          # Deployment script
```

### Adding Features

1. **Database Changes**: Add migration files in `migrations/`
2. **API Endpoints**: Add handlers in `internal/handlers/`
3. **UI Components**: Add templates in `web/templates/`
4. **Styles**: Update `web/static/css/styles.css`

### Testing the UI

The UI is designed to work without real Akiles credentials for development. You can:

1. Start the development environment
2. Mock the Akiles API responses by modifying the handlers
3. Submit test temperature data via curl
4. Test the responsive design on different screen sizes

## Security

- All external communications use HTTPS
- OAuth2 for user authentication
- HTTP Basic Auth for sensor endpoints
- CSRF protection for web forms
- Secure session management
- Security headers for XSS/clickjacking protection

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test thoroughly
5. Submit a pull request

## License

MIT License - see LICENSE file for details.

## Support

For issues and questions:
1. Check the documentation in the `docs/` folder
2. Create an issue on GitHub
3. Check the logs: `make docker-logs` or `flyctl logs`

---

*Built with ❤️ for cozy cabin management*