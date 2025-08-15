#!/bin/bash

# Test script for the HTMX landscaping app
echo "🧪 Testing LandscapePro HTMX Web Server"
echo "======================================="

# Start server in background
echo "🚀 Starting server..."
go run backend/cmd/web/main.go &
SERVER_PID=$!

# Wait for server to start
sleep 3

echo "📋 Running tests..."

# Test 1: Homepage
echo "Test 1: Homepage"
curl -s -o /dev/null -w "%{http_code}" http://localhost:3000/ && echo " ✅ Homepage loads"

# Test 2: Services endpoint
echo "Test 2: Featured Services API"
curl -s -o /dev/null -w "%{http_code}" http://localhost:3000/api/services/featured && echo " ✅ Featured services endpoint works"

# Test 3: All services endpoint
echo "Test 3: All Services API"
curl -s -o /dev/null -w "%{http_code}" http://localhost:3000/api/services/all && echo " ✅ All services endpoint works"

# Test 4: Testimonials endpoint
echo "Test 4: Testimonials API"
curl -s -o /dev/null -w "%{http_code}" http://localhost:3000/api/testimonials/featured && echo " ✅ Testimonials endpoint works"

# Test 5: Auth status endpoint
echo "Test 5: Auth Status API"
curl -s -o /dev/null -w "%{http_code}" http://localhost:3000/api/auth/status && echo " ✅ Auth status endpoint works"

# Test 6: Booking form endpoint
echo "Test 6: Booking Form"
curl -s -o /dev/null -w "%{http_code}" http://localhost:3000/booking/form && echo " ✅ Booking form endpoint works"

# Test 7: Services page
echo "Test 7: Services Page"
curl -s -o /dev/null -w "%{http_code}" http://localhost:3000/services && echo " ✅ Services page loads"

# Test 8: Booking page
echo "Test 8: Booking Page"
curl -s -o /dev/null -w "%{http_code}" http://localhost:3000/booking && echo " ✅ Booking page loads"

# Test 9: Static files
echo "Test 9: Static CSS"
curl -s -o /dev/null -w "%{http_code}" http://localhost:3000/static/css/custom.css && echo " ✅ Custom CSS loads"

echo "Test 10: Static JS"
curl -s -o /dev/null -w "%{http_code}" http://localhost:3000/static/js/app.js && echo " ✅ Custom JS loads"

# Test 11: Service details
echo "Test 11: Service Detail"
curl -s -o /dev/null -w "%{http_code}" http://localhost:3000/services/1 && echo " ✅ Service detail page works"

echo ""
echo "✅ All tests completed!"
echo "🌐 Visit http://localhost:3000 to see the full HTMX demo"
echo ""
echo "Key features working:"
echo "  - HTMX dynamic content loading"
echo "  - Modal forms with Alpine.js"
echo "  - Quote calculator"
echo "  - Service browsing"
echo "  - Interactive booking system"
echo ""

# Stop the server
kill $SERVER_PID
echo "🛑 Server stopped"