#!/bin/bash

# Quick start script for the HTMX web frontend
echo "🌿 Starting LandscapePro HTMX Web Frontend"
echo "=========================================="
echo ""
echo "Features available:"
echo "  ✅ HTMX dynamic content loading"
echo "  ✅ Alpine.js interactive components"
echo "  ✅ TailwindCSS responsive design"
echo "  ✅ Modal forms and booking system"
echo "  ✅ Quote calculator"
echo "  ✅ Service browsing and details"
echo "  ✅ Toast notifications"
echo "  ✅ Form validation"
echo ""
echo "🚀 Server starting on http://localhost:3000"
echo "📱 Fully responsive design"
echo "⚡ Server-side rendering with minimal JavaScript"
echo ""
echo "Press Ctrl+C to stop the server"
echo "=========================================="

# Start the web server
go run backend/cmd/web/main.go