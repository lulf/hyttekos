#!/bin/bash

# Deploy script for Hyttekos to fly.io
set -e

echo "🏔️ Deploying Hyttekos to fly.io..."

# Check if flyctl is installed
if ! command -v flyctl &> /dev/null; then
    echo "❌ flyctl is not installed. Please install it first:"
    echo "   curl -L https://fly.io/install.sh | sh"
    exit 1
fi

# Check if logged in to fly.io
if ! flyctl auth whoami &> /dev/null; then
    echo "❌ Not logged in to fly.io. Please login first:"
    echo "   flyctl auth login"
    exit 1
fi

# Build and deploy
echo "🔨 Building and deploying application..."
flyctl deploy --config fly.toml

# Show status
echo "📊 Checking deployment status..."
flyctl status

echo "✅ Deployment complete!"
echo ""
echo "📱 Your Hyttekos application is now available at:"
flyctl info --name hyttekos | grep -E "Hostname.*\.fly\.dev" | awk '{print "   https://" $2}'
echo ""
echo "💾 Creating persistent volume (if it doesn't exist)..."
if ! flyctl volumes list | grep -q "hyttekos_data"; then
    echo "   Creating new volume 'hyttekos_data'..."
    flyctl volumes create hyttekos_data --region arn --size 1
else
    echo "   Volume 'hyttekos_data' already exists"
fi

echo ""
echo "🔧 To set up secrets (if not done already):"
echo "   flyctl secrets set AKILES_CLIENT_ID=your_client_id"
echo "   flyctl secrets set AKILES_CLIENT_SECRET=your_client_secret"
echo "   flyctl secrets set AKILES_ORGANIZATION_ID=your_org_id"
echo "   flyctl secrets set AKILES_HEATING_GADGET_ID=your_heating_gadget_id"
echo "   flyctl secrets set AKILES_HOTWATER_GADGET_ID=your_hotwater_gadget_id"
echo "   flyctl secrets set TEMP_SENSOR_USERNAME=your_sensor_username"
echo "   flyctl secrets set TEMP_SENSOR_PASSWORD=your_sensor_password"
echo "   flyctl secrets set SESSION_SECRET=your_session_secret"
echo "   flyctl secrets set OAUTH_REDIRECT_URL=https://your-app.fly.dev/auth/callback"
echo ""
echo "📋 To view logs:"
echo "   flyctl logs"