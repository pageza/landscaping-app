package services

import (
	"bytes"
	"fmt"
	"html/template"
	"path/filepath"

	"github.com/pageza/landscaping-app/web/internal/config"
)

type TemplateService struct {
	config    *config.Config
	templates map[string]*template.Template
	funcs     template.FuncMap
}

type TemplateData struct {
	Title       string
	User        *User
	CSRFToken   string
	Flash       map[string]string
	Data        interface{}
	Request     interface{}
	IsHTMX      bool
	HXTarget    string
	HXCurrentURL string
}

func NewTemplateService(cfg *config.Config) (*TemplateService, error) {
	ts := &TemplateService{
		config:    cfg,
		templates: make(map[string]*template.Template),
		funcs:     createTemplateFuncs(),
	}

	// In a real implementation, we'd parse actual template files
	// For now, we'll create inline templates
	if err := ts.loadTemplates(); err != nil {
		return nil, err
	}

	return ts, nil
}

func (ts *TemplateService) loadTemplates() error {
	// Define inline templates for now
	// In production, these would be loaded from files
	templates := map[string]string{
		"base.html": `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}} - Landscaping Management</title>
    <script src="https://unpkg.com/htmx.org@1.9.10"></script>
    <script src="https://unpkg.com/alpinejs@3.x.x/dist/cdn.min.js" defer></script>
    <script src="https://cdn.tailwindcss.com"></script>
    <link rel="preconnect" href="https://fonts.googleapis.com">
    <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
    <link href="https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600;700&display=swap" rel="stylesheet">
    <style>
        body { font-family: 'Inter', sans-serif; }
        [x-cloak] { display: none !important; }
    </style>
</head>
<body class="bg-gray-50">
    {{template "content" .}}
</body>
</html>`,

		"login.html": `{{define "content"}}
<div class="min-h-screen flex items-center justify-center py-12 px-4 sm:px-6 lg:px-8">
    <div class="max-w-md w-full space-y-8">
        <div>
            <h2 class="mt-6 text-center text-3xl font-extrabold text-gray-900">
                Sign in to your account
            </h2>
        </div>
        <form class="mt-8 space-y-6" hx-post="/login" hx-target="#login-form" hx-swap="outerHTML">
            <div id="login-form" class="rounded-md shadow-sm -space-y-px">
                {{if .Flash.error}}
                <div class="bg-red-50 border border-red-200 rounded-md p-4 mb-4">
                    <div class="text-sm text-red-700">{{.Flash.error}}</div>
                </div>
                {{end}}
                <div>
                    <input id="email" name="email" type="email" required 
                           class="relative block w-full px-3 py-2 border border-gray-300 rounded-t-md placeholder-gray-500 text-gray-900 focus:outline-none focus:ring-blue-500 focus:border-blue-500" 
                           placeholder="Email address">
                </div>
                <div>
                    <input id="password" name="password" type="password" required 
                           class="relative block w-full px-3 py-2 border border-gray-300 rounded-b-md placeholder-gray-500 text-gray-900 focus:outline-none focus:ring-blue-500 focus:border-blue-500" 
                           placeholder="Password">
                </div>
            </div>
            <div>
                <button type="submit" 
                        class="group relative w-full flex justify-center py-2 px-4 border border-transparent text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500">
                    Sign in
                </button>
            </div>
            <div class="text-center">
                <a href="/register" class="text-blue-600 hover:text-blue-500">
                    Don't have an account? Sign up
                </a>
            </div>
        </form>
    </div>
</div>
{{end}}`,

		"dashboard.html": `{{define "content"}}
<div class="min-h-screen bg-gray-50">
    {{template "sidebar" .}}
    <div class="lg:pl-64">
        {{template "header" .}}
        <main class="py-10">
            <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
                <h1 class="text-3xl font-bold text-gray-900">Dashboard</h1>
                <div class="mt-8 grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
                    <!-- KPI Cards -->
                    <div class="bg-white overflow-hidden shadow rounded-lg">
                        <div class="p-5">
                            <div class="flex items-center">
                                <div class="flex-shrink-0">
                                    <div class="w-8 h-8 bg-blue-500 rounded-md flex items-center justify-center">
                                        <span class="text-white text-sm font-medium">J</span>
                                    </div>
                                </div>
                                <div class="ml-5 w-0 flex-1">
                                    <dl>
                                        <dt class="text-sm font-medium text-gray-500 truncate">Active Jobs</dt>
                                        <dd class="text-lg font-medium text-gray-900">24</dd>
                                    </dl>
                                </div>
                            </div>
                        </div>
                    </div>
                    <!-- More KPI cards... -->
                </div>
            </div>
        </main>
    </div>
</div>
{{end}}`,

		"components/sidebar.html": `{{define "sidebar"}}
<div class="hidden lg:flex lg:w-64 lg:flex-col lg:fixed lg:inset-y-0" id="desktop-sidebar">
    <div class="flex-1 flex flex-col min-h-0 bg-gray-800">
        <div class="flex-1 flex flex-col pt-5 pb-4 overflow-y-auto">
            <div class="flex items-center flex-shrink-0 px-4">
                <div class="w-8 h-8 bg-green-600 rounded-lg flex items-center justify-center mr-3">
                    <svg class="w-5 h-5 text-white" fill="currentColor" viewBox="0 0 20 20">
                        <path fill-rule="evenodd" d="M3 3a1 1 0 000 2v8a2 2 0 002 2h2.586l-1.293 1.293a1 1 0 101.414 1.414L10 15.414l2.293 2.293a1 1 0 001.414-1.414L12.414 15H15a2 2 0 002-2V5a1 1 0 100-2H3zm11.707 4.707a1 1 0 00-1.414-1.414L10 9.586 8.707 8.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd"/>
                    </svg>
                </div>
                <h1 class="text-white text-xl font-bold">LandscapeApp</h1>
            </div>
            <nav class="mt-5 flex-1 px-2 space-y-1">
                <!-- Dashboard -->
                <a href="/admin" class="nav-item text-gray-300 hover:bg-gray-700 hover:text-white group flex items-center px-2 py-2 text-sm font-medium rounded-md">
                    <svg class="text-gray-400 group-hover:text-gray-300 mr-3 flex-shrink-0 h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2H5a2 2 0 00-2-2z"/>
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 5a2 2 0 012-2h4a2 2 0 012 2v2a2 2 0 002 2H6a2 2 0 002-2V5z"/>
                    </svg>
                    Dashboard
                </a>

                <!-- Customers -->
                <a href="/admin/customers" class="nav-item text-gray-300 hover:bg-gray-700 hover:text-white group flex items-center px-2 py-2 text-sm font-medium rounded-md">
                    <svg class="text-gray-400 group-hover:text-gray-300 mr-3 flex-shrink-0 h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197m13.5-9a2.5 2.5 0 11-5 0 2.5 2.5 0 015 0z"/>
                    </svg>
                    Customers
                </a>

                <!-- Properties -->
                <a href="/admin/properties" class="nav-item text-gray-300 hover:bg-gray-700 hover:text-white group flex items-center px-2 py-2 text-sm font-medium rounded-md">
                    <svg class="text-gray-400 group-hover:text-gray-300 mr-3 flex-shrink-0 h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4"/>
                    </svg>
                    Properties
                </a>

                <!-- Jobs -->
                <a href="/admin/jobs" class="nav-item text-gray-300 hover:bg-gray-700 hover:text-white group flex items-center px-2 py-2 text-sm font-medium rounded-md">
                    <svg class="text-gray-400 group-hover:text-gray-300 mr-3 flex-shrink-0 h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3a2 2 0 012-2h4a2 2 0 012 2v4m-6 9l2 2 4-4m6-7H2a1 1 0 00-1 1v10a1 1 0 001 1h20a1 1 0 001-1V8a1 1 0 00-1-1z"/>
                    </svg>
                    Jobs
                </a>

                <!-- Calendar -->
                <a href="/admin/jobs/calendar" class="nav-item text-gray-300 hover:bg-gray-700 hover:text-white group flex items-center px-2 py-2 text-sm font-medium rounded-md">
                    <svg class="text-gray-400 group-hover:text-gray-300 mr-3 flex-shrink-0 h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3a2 2 0 012-2h4a2 2 0 012 2v4m-6 9l2 2 4-4m6-7H2a1 1 0 00-1 1v10a1 1 0 001 1h20a1 1 0 001-1V8a1 1 0 00-1-1z"/>
                    </svg>
                    Calendar
                </a>

                <!-- Quotes -->
                <a href="/admin/quotes" class="nav-item text-gray-300 hover:bg-gray-700 hover:text-white group flex items-center px-2 py-2 text-sm font-medium rounded-md">
                    <svg class="text-gray-400 group-hover:text-gray-300 mr-3 flex-shrink-0 h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"/>
                    </svg>
                    Quotes
                </a>

                <!-- Invoices -->
                <a href="/admin/invoices" class="nav-item text-gray-300 hover:bg-gray-700 hover:text-white group flex items-center px-2 py-2 text-sm font-medium rounded-md">
                    <svg class="text-gray-400 group-hover:text-gray-300 mr-3 flex-shrink-0 h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
                    </svg>
                    Invoices
                </a>

                <!-- Equipment -->
                <a href="/admin/equipment" class="nav-item text-gray-300 hover:bg-gray-700 hover:text-white group flex items-center px-2 py-2 text-sm font-medium rounded-md">
                    <svg class="text-gray-400 group-hover:text-gray-300 mr-3 flex-shrink-0 h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19.428 15.428a2 2 0 00-1.022-.547l-2.387-.477a6 6 0 00-3.86.517l-.318.158a6 6 0 01-3.86.517L6.05 15.21a2 2 0 00-1.806.547M8 4h8l-1 1v5.172a2 2 0 00.586 1.414l5 5c1.26 1.26.367 3.414-1.415 3.414H4.828c-1.782 0-2.674-2.154-1.414-3.414l5-5A2 2 0 009 10.172V5L8 4z"/>
                    </svg>
                    Equipment
                </a>

                <!-- Reports -->
                <a href="/admin/reports" class="nav-item text-gray-300 hover:bg-gray-700 hover:text-white group flex items-center px-2 py-2 text-sm font-medium rounded-md">
                    <svg class="text-gray-400 group-hover:text-gray-300 mr-3 flex-shrink-0 h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"/>
                    </svg>
                    Reports
                </a>

                <!-- Settings -->
                <a href="/admin/settings" class="nav-item text-gray-300 hover:bg-gray-700 hover:text-white group flex items-center px-2 py-2 text-sm font-medium rounded-md">
                    <svg class="text-gray-400 group-hover:text-gray-300 mr-3 flex-shrink-0 h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"/>
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"/>
                    </svg>
                    Settings
                </a>
            </nav>
        </div>
    </div>
</div>

<!-- Mobile sidebar -->
<div class="lg:hidden" x-data="{ sidebarOpen: false }">
    <!-- Overlay -->
    <div x-show="sidebarOpen" x-transition:enter="transition-opacity ease-linear duration-300" x-transition:enter-start="opacity-0" x-transition:enter-end="opacity-100" x-transition:leave="transition-opacity ease-linear duration-300" x-transition:leave-start="opacity-100" x-transition:leave-end="opacity-0" class="fixed inset-0 z-40 bg-gray-600 bg-opacity-75"></div>

    <!-- Mobile sidebar -->
    <div x-show="sidebarOpen" x-transition:enter="transition ease-in-out duration-300 transform" x-transition:enter-start="-translate-x-full" x-transition:enter-end="translate-x-0" x-transition:leave="transition ease-in-out duration-300 transform" x-transition:leave-start="translate-x-0" x-transition:leave-end="-translate-x-full" class="fixed inset-y-0 left-0 z-50 w-64 bg-gray-800 lg:hidden">
        <div class="flex items-center justify-between px-4 py-3">
            <div class="flex items-center">
                <div class="w-8 h-8 bg-green-600 rounded-lg flex items-center justify-center mr-3">
                    <svg class="w-5 h-5 text-white" fill="currentColor" viewBox="0 0 20 20">
                        <path fill-rule="evenodd" d="M3 3a1 1 0 000 2v8a2 2 0 002 2h2.586l-1.293 1.293a1 1 0 101.414 1.414L10 15.414l2.293 2.293a1 1 0 001.414-1.414L12.414 15H15a2 2 0 002-2V5a1 1 0 100-2H3zm11.707 4.707a1 1 0 00-1.414-1.414L10 9.586 8.707 8.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd"/>
                    </svg>
                </div>
                <h1 class="text-white text-xl font-bold">LandscapeApp</h1>
            </div>
            <button @click="sidebarOpen = false" class="text-gray-300 hover:text-white">
                <svg class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
                </svg>
            </button>
        </div>
        <nav class="mt-2 px-2 space-y-1">
            <!-- Same navigation items as desktop -->
        </nav>
    </div>

    <!-- Mobile menu button -->
    <div class="lg:hidden flex items-center px-4 py-2">
        <button @click="sidebarOpen = true" class="text-gray-600 hover:text-gray-900">
            <svg class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16"/>
            </svg>
        </button>
    </div>
</div>
{{end}}`,

		"components/header.html": `{{define "header"}}
<div class="relative z-10 flex-shrink-0 flex h-16 bg-white shadow">
    <div class="flex-1 px-4 flex justify-between">
        <div class="flex-1 flex">
            <!-- Search bar can go here -->
        </div>
        <div class="ml-4 flex items-center md:ml-6">
            <div class="ml-3 relative">
                <div x-data="{ open: false }">
                    <button @click="open = !open" class="max-w-xs bg-white flex items-center text-sm rounded-full focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500">
                        <span class="sr-only">Open user menu</span>
                        <div class="h-8 w-8 rounded-full bg-gray-300 flex items-center justify-center">
                            <span class="text-sm font-medium text-gray-700">{{if .User}}{{substr .User.FirstName 0 1}}{{end}}</span>
                        </div>
                    </button>
                    <div x-show="open" @click.away="open = false" x-cloak 
                         class="origin-top-right absolute right-0 mt-2 w-48 rounded-md shadow-lg py-1 bg-white ring-1 ring-black ring-opacity-5">
                        <a href="/logout" class="block px-4 py-2 text-sm text-gray-700 hover:bg-gray-100">Sign out</a>
                    </div>
                </div>
            </div>
        </div>
    </div>
</div>
{{end}}`,

		"register.html": `{{define "content"}}
<div class="min-h-screen flex items-center justify-center py-12 px-4 sm:px-6 lg:px-8">
    <div class="max-w-md w-full space-y-8">
        <div>
            <h2 class="mt-6 text-center text-3xl font-extrabold text-gray-900">
                Create your account
            </h2>
        </div>
        <form class="mt-8 space-y-6" hx-post="/register" hx-target="#register-form" hx-swap="outerHTML">
            <div id="register-form" class="space-y-4">
                {{if .Flash.error}}
                <div class="bg-red-50 border border-red-200 rounded-md p-4">
                    <div class="text-sm text-red-700">{{.Flash.error}}</div>
                </div>
                {{end}}
                {{if .Flash.success}}
                <div class="bg-green-50 border border-green-200 rounded-md p-4">
                    <div class="text-sm text-green-700">{{.Flash.success}}</div>
                </div>
                {{end}}
                <div class="grid grid-cols-2 gap-4">
                    <div>
                        <input id="first_name" name="first_name" type="text" required 
                               class="relative block w-full px-3 py-2 border border-gray-300 rounded-md placeholder-gray-500 text-gray-900 focus:outline-none focus:ring-blue-500 focus:border-blue-500" 
                               placeholder="First name">
                    </div>
                    <div>
                        <input id="last_name" name="last_name" type="text" required 
                               class="relative block w-full px-3 py-2 border border-gray-300 rounded-md placeholder-gray-500 text-gray-900 focus:outline-none focus:ring-blue-500 focus:border-blue-500" 
                               placeholder="Last name">
                    </div>
                </div>
                <div>
                    <input id="email" name="email" type="email" required 
                           class="relative block w-full px-3 py-2 border border-gray-300 rounded-md placeholder-gray-500 text-gray-900 focus:outline-none focus:ring-blue-500 focus:border-blue-500" 
                           placeholder="Email address">
                </div>
                <div>
                    <input id="company_name" name="company_name" type="text" 
                           class="relative block w-full px-3 py-2 border border-gray-300 rounded-md placeholder-gray-500 text-gray-900 focus:outline-none focus:ring-blue-500 focus:border-blue-500" 
                           placeholder="Company name (optional)">
                </div>
                <div>
                    <input id="password" name="password" type="password" required 
                           class="relative block w-full px-3 py-2 border border-gray-300 rounded-md placeholder-gray-500 text-gray-900 focus:outline-none focus:ring-blue-500 focus:border-blue-500" 
                           placeholder="Password">
                </div>
                <div>
                    <input id="confirm_password" name="confirm_password" type="password" required 
                           class="relative block w-full px-3 py-2 border border-gray-300 rounded-md placeholder-gray-500 text-gray-900 focus:outline-none focus:ring-blue-500 focus:border-blue-500" 
                           placeholder="Confirm password">
                </div>
            </div>
            <div>
                <button type="submit" 
                        class="group relative w-full flex justify-center py-2 px-4 border border-transparent text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500">
                    Create account
                </button>
            </div>
            <div class="text-center">
                <a href="/login" class="text-blue-600 hover:text-blue-500">
                    Already have an account? Sign in
                </a>
            </div>
        </form>
    </div>
</div>
{{end}}`,

		"login_form.html": `{{define "content"}}
<div id="login-form" class="rounded-md shadow-sm -space-y-px">
    {{if .Flash.error}}
    <div class="bg-red-50 border border-red-200 rounded-md p-4 mb-4">
        <div class="text-sm text-red-700">{{.Flash.error}}</div>
    </div>
    {{end}}
    {{if .Flash.success}}
    <div class="bg-green-50 border border-green-200 rounded-md p-4 mb-4">
        <div class="text-sm text-green-700">{{.Flash.success}}</div>
    </div>
    {{end}}
    <div>
        <input id="email" name="email" type="email" required 
               class="relative block w-full px-3 py-2 border border-gray-300 rounded-t-md placeholder-gray-500 text-gray-900 focus:outline-none focus:ring-blue-500 focus:border-blue-500" 
               placeholder="Email address">
    </div>
    <div>
        <input id="password" name="password" type="password" required 
               class="relative block w-full px-3 py-2 border border-gray-300 rounded-b-md placeholder-gray-500 text-gray-900 focus:outline-none focus:ring-blue-500 focus:border-blue-500" 
               placeholder="Password">
    </div>
</div>
{{end}}`,

		"forgot_password.html": `{{define "content"}}
<div class="min-h-screen flex items-center justify-center py-12 px-4 sm:px-6 lg:px-8">
    <div class="max-w-md w-full space-y-8">
        <div>
            <h2 class="mt-6 text-center text-3xl font-extrabold text-gray-900">
                Forgot your password?
            </h2>
            <p class="mt-2 text-center text-sm text-gray-600">
                Enter your email address and we'll send you a link to reset your password.
            </p>
        </div>
        <form class="mt-8 space-y-6" hx-post="/forgot-password" hx-target="#forgot-form" hx-swap="outerHTML">
            <div id="forgot-form">
                {{if .Flash.error}}
                <div class="bg-red-50 border border-red-200 rounded-md p-4 mb-4">
                    <div class="text-sm text-red-700">{{.Flash.error}}</div>
                </div>
                {{end}}
                {{if .Flash.success}}
                <div class="bg-green-50 border border-green-200 rounded-md p-4 mb-4">
                    <div class="text-sm text-green-700">{{.Flash.success}}</div>
                </div>
                {{end}}
                <div>
                    <input id="email" name="email" type="email" required 
                           class="relative block w-full px-3 py-2 border border-gray-300 rounded-md placeholder-gray-500 text-gray-900 focus:outline-none focus:ring-blue-500 focus:border-blue-500" 
                           placeholder="Email address">
                </div>
            </div>
            <div>
                <button type="submit" 
                        class="group relative w-full flex justify-center py-2 px-4 border border-transparent text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500">
                    Send reset link
                </button>
            </div>
            <div class="text-center">
                <a href="/login" class="text-blue-600 hover:text-blue-500">
                    Back to login
                </a>
            </div>
        </form>
    </div>
</div>
{{end}}`,

		"reset_password.html": `{{define "content"}}
<div class="min-h-screen flex items-center justify-center py-12 px-4 sm:px-6 lg:px-8">
    <div class="max-w-md w-full space-y-8">
        <div>
            <h2 class="mt-6 text-center text-3xl font-extrabold text-gray-900">
                Reset your password
            </h2>
        </div>
        <form class="mt-8 space-y-6" hx-post="/reset-password" hx-target="#reset-form" hx-swap="outerHTML">
            <input type="hidden" name="token" value="{{.Data.token}}">
            <div id="reset-form" class="space-y-4">
                {{if .Flash.error}}
                <div class="bg-red-50 border border-red-200 rounded-md p-4">
                    <div class="text-sm text-red-700">{{.Flash.error}}</div>
                </div>
                {{end}}
                <div>
                    <input id="password" name="password" type="password" required 
                           class="relative block w-full px-3 py-2 border border-gray-300 rounded-md placeholder-gray-500 text-gray-900 focus:outline-none focus:ring-blue-500 focus:border-blue-500" 
                           placeholder="New password">
                </div>
                <div>
                    <input id="confirm_password" name="confirm_password" type="password" required 
                           class="relative block w-full px-3 py-2 border border-gray-300 rounded-md placeholder-gray-500 text-gray-900 focus:outline-none focus:ring-blue-500 focus:border-blue-500" 
                           placeholder="Confirm new password">
                </div>
            </div>
            <div>
                <button type="submit" 
                        class="group relative w-full flex justify-center py-2 px-4 border border-transparent text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500">
                    Reset password
                </button>
            </div>
        </form>
    </div>
</div>
{{end}}`,

		"admin_dashboard.html": `{{define "content"}}
<div class="min-h-screen bg-gray-50">
    {{template "sidebar" .}}
    <div class="lg:pl-64">
        {{template "header" .}}
        <main class="py-10">
            <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
                <div class="md:flex md:items-center md:justify-between">
                    <div class="flex-1 min-w-0">
                        <h2 class="text-2xl font-bold leading-7 text-gray-900 sm:text-3xl sm:truncate">
                            Dashboard Overview
                        </h2>
                    </div>
                    <div class="mt-4 flex md:mt-0 md:ml-4">
                        <button type="button" 
                                hx-get="/api/v1/dashboard/stats"
                                hx-target="#stats-container"
                                hx-swap="innerHTML"
                                class="ml-3 inline-flex items-center px-4 py-2 border border-gray-300 rounded-md shadow-sm text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500">
                            Refresh
                        </button>
                    </div>
                </div>

                <!-- Stats Grid -->
                <div id="stats-container" class="mt-8">
                    {{template "dashboard_stats" .}}
                </div>

                <!-- Charts and Tables Grid -->
                <div class="mt-8 grid grid-cols-1 gap-8 lg:grid-cols-2">
                    <!-- Recent Jobs -->
                    <div class="bg-white overflow-hidden shadow rounded-lg">
                        <div class="p-6">
                            <h3 class="text-lg leading-6 font-medium text-gray-900 mb-4">Recent Jobs</h3>
                            <div hx-get="/api/v1/jobs/table?limit=5" hx-trigger="load" hx-swap="innerHTML">
                                Loading...
                            </div>
                        </div>
                    </div>

                    <!-- Pending Quotes -->
                    <div class="bg-white overflow-hidden shadow rounded-lg">
                        <div class="p-6">
                            <h3 class="text-lg leading-6 font-medium text-gray-900 mb-4">Pending Quotes</h3>
                            <div hx-get="/api/v1/quotes/table?status=pending&limit=5" hx-trigger="load" hx-swap="innerHTML">
                                Loading...
                            </div>
                        </div>
                    </div>

                    <!-- Revenue Chart -->
                    <div class="bg-white overflow-hidden shadow rounded-lg">
                        <div class="p-6">
                            <h3 class="text-lg leading-6 font-medium text-gray-900 mb-4">Revenue Trend</h3>
                            <div class="h-64 flex items-center justify-center text-gray-500">
                                Chart will be rendered here
                            </div>
                        </div>
                    </div>

                    <!-- Quick Actions -->
                    <div class="bg-white overflow-hidden shadow rounded-lg">
                        <div class="p-6">
                            <h3 class="text-lg leading-6 font-medium text-gray-900 mb-4">Quick Actions</h3>
                            <div class="space-y-3">
                                <a href="/admin/customers/new" 
                                   class="inline-flex w-full items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md text-blue-700 bg-blue-100 hover:bg-blue-200 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500">
                                    Add New Customer
                                </a>
                                <a href="/admin/jobs/new" 
                                   class="inline-flex w-full items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md text-green-700 bg-green-100 hover:bg-green-200 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-green-500">
                                    Schedule Job
                                </a>
                                <a href="/admin/quotes/new" 
                                   class="inline-flex w-full items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md text-purple-700 bg-purple-100 hover:bg-purple-200 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-purple-500">
                                    Create Quote
                                </a>
                                <a href="/admin/invoices/new" 
                                   class="inline-flex w-full items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md text-orange-700 bg-orange-100 hover:bg-orange-200 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-orange-500">
                                    Generate Invoice
                                </a>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </main>
    </div>
</div>
{{end}}`,

		"dashboard_stats.html": `{{define "content"}}
<div class="grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-4">
    <!-- Active Jobs -->
    <div class="bg-white overflow-hidden shadow rounded-lg">
        <div class="p-5">
            <div class="flex items-center">
                <div class="flex-shrink-0">
                    <div class="w-8 h-8 bg-blue-500 rounded-md flex items-center justify-center">
                        <svg class="w-5 h-5 text-white" fill="currentColor" viewBox="0 0 20 20">
                            <path fill-rule="evenodd" d="M6 2a1 1 0 00-1 1v1H4a2 2 0 00-2 2v10a2 2 0 002 2h12a2 2 0 002-2V6a2 2 0 00-2-2h-1V3a1 1 0 10-2 0v1H7V3a1 1 0 00-1-1zm0 5a1 1 0 000 2h8a1 1 0 100-2H6z" clip-rule="evenodd"/>
                        </svg>
                    </div>
                </div>
                <div class="ml-5 w-0 flex-1">
                    <dl>
                        <dt class="text-sm font-medium text-gray-500 truncate">Active Jobs</dt>
                        <dd class="text-lg font-medium text-gray-900">{{.Data.ActiveJobs}}</dd>
                    </dl>
                </div>
            </div>
        </div>
    </div>

    <!-- Pending Quotes -->
    <div class="bg-white overflow-hidden shadow rounded-lg">
        <div class="p-5">
            <div class="flex items-center">
                <div class="flex-shrink-0">
                    <div class="w-8 h-8 bg-yellow-500 rounded-md flex items-center justify-center">
                        <svg class="w-5 h-5 text-white" fill="currentColor" viewBox="0 0 20 20">
                            <path fill-rule="evenodd" d="M4 4a2 2 0 00-2 2v4a2 2 0 002 2V6h10a2 2 0 00-2-2H4zm2 6a2 2 0 012-2h8a2 2 0 012 2v4a2 2 0 01-2 2H8a2 2 0 01-2-2v-4zm6 4a2 2 0 100-4 2 2 0 000 4z" clip-rule="evenodd"/>
                        </svg>
                    </div>
                </div>
                <div class="ml-5 w-0 flex-1">
                    <dl>
                        <dt class="text-sm font-medium text-gray-500 truncate">Pending Quotes</dt>
                        <dd class="text-lg font-medium text-gray-900">{{.Data.PendingQuotes}}</dd>
                    </dl>
                </div>
            </div>
        </div>
    </div>

    <!-- Revenue -->
    <div class="bg-white overflow-hidden shadow rounded-lg">
        <div class="p-5">
            <div class="flex items-center">
                <div class="flex-shrink-0">
                    <div class="w-8 h-8 bg-green-500 rounded-md flex items-center justify-center">
                        <svg class="w-5 h-5 text-white" fill="currentColor" viewBox="0 0 20 20">
                            <path d="M8.433 7.418c.155-.103.346-.196.567-.267v1.698a2.305 2.305 0 01-.567-.267C8.07 8.34 8 8.114 8 8c0-.114.07-.34.433-.582zM11 12.849v-1.698c.22.071.412.164.567.267.364.243.433.468.433.582 0 .114-.07.34-.433.582a2.305 2.305 0 01-.567.267z"/>
                            <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm1-13a1 1 0 10-2 0v.092a4.535 4.535 0 00-1.676.662C6.602 6.234 6 7.009 6 8c0 .99.602 1.765 1.324 2.246.48.32 1.054.545 1.676.662v1.941c-.391-.127-.68-.317-.843-.504a1 1 0 10-1.51 1.31c.562.649 1.413 1.076 2.353 1.253V15a1 1 0 102 0v-.092a4.535 4.535 0 001.676-.662C13.398 13.766 14 12.991 14 12c0-.99-.602-1.765-1.324-2.246A4.535 4.535 0 0011 9.092V7.151c.391.127.68.317.843.504a1 1 0 101.511-1.31c-.563-.649-1.413-1.076-2.354-1.253V5z" clip-rule="evenodd"/>
                        </svg>
                    </div>
                </div>
                <div class="ml-5 w-0 flex-1">
                    <dl>
                        <dt class="text-sm font-medium text-gray-500 truncate">Revenue (YTD)</dt>
                        <dd class="text-lg font-medium text-gray-900">{{formatCurrency .Data.Revenue}}</dd>
                    </dl>
                </div>
            </div>
        </div>
    </div>

    <!-- Customers -->
    <div class="bg-white overflow-hidden shadow rounded-lg">
        <div class="p-5">
            <div class="flex items-center">
                <div class="flex-shrink-0">
                    <div class="w-8 h-8 bg-purple-500 rounded-md flex items-center justify-center">
                        <svg class="w-5 h-5 text-white" fill="currentColor" viewBox="0 0 20 20">
                            <path d="M13 6a3 3 0 11-6 0 3 3 0 016 0zM18 8a2 2 0 11-4 0 2 2 0 014 0zM14 15a4 4 0 00-8 0v3h8v-3z"/>
                        </svg>
                    </div>
                </div>
                <div class="ml-5 w-0 flex-1">
                    <dl>
                        <dt class="text-sm font-medium text-gray-500 truncate">Total Customers</dt>
                        <dd class="text-lg font-medium text-gray-900">{{.Data.Customers}}</dd>
                    </dl>
                </div>
            </div>
        </div>
    </div>
</div>
{{end}}`,

		"customer_portal.html": `{{define "content"}}
<div class="min-h-screen bg-gray-50">
    {{template "customer_sidebar" .}}
    <div class="lg:pl-64">
        {{template "header" .}}
        <main class="py-10">
            <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
                <div class="md:flex md:items-center md:justify-between">
                    <div class="flex-1 min-w-0">
                        <h2 class="text-2xl font-bold leading-7 text-gray-900 sm:text-3xl sm:truncate">
                            Welcome back, {{.User.FirstName}}
                        </h2>
                    </div>
                </div>

                <!-- Customer Stats -->
                <div class="mt-8 grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-3">
                    <div class="bg-white overflow-hidden shadow rounded-lg">
                        <div class="p-5">
                            <div class="flex items-center">
                                <div class="flex-shrink-0">
                                    <div class="w-8 h-8 bg-blue-500 rounded-md flex items-center justify-center">
                                        <svg class="w-5 h-5 text-white" fill="currentColor" viewBox="0 0 20 20">
                                            <path fill-rule="evenodd" d="M6 2a1 1 0 00-1 1v1H4a2 2 0 00-2 2v10a2 2 0 002 2h12a2 2 0 002-2V6a2 2 0 00-2-2h-1V3a1 1 0 10-2 0v1H7V3a1 1 0 00-1-1zm0 5a1 1 0 000 2h8a1 1 0 100-2H6z" clip-rule="evenodd"/>
                                        </svg>
                                    </div>
                                </div>
                                <div class="ml-5 w-0 flex-1">
                                    <dl>
                                        <dt class="text-sm font-medium text-gray-500 truncate">Upcoming Services</dt>
                                        <dd class="text-lg font-medium text-gray-900">3</dd>
                                    </dl>
                                </div>
                            </div>
                        </div>
                    </div>

                    <div class="bg-white overflow-hidden shadow rounded-lg">
                        <div class="p-5">
                            <div class="flex items-center">
                                <div class="flex-shrink-0">
                                    <div class="w-8 h-8 bg-green-500 rounded-md flex items-center justify-center">
                                        <svg class="w-5 h-5 text-white" fill="currentColor" viewBox="0 0 20 20">
                                            <path fill-rule="evenodd" d="M4 4a2 2 0 00-2 2v4a2 2 0 002 2V6h10a2 2 0 00-2-2H4zm2 6a2 2 0 012-2h8a2 2 0 012 2v4a2 2 0 01-2 2H8a2 2 0 01-2-2v-4zm6 4a2 2 0 100-4 2 2 0 000 4z" clip-rule="evenodd"/>
                                        </svg>
                                    </div>
                                </div>
                                <div class="ml-5 w-0 flex-1">
                                    <dl>
                                        <dt class="text-sm font-medium text-gray-500 truncate">Pending Invoices</dt>
                                        <dd class="text-lg font-medium text-gray-900">1</dd>
                                    </dl>
                                </div>
                            </div>
                        </div>
                    </div>

                    <div class="bg-white overflow-hidden shadow rounded-lg">
                        <div class="p-5">
                            <div class="flex items-center">
                                <div class="flex-shrink-0">
                                    <div class="w-8 h-8 bg-purple-500 rounded-md flex items-center justify-center">
                                        <svg class="w-5 h-5 text-white" fill="currentColor" viewBox="0 0 20 20">
                                            <path d="M3 4a1 1 0 011-1h12a1 1 0 011 1v2a1 1 0 01-1 1H4a1 1 0 01-1-1V4zM3 10a1 1 0 011-1h6a1 1 0 011 1v6a1 1 0 01-1 1H4a1 1 0 01-1-1v-6zM14 9a1 1 0 00-1 1v6a1 1 0 001 1h2a1 1 0 001-1v-6a1 1 0 00-1-1h-2z"/>
                                        </svg>
                                    </div>
                                </div>
                                <div class="ml-5 w-0 flex-1">
                                    <dl>
                                        <dt class="text-sm font-medium text-gray-500 truncate">Properties</dt>
                                        <dd class="text-lg font-medium text-gray-900">2</dd>
                                    </dl>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>

                <!-- Quick Actions -->
                <div class="mt-8 bg-white shadow rounded-lg">
                    <div class="p-6">
                        <h3 class="text-lg leading-6 font-medium text-gray-900 mb-4">Quick Actions</h3>
                        <div class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
                            <button class="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md text-blue-700 bg-blue-100 hover:bg-blue-200 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500">
                                Request Service
                            </button>
                            <a href="/portal/billing" class="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md text-green-700 bg-green-100 hover:bg-green-200 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-green-500">
                                View Billing
                            </a>
                            <a href="/portal/services" class="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md text-purple-700 bg-purple-100 hover:bg-purple-200 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-purple-500">
                                Service History
                            </a>
                            <button class="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md text-orange-700 bg-orange-100 hover:bg-orange-200 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-orange-500">
                                Contact Support
                            </button>
                        </div>
                    </div>
                </div>
            </div>
        </main>
    </div>
</div>
{{end}}`,

		"components/customer_sidebar.html": `{{define "customer_sidebar"}}
<div class="hidden lg:flex lg:w-64 lg:flex-col lg:fixed lg:inset-y-0">
    <div class="flex-1 flex flex-col min-h-0 bg-blue-800">
        <div class="flex-1 flex flex-col pt-5 pb-4 overflow-y-auto">
            <div class="flex items-center flex-shrink-0 px-4">
                <h1 class="text-white text-xl font-bold">Customer Portal</h1>
            </div>
            <nav class="mt-5 flex-1 px-2 space-y-1">
                <a href="/portal" class="text-blue-100 hover:bg-blue-700 hover:text-white group flex items-center px-2 py-2 text-sm font-medium rounded-md">
                    Dashboard
                </a>
                <a href="/portal/services" class="text-blue-100 hover:bg-blue-700 hover:text-white group flex items-center px-2 py-2 text-sm font-medium rounded-md">
                    My Services
                </a>
                <a href="/portal/billing" class="text-blue-100 hover:bg-blue-700 hover:text-white group flex items-center px-2 py-2 text-sm font-medium rounded-md">
                    Billing
                </a>
                <a href="/portal/quotes" class="text-blue-100 hover:bg-blue-700 hover:text-white group flex items-center px-2 py-2 text-sm font-medium rounded-md">
                    My Quotes
                </a>
            </nav>
        </div>
    </div>
</div>
{{end}}`,

		"customers_list.html": `{{define "content"}}
<div class="min-h-screen bg-gray-50">
    {{template "sidebar" .}}
    <div class="lg:pl-64">
        {{template "header" .}}
        <main class="py-10">
            <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
                <div class="md:flex md:items-center md:justify-between">
                    <div class="flex-1 min-w-0">
                        <h2 class="text-2xl font-bold leading-7 text-gray-900 sm:text-3xl sm:truncate">
                            Customers
                        </h2>
                    </div>
                    <div class="mt-4 flex md:mt-0 md:ml-4">
                        <a href="/admin/customers/new" 
                           class="ml-3 inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500">
                            Add Customer
                        </a>
                    </div>
                </div>

                <!-- Search and Filters -->
                <div class="mt-8 bg-white shadow rounded-lg">
                    <div class="p-6">
                        <div class="grid grid-cols-1 gap-4 sm:grid-cols-3">
                            <div>
                                <input type="text" 
                                       id="customer-search"
                                       placeholder="Search customers..."
                                       hx-get="/api/v1/customers/table"
                                       hx-trigger="input changed delay:300ms, search"
                                       hx-target="#customers-table-container"
                                       hx-include="[data-search-filter]"
                                       class="block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm placeholder-gray-400 focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm">
                            </div>
                            <div>
                                <select data-search-filter
                                        hx-get="/api/v1/customers/table"
                                        hx-trigger="change"
                                        hx-target="#customers-table-container"
                                        hx-include="[data-search-filter], #customer-search"
                                        class="block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm">
                                    <option value="">All Status</option>
                                    <option value="active">Active</option>
                                    <option value="inactive">Inactive</option>
                                </select>
                            </div>
                        </div>
                    </div>
                </div>

                <!-- Customers Table -->
                <div class="mt-8">
                    <div id="customers-table-container" 
                         hx-get="/api/v1/customers/table" 
                         hx-trigger="load"
                         hx-swap="innerHTML">
                        Loading customers...
                    </div>
                </div>
            </div>
        </main>
    </div>
</div>
{{end}}`,

		"customers_table.html": `{{define "content"}}
<div class="bg-white shadow overflow-hidden sm:rounded-md">
    <ul class="divide-y divide-gray-200">
        {{range .Data.customers}}
        <li>
            <a href="/admin/customers/{{.ID}}" class="block hover:bg-gray-50">
                <div class="px-4 py-4 sm:px-6">
                    <div class="flex items-center justify-between">
                        <div class="flex items-center">
                            <div class="flex-shrink-0 h-10 w-10">
                                <div class="h-10 w-10 rounded-full bg-gray-300 flex items-center justify-center">
                                    <span class="text-sm font-medium text-gray-700">{{substr .FirstName 0 1}}{{substr .LastName 0 1}}</span>
                                </div>
                            </div>
                            <div class="ml-4">
                                <div class="text-sm font-medium text-gray-900">{{.FirstName}} {{.LastName}}</div>
                                <div class="text-sm text-gray-500">{{.Email}}</div>
                            </div>
                        </div>
                        <div class="flex items-center">
                            <div class="text-sm text-gray-500">{{.Phone}}</div>
                            <div class="ml-4">
                                {{if eq .Status "active"}}
                                <span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-800">
                                    Active
                                </span>
                                {{else}}
                                <span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-red-100 text-red-800">
                                    Inactive
                                </span>
                                {{end}}
                            </div>
                        </div>
                    </div>
                </div>
            </a>
        </li>
        {{end}}
    </ul>
    
    <!-- Pagination -->
    {{if gt .Data.totalCount .Data.limit}}
    <div class="bg-white px-4 py-3 flex items-center justify-between border-t border-gray-200 sm:px-6">
        <div class="flex-1 flex justify-between sm:hidden">
            {{if gt .Data.currentPage 1}}
            <button hx-get="/api/v1/customers/table?page={{add .Data.currentPage -1}}"
                    hx-target="#customers-table-container"
                    class="relative inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50">
                Previous
            </button>
            {{end}}
            <button hx-get="/api/v1/customers/table?page={{add .Data.currentPage 1}}"
                    hx-target="#customers-table-container"
                    class="ml-3 relative inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50">
                Next
            </button>
        </div>
    </div>
    {{end}}
</div>
{{end}}`,

		"customer_form.html": `{{define "content"}}
<div class="min-h-screen bg-gray-50">
    {{template "sidebar" .}}
    <div class="lg:pl-64">
        {{template "header" .}}
        <main class="py-10">
            <div class="max-w-3xl mx-auto px-4 sm:px-6 lg:px-8">
                <div class="md:flex md:items-center md:justify-between">
                    <div class="flex-1 min-w-0">
                        <h2 class="text-2xl font-bold leading-7 text-gray-900 sm:text-3xl sm:truncate">
                            {{.Title}}
                        </h2>
                    </div>
                </div>

                <div class="mt-8">
                    <div class="bg-white shadow rounded-lg">
                        <form hx-post="/admin/customers" hx-target="#customer-form" hx-swap="outerHTML">
                            <div id="customer-form" class="px-6 py-8">
                                {{if .Flash.error}}
                                <div class="mb-6 bg-red-50 border border-red-200 rounded-md p-4">
                                    <div class="text-sm text-red-700">{{.Flash.error}}</div>
                                </div>
                                {{end}}

                                <div class="grid grid-cols-1 gap-6 sm:grid-cols-2">
                                    <div>
                                        <label for="first_name" class="block text-sm font-medium text-gray-700">First Name *</label>
                                        <input type="text" name="first_name" id="first_name" required
                                               value="{{if .Data}}{{.Data.FirstName}}{{end}}"
                                               class="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm placeholder-gray-400 focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm">
                                    </div>

                                    <div>
                                        <label for="last_name" class="block text-sm font-medium text-gray-700">Last Name *</label>
                                        <input type="text" name="last_name" id="last_name" required
                                               value="{{if .Data}}{{.Data.LastName}}{{end}}"
                                               class="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm placeholder-gray-400 focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm">
                                    </div>

                                    <div>
                                        <label for="email" class="block text-sm font-medium text-gray-700">Email</label>
                                        <input type="email" name="email" id="email"
                                               value="{{if .Data}}{{if .Data.Email}}{{.Data.Email}}{{end}}{{end}}"
                                               class="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm placeholder-gray-400 focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm">
                                    </div>

                                    <div>
                                        <label for="phone" class="block text-sm font-medium text-gray-700">Phone</label>
                                        <input type="tel" name="phone" id="phone"
                                               value="{{if .Data}}{{if .Data.Phone}}{{.Data.Phone}}{{end}}{{end}}"
                                               class="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm placeholder-gray-400 focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm">
                                    </div>

                                    <div class="sm:col-span-2">
                                        <label for="address_line1" class="block text-sm font-medium text-gray-700">Address Line 1</label>
                                        <input type="text" name="address_line1" id="address_line1"
                                               value="{{if .Data}}{{if .Data.AddressLine1}}{{.Data.AddressLine1}}{{end}}{{end}}"
                                               class="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm placeholder-gray-400 focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm">
                                    </div>

                                    <div class="sm:col-span-2">
                                        <label for="address_line2" class="block text-sm font-medium text-gray-700">Address Line 2</label>
                                        <input type="text" name="address_line2" id="address_line2"
                                               value="{{if .Data}}{{if .Data.AddressLine2}}{{.Data.AddressLine2}}{{end}}{{end}}"
                                               class="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm placeholder-gray-400 focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm">
                                    </div>

                                    <div>
                                        <label for="city" class="block text-sm font-medium text-gray-700">City</label>
                                        <input type="text" name="city" id="city"
                                               value="{{if .Data}}{{if .Data.City}}{{.Data.City}}{{end}}{{end}}"
                                               class="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm placeholder-gray-400 focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm">
                                    </div>

                                    <div>
                                        <label for="state" class="block text-sm font-medium text-gray-700">State</label>
                                        <input type="text" name="state" id="state"
                                               value="{{if .Data}}{{if .Data.State}}{{.Data.State}}{{end}}{{end}}"
                                               class="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm placeholder-gray-400 focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm">
                                    </div>

                                    <div>
                                        <label for="zip_code" class="block text-sm font-medium text-gray-700">ZIP Code</label>
                                        <input type="text" name="zip_code" id="zip_code"
                                               value="{{if .Data}}{{if .Data.ZipCode}}{{.Data.ZipCode}}{{end}}{{end}}"
                                               class="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm placeholder-gray-400 focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm">
                                    </div>
                                </div>

                                <div class="mt-8 flex justify-end space-x-3">
                                    <a href="/admin/customers" 
                                       class="inline-flex items-center px-4 py-2 border border-gray-300 rounded-md shadow-sm text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500">
                                        Cancel
                                    </a>
                                    <button type="submit" 
                                            class="inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500">
                                        Save Customer
                                    </button>
                                </div>
                            </div>
                        </form>
                    </div>
                </div>
            </div>
        </main>
    </div>
</div>
{{end}}`,

		"placeholder.html": `{{define "content"}}
<div class="min-h-screen bg-gray-50">
    {{template "sidebar" .}}
    <div class="lg:pl-64">
        {{template "header" .}}
        <main class="py-10">
            <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
                <div class="text-center">
                    <h2 class="text-2xl font-bold text-gray-900 mb-4">{{.Title}}</h2>
                    <p class="text-gray-600">{{.Data.message}}</p>
                    <div class="mt-8">
                        <div class="inline-flex items-center px-4 py-2 bg-blue-100 text-blue-800 rounded-md">
                            <svg class="w-5 h-5 mr-2" fill="currentColor" viewBox="0 0 20 20">
                                <path fill-rule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z" clip-rule="evenodd"/>
                            </svg>
                            Coming Soon
                        </div>
                    </div>
                </div>
            </div>
        </main>
    </div>
</div>
{{end}}`,

		"properties_list.html": `{{define "content"}}
<div class="min-h-screen bg-gray-50">
    {{template "sidebar" .}}
    <div class="lg:pl-64">
        {{template "header" .}}
        <main class="py-10">
            <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
                <div class="md:flex md:items-center md:justify-between">
                    <div class="flex-1 min-w-0">
                        <h2 class="text-2xl font-bold leading-7 text-gray-900 sm:text-3xl sm:truncate">
                            Property Management
                        </h2>
                        <p class="mt-1 text-sm text-gray-500">Manage properties, locations, and geographic data</p>
                    </div>
                    <div class="mt-4 flex md:mt-0 md:ml-4 space-x-3">
                        <button id="toggle-map-view" 
                                class="inline-flex items-center px-4 py-2 border border-gray-300 rounded-md shadow-sm text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500">
                            <svg class="mr-2 h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 20l-5.447-2.724A1 1 0 013 16.382V5.618a1 1 0 011.447-.894L9 7m0 13l6-3m-6 3V7m6 10l4.553 2.276A1 1 0 0021 18.382V7.618a1 1 0 00-1.447-.894L15 4m0 13V4m-6 3l6-3"/>
                            </svg>
                            Toggle Map
                        </button>
                        <a href="/admin/properties/new" 
                           class="inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500">
                            <svg class="mr-2 h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"/>
                            </svg>
                            Add Property
                        </a>
                    </div>
                </div>

                <!-- View Toggle & Search -->
                <div class="mt-8 bg-white shadow rounded-lg">
                    <div class="p-6">
                        <div class="grid grid-cols-1 gap-4 sm:grid-cols-4">
                            <div class="sm:col-span-2">
                                <input type="text" 
                                       id="property-search"
                                       placeholder="Search properties by address, customer, or description..."
                                       hx-get="/api/v1/properties/table"
                                       hx-trigger="input changed delay:300ms, search"
                                       hx-target="#properties-content"
                                       hx-include="[data-search-filter]"
                                       class="block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm placeholder-gray-400 focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm">
                            </div>
                            <div>
                                <select data-search-filter name="customer_id"
                                        hx-get="/api/v1/properties/table"
                                        hx-trigger="change"
                                        hx-target="#properties-content"
                                        hx-include="[data-search-filter], #property-search"
                                        class="block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm">
                                    <option value="">All Customers</option>
                                    <!-- Customer options loaded dynamically -->
                                </select>
                            </div>
                            <div>
                                <select data-search-filter name="service_type"
                                        hx-get="/api/v1/properties/table"
                                        hx-trigger="change"
                                        hx-target="#properties-content"
                                        hx-include="[data-search-filter], #property-search"
                                        class="block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm">
                                    <option value="">All Service Types</option>
                                    <option value="lawn_care">Lawn Care</option>
                                    <option value="landscaping">Landscaping</option>
                                    <option value="snow_removal">Snow Removal</option>
                                    <option value="irrigation">Irrigation</option>
                                </select>
                            </div>
                        </div>
                    </div>
                </div>

                <!-- Split View: Map & Properties -->
                <div class="mt-8 grid grid-cols-1 lg:grid-cols-2 gap-8" id="properties-main-content">
                    <!-- Map View -->
                    <div class="bg-white shadow rounded-lg" id="map-container">
                        <div class="p-6">
                            <h3 class="text-lg leading-6 font-medium text-gray-900 mb-4">Property Locations</h3>
                            <div id="properties-map" class="h-96 rounded-lg bg-gray-100"></div>
                            <div class="mt-4 flex items-center justify-between text-sm text-gray-500">
                                <span id="map-status">Loading map...</span>
                                <button onclick="LandscapingApp.centerMapOnAll()" 
                                        class="text-blue-600 hover:text-blue-500">
                                    Center All Properties
                                </button>
                            </div>
                        </div>
                    </div>

                    <!-- Properties List/Table -->
                    <div class="bg-white shadow rounded-lg">
                        <div class="p-6">
                            <h3 class="text-lg leading-6 font-medium text-gray-900 mb-4">Properties</h3>
                            <div id="properties-content" 
                                 hx-get="/api/v1/properties/table" 
                                 hx-trigger="load"
                                 hx-swap="innerHTML">
                                <div class="flex items-center justify-center h-32">
                                    <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>

                <!-- Statistics Cards -->
                <div class="mt-8 grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-4">
                    <div class="bg-white overflow-hidden shadow rounded-lg">
                        <div class="p-5">
                            <div class="flex items-center">
                                <div class="flex-shrink-0">
                                    <div class="w-8 h-8 bg-blue-500 rounded-md flex items-center justify-center">
                                        <svg class="w-5 h-5 text-white" fill="currentColor" viewBox="0 0 20 20">
                                            <path fill-rule="evenodd" d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4"/>
                                        </svg>
                                    </div>
                                </div>
                                <div class="ml-5 w-0 flex-1">
                                    <dl>
                                        <dt class="text-sm font-medium text-gray-500 truncate">Total Properties</dt>
                                        <dd class="text-lg font-medium text-gray-900" id="total-properties">--</dd>
                                    </dl>
                                </div>
                            </div>
                        </div>
                    </div>

                    <div class="bg-white overflow-hidden shadow rounded-lg">
                        <div class="p-5">
                            <div class="flex items-center">
                                <div class="flex-shrink-0">
                                    <div class="w-8 h-8 bg-green-500 rounded-md flex items-center justify-center">
                                        <svg class="w-5 h-5 text-white" fill="currentColor" viewBox="0 0 20 20">
                                            <path fill-rule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clip-rule="evenodd"/>
                                        </svg>
                                    </div>
                                </div>
                                <div class="ml-5 w-0 flex-1">
                                    <dl>
                                        <dt class="text-sm font-medium text-gray-500 truncate">Active Services</dt>
                                        <dd class="text-lg font-medium text-gray-900" id="active-services">--</dd>
                                    </dl>
                                </div>
                            </div>
                        </div>
                    </div>

                    <div class="bg-white overflow-hidden shadow rounded-lg">
                        <div class="p-5">
                            <div class="flex items-center">
                                <div class="flex-shrink-0">
                                    <div class="w-8 h-8 bg-yellow-500 rounded-md flex items-center justify-center">
                                        <svg class="w-5 h-5 text-white" fill="currentColor" viewBox="0 0 20 20">
                                            <path fill-rule="evenodd" d="M3 6a3 3 0 013-3h10a1 1 0 01.8 1.6L14.25 8l2.55 3.4A1 1 0 0116 13H6a1 1 0 00-1 1v3a1 1 0 11-2 0V6z" clip-rule="evenodd"/>
                                        </svg>
                                    </div>
                                </div>
                                <div class="ml-5 w-0 flex-1">
                                    <dl>
                                        <dt class="text-sm font-medium text-gray-500 truncate">Avg Property Size</dt>
                                        <dd class="text-lg font-medium text-gray-900" id="avg-property-size">--</dd>
                                    </dl>
                                </div>
                            </div>
                        </div>
                    </div>

                    <div class="bg-white overflow-hidden shadow rounded-lg">
                        <div class="p-5">
                            <div class="flex items-center">
                                <div class="flex-shrink-0">
                                    <div class="w-8 h-8 bg-purple-500 rounded-md flex items-center justify-center">
                                        <svg class="w-5 h-5 text-white" fill="currentColor" viewBox="0 0 20 20">
                                            <path d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"/>
                                        </svg>
                                    </div>
                                </div>
                                <div class="ml-5 w-0 flex-1">
                                    <dl>
                                        <dt class="text-sm font-medium text-gray-500 truncate">Service Areas</dt>
                                        <dd class="text-lg font-medium text-gray-900" id="service-areas">--</dd>
                                    </dl>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </main>
    </div>
</div>

<script>
document.addEventListener('DOMContentLoaded', function() {
    // Initialize property map
    const map = LandscapingApp.initializeMap('properties-map', {
        center: [39.8283, -98.5795], // Center of US
        zoom: 4
    });
    
    if (map) {
        // Store map instance globally for property management
        window.propertyMap = map;
        document.getElementById('map-status').textContent = 'Map loaded successfully';
        
        // Load properties on map
        loadPropertiesOnMap();
    }
    
    // Toggle map view
    document.getElementById('toggle-map-view').addEventListener('click', function() {
        const mapContainer = document.getElementById('map-container');
        const propertiesContent = document.getElementById('properties-main-content');
        
        if (mapContainer.style.display === 'none') {
            mapContainer.style.display = 'block';
            propertiesContent.classList.remove('lg:grid-cols-1');
            propertiesContent.classList.add('lg:grid-cols-2');
            this.querySelector('svg').style.transform = 'rotate(0deg)';
        } else {
            mapContainer.style.display = 'none';
            propertiesContent.classList.remove('lg:grid-cols-2');
            propertiesContent.classList.add('lg:grid-cols-1');
            this.querySelector('svg').style.transform = 'rotate(180deg)';
        }
    });
});

function loadPropertiesOnMap() {
    // This would typically fetch from API
    fetch('/api/v1/properties?format=geojson')
        .then(response => response.json())
        .then(data => {
            if (window.propertyMap && data.properties) {
                data.properties.forEach(property => {
                    if (property.geometry && property.geometry.coordinates) {
                        const [lng, lat] = property.geometry.coordinates;
                        L.marker([lat, lng])
                            .addTo(window.propertyMap)
                            .bindPopup(\`
                                <div class="p-2">
                                    <h4 class="font-semibold">\${property.properties.name}</h4>
                                    <p class="text-sm text-gray-600">\${property.properties.address}</p>
                                    <p class="text-sm text-gray-600">Customer: \${property.properties.customer_name}</p>
                                    <a href="/admin/properties/\${property.properties.id}" class="text-blue-600 text-sm">View Details</a>
                                </div>
                            \`);
                    }
                });
            }
        })
        .catch(error => {
            console.error('Error loading properties on map:', error);
            document.getElementById('map-status').textContent = 'Error loading properties';
        });
}

// Add to global LandscapingApp
window.LandscapingApp.centerMapOnAll = function() {
    if (window.propertyMap) {
        window.propertyMap.fitBounds(window.propertyMap.getBounds());
    }
};
</script>
{{end}}`,

		"properties_table.html": `{{define "content"}}
<div class="overflow-hidden">
    <div class="space-y-4">
        {{range .Data.properties}}
        <div class="border border-gray-200 rounded-lg p-4 hover:shadow-md transition-shadow duration-150 ease-in-out">
            <div class="flex items-start justify-between">
                <div class="flex-1">
                    <div class="flex items-center space-x-3 mb-2">
                        <h4 class="text-lg font-medium text-gray-900">
                            <a href="/admin/properties/{{.ID}}" class="hover:text-blue-600">{{.Name}}</a>
                        </h4>
                        <span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium {{if eq .ServiceType "lawn_care"}}bg-green-100 text-green-800{{else if eq .ServiceType "landscaping"}}bg-blue-100 text-blue-800{{else if eq .ServiceType "snow_removal"}}bg-cyan-100 text-cyan-800{{else}}bg-gray-100 text-gray-800{{end}}">
                            {{.ServiceType}}
                        </span>
                    </div>
                    
                    <div class="text-sm text-gray-600 mb-2">
                        <div class="flex items-center space-x-1 mb-1">
                            <svg class="w-4 h-4 text-gray-400" fill="currentColor" viewBox="0 0 20 20">
                                <path fill-rule="evenodd" d="M5.05 4.05a7 7 0 119.9 9.9L10 18.9l-4.95-4.95a7 7 0 010-9.9zM10 11a2 2 0 100-4 2 2 0 000 4z" clip-rule="evenodd"/>
                            </svg>
                            <span>{{.Address}}</span>
                        </div>
                        <div class="flex items-center space-x-1 mb-1">
                            <svg class="w-4 h-4 text-gray-400" fill="currentColor" viewBox="0 0 20 20">
                                <path fill-rule="evenodd" d="M10 9a3 3 0 100-6 3 3 0 000 6zm-7 9a7 7 0 1114 0H3z" clip-rule="evenodd"/>
                            </svg>
                            <span>Customer: {{.CustomerName}}</span>
                        </div>
                        {{if .PropertySize}}
                        <div class="flex items-center space-x-1">
                            <svg class="w-4 h-4 text-gray-400" fill="currentColor" viewBox="0 0 20 20">
                                <path fill-rule="evenodd" d="M3 6a3 3 0 013-3h10a1 1 0 01.8 1.6L14.25 8l2.55 3.4A1 1 0 0116 13H6a1 1 0 00-1 1v3a1 1 0 11-2 0V6z" clip-rule="evenodd"/>
                            </svg>
                            <span>{{.PropertySize}} sq ft</span>
                        </div>
                        {{end}}
                    </div>
                    
                    {{if .Description}}
                    <p class="text-sm text-gray-600 mb-2">{{.Description}}</p>
                    {{end}}
                    
                    <div class="flex items-center space-x-4 text-xs text-gray-500">
                        <span>Last Service: {{formatDate .LastServiceDate}}</span>
                        <span>Next Service: {{formatDate .NextServiceDate}}</span>
                        {{if .JobsCount}}
                        <span>Jobs: {{.JobsCount}}</span>
                        {{end}}
                    </div>
                </div>
                
                <div class="flex-shrink-0 ml-4">
                    <div class="flex items-center space-x-2">
                        <button onclick="showPropertyOnMap({{.Latitude}}, {{.Longitude}}, '{{.Name}}')"
                                class="p-2 text-gray-400 hover:text-blue-600 rounded-full hover:bg-gray-100">
                            <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
                                <path fill-rule="evenodd" d="M5.05 4.05a7 7 0 119.9 9.9L10 18.9l-4.95-4.95a7 7 0 010-9.9zM10 11a2 2 0 100-4 2 2 0 000 4z" clip-rule="evenodd"/>
                            </svg>
                        </button>
                        <a href="/admin/properties/{{.ID}}" 
                           class="p-2 text-gray-400 hover:text-blue-600 rounded-full hover:bg-gray-100">
                            <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
                                <path d="M10 12a2 2 0 100-4 2 2 0 000 4z"/>
                                <path fill-rule="evenodd" d="M.458 10C1.732 5.943 5.522 3 10 3s8.268 2.943 9.542 7c-1.274 4.057-5.064 7-9.542 7S1.732 14.057.458 10zM14 10a4 4 0 11-8 0 4 4 0 018 0z" clip-rule="evenodd"/>
                            </svg>
                        </a>
                        <button hx-post="/admin/properties/{{.ID}}/toggle-favorite"
                                hx-target="closest .border"
                                hx-swap="outerHTML"
                                class="p-2 text-gray-400 hover:text-yellow-600 rounded-full hover:bg-gray-100">
                            <svg class="w-4 h-4" fill="{{if .IsFavorite}}currentColor{{else}}none{{end}}" stroke="currentColor" viewBox="0 0 24 24">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11.049 2.927c.3-.921 1.603-.921 1.902 0l1.519 4.674a1 1 0 00.95.69h4.915c.969 0 1.371 1.24.588 1.81l-3.976 2.888a1 1 0 00-.363 1.118l1.518 4.674c.3.922-.755 1.688-1.538 1.118l-3.976-2.888a1 1 0 00-1.176 0l-3.976 2.888c-.783.57-1.838-.197-1.538-1.118l1.518-4.674a1 1 0 00-.363-1.118l-3.976-2.888c-.784-.57-.38-1.81.588-1.81h4.914a1 1 0 00.951-.69l1.519-4.674z"/>
                            </svg>
                        </button>
                    </div>
                </div>
            </div>
        </div>
        {{end}}
        
        {{if eq (len .Data.properties) 0}}
        <div class="text-center py-12">
            <svg class="mx-auto h-12 w-12 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4"/>
            </svg>
            <h3 class="mt-2 text-sm font-medium text-gray-900">No properties found</h3>
            <p class="mt-1 text-sm text-gray-500">Get started by creating a new property.</p>
            <div class="mt-6">
                <a href="/admin/properties/new" 
                   class="inline-flex items-center px-4 py-2 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500">
                    <svg class="mr-2 h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"/>
                    </svg>
                    Add Property
                </a>
            </div>
        </div>
        {{end}}
    </div>
    
    <!-- Pagination -->
    {{if gt .Data.totalCount .Data.limit}}
    <div class="mt-6 flex items-center justify-between border-t border-gray-200 pt-6">
        <div class="flex-1 flex justify-between sm:hidden">
            {{if gt .Data.currentPage 1}}
            <button hx-get="/api/v1/properties/table?page={{add .Data.currentPage -1}}"
                    hx-target="#properties-content"
                    hx-include="[data-search-filter], #property-search"
                    class="relative inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50">
                Previous
            </button>
            {{end}}
            {{if lt .Data.currentPage .Data.totalPages}}
            <button hx-get="/api/v1/properties/table?page={{add .Data.currentPage 1}}"
                    hx-target="#properties-content"
                    hx-include="[data-search-filter], #property-search"
                    class="ml-3 relative inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50">
                Next
            </button>
            {{end}}
        </div>
        
        <div class="hidden sm:flex-1 sm:flex sm:items-center sm:justify-between">
            <div>
                <p class="text-sm text-gray-700">
                    Showing <span class="font-medium">{{.Data.offset}}</span> to 
                    <span class="font-medium">{{add .Data.offset (len .Data.properties)}}</span> of 
                    <span class="font-medium">{{.Data.totalCount}}</span> properties
                </p>
            </div>
            <div>
                <nav class="relative z-0 inline-flex rounded-md shadow-sm -space-x-px">
                    {{if gt .Data.currentPage 1}}
                    <button hx-get="/api/v1/properties/table?page={{add .Data.currentPage -1}}"
                            hx-target="#properties-content"
                            hx-include="[data-search-filter], #property-search"
                            class="relative inline-flex items-center px-2 py-2 rounded-l-md border border-gray-300 bg-white text-sm font-medium text-gray-500 hover:bg-gray-50">
                        Previous
                    </button>
                    {{end}}
                    
                    {{range $page := .Data.pageNumbers}}
                    <button hx-get="/api/v1/properties/table?page={{$page}}"
                            hx-target="#properties-content"
                            hx-include="[data-search-filter], #property-search"
                            class="relative inline-flex items-center px-4 py-2 border text-sm font-medium {{if eq $page $.Data.currentPage}}bg-blue-50 border-blue-500 text-blue-600{{else}}bg-white border-gray-300 text-gray-500 hover:bg-gray-50{{end}}">
                        {{$page}}
                    </button>
                    {{end}}
                    
                    {{if lt .Data.currentPage .Data.totalPages}}
                    <button hx-get="/api/v1/properties/table?page={{add .Data.currentPage 1}}"
                            hx-target="#properties-content"
                            hx-include="[data-search-filter], #property-search"
                            class="relative inline-flex items-center px-2 py-2 rounded-r-md border border-gray-300 bg-white text-sm font-medium text-gray-500 hover:bg-gray-50">
                        Next
                    </button>
                    {{end}}
                </nav>
            </div>
        </div>
    </div>
    {{end}}
</div>

<script>
function showPropertyOnMap(lat, lng, name) {
    if (window.propertyMap) {
        window.propertyMap.setView([lat, lng], 16);
        L.popup()
            .setLatLng([lat, lng])
            .setContent(\`<div class="p-2"><h4 class="font-semibold">\${name}</h4></div>\`)
            .openOn(window.propertyMap);
        
        // Scroll to map
        document.getElementById('map-container').scrollIntoView({ behavior: 'smooth' });
    }
}
</script>
{{end}}`,

		"jobs_list.html": `{{define "content"}}
<div class="min-h-screen bg-gray-50">
    {{template "sidebar" .}}
    <div class="lg:pl-64">
        {{template "header" .}}
        <main class="py-10">
            <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
                <div class="md:flex md:items-center md:justify-between">
                    <div class="flex-1 min-w-0">
                        <h2 class="text-2xl font-bold leading-7 text-gray-900 sm:text-3xl sm:truncate">
                            Job Management
                        </h2>
                        <p class="mt-1 text-sm text-gray-500">Schedule, track, and manage landscaping jobs</p>
                    </div>
                    <div class="mt-4 flex md:mt-0 md:ml-4 space-x-3">
                        <a href="/admin/jobs/calendar" 
                           class="inline-flex items-center px-4 py-2 border border-gray-300 rounded-md shadow-sm text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500">
                            <svg class="mr-2 h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3a2 2 0 012-2h4a2 2 0 012 2v4m-6 9l2 2 4-4m6-7H2a1 1 0 00-1 1v10a1 1 0 001 1h20a1 1 0 001-1V8a1 1 0 00-1-1z"/>
                            </svg>
                            Calendar View
                        </a>
                        <a href="/admin/jobs/new" 
                           class="inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500">
                            <svg class="mr-2 h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"/>
                            </svg>
                            Schedule Job
                        </a>
                    </div>
                </div>

                <!-- Search and Filters -->
                <div class="mt-8 bg-white shadow rounded-lg">
                    <div class="p-6">
                        <div class="grid grid-cols-1 gap-4 sm:grid-cols-5">
                            <div class="sm:col-span-2">
                                <input type="text" 
                                       id="job-search"
                                       placeholder="Search jobs by customer, property, or description..."
                                       hx-get="/api/v1/jobs/table"
                                       hx-trigger="input changed delay:300ms, search"
                                       hx-target="#jobs-content"
                                       hx-include="[data-search-filter]"
                                       class="block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm placeholder-gray-400 focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm">
                            </div>
                            <div>
                                <select data-search-filter name="status"
                                        hx-get="/api/v1/jobs/table"
                                        hx-trigger="change"
                                        hx-target="#jobs-content"
                                        hx-include="[data-search-filter], #job-search"
                                        class="block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm">
                                    <option value="">All Status</option>
                                    <option value="scheduled">Scheduled</option>
                                    <option value="in_progress">In Progress</option>
                                    <option value="completed">Completed</option>
                                    <option value="cancelled">Cancelled</option>
                                </select>
                            </div>
                            <div>
                                <select data-search-filter name="service_type"
                                        hx-get="/api/v1/jobs/table"
                                        hx-trigger="change"
                                        hx-target="#jobs-content"
                                        hx-include="[data-search-filter], #job-search"
                                        class="block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm">
                                    <option value="">All Services</option>
                                    <option value="lawn_mowing">Lawn Mowing</option>
                                    <option value="hedge_trimming">Hedge Trimming</option>
                                    <option value="fertilizing">Fertilizing</option>
                                    <option value="snow_removal">Snow Removal</option>
                                    <option value="landscaping">Landscaping</option>
                                </select>
                            </div>
                            <div>
                                <select data-search-filter name="crew_id"
                                        hx-get="/api/v1/jobs/table"
                                        hx-trigger="change"
                                        hx-target="#jobs-content"
                                        hx-include="[data-search-filter], #job-search"
                                        class="block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm">
                                    <option value="">All Crews</option>
                                    <option value="1">Team Alpha</option>
                                    <option value="2">Team Beta</option>
                                    <option value="3">Team Gamma</option>
                                </select>
                            </div>
                        </div>
                    </div>
                </div>

                <!-- Job Status Overview -->
                <div class="mt-8 grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-4">
                    <div class="bg-white overflow-hidden shadow rounded-lg">
                        <div class="p-5">
                            <div class="flex items-center">
                                <div class="flex-shrink-0">
                                    <div class="w-8 h-8 bg-yellow-500 rounded-md flex items-center justify-center">
                                        <svg class="w-5 h-5 text-white" fill="currentColor" viewBox="0 0 20 20">
                                            <path fill-rule="evenodd" d="M6 2a1 1 0 00-1 1v1H4a2 2 0 00-2 2v10a2 2 0 002 2h12a2 2 0 002-2V6a2 2 0 00-2-2h-1V3a1 1 0 10-2 0v1H7V3a1 1 0 00-1-1zm0 5a1 1 0 000 2h8a1 1 0 100-2H6z" clip-rule="evenodd"/>
                                        </svg>
                                    </div>
                                </div>
                                <div class="ml-5 w-0 flex-1">
                                    <dl>
                                        <dt class="text-sm font-medium text-gray-500 truncate">Scheduled Today</dt>
                                        <dd class="text-lg font-medium text-gray-900" id="scheduled-today">8</dd>
                                    </dl>
                                </div>
                            </div>
                        </div>
                    </div>

                    <div class="bg-white overflow-hidden shadow rounded-lg">
                        <div class="p-5">
                            <div class="flex items-center">
                                <div class="flex-shrink-0">
                                    <div class="w-8 h-8 bg-blue-500 rounded-md flex items-center justify-center">
                                        <svg class="w-5 h-5 text-white" fill="currentColor" viewBox="0 0 20 20">
                                            <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm1-12a1 1 0 10-2 0v4a1 1 0 00.293.707l2.828 2.829a1 1 0 101.415-1.415L11 9.586V6z" clip-rule="evenodd"/>
                                        </svg>
                                    </div>
                                </div>
                                <div class="ml-5 w-0 flex-1">
                                    <dl>
                                        <dt class="text-sm font-medium text-gray-500 truncate">In Progress</dt>
                                        <dd class="text-lg font-medium text-gray-900" id="in-progress">3</dd>
                                    </dl>
                                </div>
                            </div>
                        </div>
                    </div>

                    <div class="bg-white overflow-hidden shadow rounded-lg">
                        <div class="p-5">
                            <div class="flex items-center">
                                <div class="flex-shrink-0">
                                    <div class="w-8 h-8 bg-green-500 rounded-md flex items-center justify-center">
                                        <svg class="w-5 h-5 text-white" fill="currentColor" viewBox="0 0 20 20">
                                            <path fill-rule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clip-rule="evenodd"/>
                                        </svg>
                                    </div>
                                </div>
                                <div class="ml-5 w-0 flex-1">
                                    <dl>
                                        <dt class="text-sm font-medium text-gray-500 truncate">Completed Today</dt>
                                        <dd class="text-lg font-medium text-gray-900" id="completed-today">12</dd>
                                    </dl>
                                </div>
                            </div>
                        </div>
                    </div>

                    <div class="bg-white overflow-hidden shadow rounded-lg">
                        <div class="p-5">
                            <div class="flex items-center">
                                <div class="flex-shrink-0">
                                    <div class="w-8 h-8 bg-purple-500 rounded-md flex items-center justify-center">
                                        <svg class="w-5 h-5 text-white" fill="currentColor" viewBox="0 0 20 20">
                                            <path d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"/>
                                        </svg>
                                    </div>
                                </div>
                                <div class="ml-5 w-0 flex-1">
                                    <dl>
                                        <dt class="text-sm font-medium text-gray-500 truncate">Total Revenue</dt>
                                        <dd class="text-lg font-medium text-gray-900" id="total-revenue">$2,850</dd>
                                    </dl>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>

                <!-- Jobs List -->
                <div class="mt-8 bg-white shadow rounded-lg">
                    <div class="p-6">
                        <h3 class="text-lg leading-6 font-medium text-gray-900 mb-4">Job Schedule</h3>
                        <div id="jobs-content" 
                             hx-get="/api/v1/jobs/table" 
                             hx-trigger="load"
                             hx-swap="innerHTML">
                            <div class="flex items-center justify-center h-32">
                                <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </main>
    </div>
</div>
{{end}}`,

		"job_calendar.html": `{{define "content"}}
<div class="min-h-screen bg-gray-50">
    {{template "sidebar" .}}
    <div class="lg:pl-64">
        {{template "header" .}}
        <main class="py-10">
            <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
                <div class="md:flex md:items-center md:justify-between">
                    <div class="flex-1 min-w-0">
                        <h2 class="text-2xl font-bold leading-7 text-gray-900 sm:text-3xl sm:truncate">
                            Job Calendar
                        </h2>
                        <p class="mt-1 text-sm text-gray-500">Drag and drop to reschedule jobs, view crew assignments</p>
                    </div>
                    <div class="mt-4 flex md:mt-0 md:ml-4 space-x-3">
                        <a href="/admin/jobs" 
                           class="inline-flex items-center px-4 py-2 border border-gray-300 rounded-md shadow-sm text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500">
                            <svg class="mr-2 h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 10h16M4 14h16M4 18h16"/>
                            </svg>
                            List View
                        </a>
                        <button onclick="showRouteOptimization()"
                                class="inline-flex items-center px-4 py-2 border border-gray-300 rounded-md shadow-sm text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500">
                            <svg class="mr-2 h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 20l-5.447-2.724A1 1 0 013 16.382V5.618a1 1 0 011.447-.894L9 7m0 13l6-3m-6 3V7m6 10l4.553 2.276A1 1 0 0021 18.382V7.618a1 1 0 00-1.447-.894L15 4m0 13V4m-6 3l6-3"/>
                            </svg>
                            Route Optimization
                        </button>
                        <a href="/admin/jobs/new" 
                           class="inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500">
                            <svg class="mr-2 h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"/>
                            </svg>
                            Schedule Job
                        </a>
                    </div>
                </div>

                <!-- Calendar Controls -->
                <div class="mt-8 bg-white shadow rounded-lg">
                    <div class="p-6">
                        <div class="flex items-center justify-between mb-4">
                            <div class="flex items-center space-x-4">
                                <div class="flex items-center space-x-2">
                                    <label class="text-sm font-medium text-gray-700">View:</label>
                                    <select id="calendar-view" class="border border-gray-300 rounded-md text-sm">
                                        <option value="dayGridMonth">Month</option>
                                        <option value="timeGridWeek">Week</option>
                                        <option value="timeGridDay">Day</option>
                                        <option value="listWeek">List</option>
                                    </select>
                                </div>
                                <div class="flex items-center space-x-2">
                                    <label class="text-sm font-medium text-gray-700">Crew:</label>
                                    <select id="crew-filter" class="border border-gray-300 rounded-md text-sm">
                                        <option value="">All Crews</option>
                                        <option value="1">Team Alpha</option>
                                        <option value="2">Team Beta</option>
                                        <option value="3">Team Gamma</option>
                                    </select>
                                </div>
                            </div>
                            <div class="flex items-center space-x-2">
                                <button onclick="calendar.today()" class="px-3 py-1 border border-gray-300 rounded text-sm hover:bg-gray-50">Today</button>
                                <button onclick="calendar.prev()" class="px-3 py-1 border border-gray-300 rounded text-sm hover:bg-gray-50">‹</button>
                                <button onclick="calendar.next()" class="px-3 py-1 border border-gray-300 rounded text-sm hover:bg-gray-50">›</button>
                            </div>
                        </div>

                        <!-- Legend -->
                        <div class="flex items-center space-x-6 mb-4 text-sm">
                            <div class="flex items-center space-x-2">
                                <div class="w-3 h-3 bg-yellow-400 rounded"></div>
                                <span>Scheduled</span>
                            </div>
                            <div class="flex items-center space-x-2">
                                <div class="w-3 h-3 bg-blue-500 rounded"></div>
                                <span>In Progress</span>
                            </div>
                            <div class="flex items-center space-x-2">
                                <div class="w-3 h-3 bg-green-500 rounded"></div>
                                <span>Completed</span>
                            </div>
                            <div class="flex items-center space-x-2">
                                <div class="w-3 h-3 bg-red-500 rounded"></div>
                                <span>Overdue</span>
                            </div>
                            <div class="flex items-center space-x-2">
                                <div class="w-3 h-3 bg-gray-400 rounded"></div>
                                <span>Cancelled</span>
                            </div>
                        </div>

                        <!-- Calendar Container -->
                        <div id="job-calendar" class="h-96"></div>
                    </div>
                </div>

                <!-- Route Optimization Modal -->
                <div id="route-optimization-modal" class="hidden fixed inset-0 z-50 overflow-y-auto">
                    <div class="flex items-center justify-center min-h-screen pt-4 px-4 pb-20 text-center sm:block sm:p-0">
                        <div class="fixed inset-0 transition-opacity" aria-hidden="true">
                            <div class="absolute inset-0 bg-gray-500 opacity-75"></div>
                        </div>
                        <div class="inline-block align-bottom bg-white rounded-lg text-left overflow-hidden shadow-xl transform transition-all sm:my-8 sm:align-middle sm:max-w-4xl sm:w-full">
                            <div class="bg-white px-4 pt-5 pb-4 sm:p-6 sm:pb-4">
                                <div class="flex items-center justify-between mb-4">
                                    <h3 class="text-lg leading-6 font-medium text-gray-900">Route Optimization</h3>
                                    <button onclick="hideRouteOptimization()" class="text-gray-400 hover:text-gray-600">
                                        <svg class="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
                                        </svg>
                                    </button>
                                </div>
                                <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
                                    <div>
                                        <h4 class="text-md font-medium text-gray-900 mb-3">Today's Jobs</h4>
                                        <div id="route-jobs" class="space-y-2 max-h-64 overflow-y-auto">
                                            <!-- Jobs will be populated here -->
                                        </div>
                                        <button onclick="optimizeRoute()" class="mt-4 w-full px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700">
                                            Optimize Route
                                        </button>
                                    </div>
                                    <div>
                                        <h4 class="text-md font-medium text-gray-900 mb-3">Optimized Route</h4>
                                        <div id="route-map" class="h-64 bg-gray-100 rounded"></div>
                                        <div class="mt-2 text-sm text-gray-600">
                                            <span id="route-info">Click "Optimize Route" to calculate best path</span>
                                        </div>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </main>
    </div>
</div>

<script>
let calendar;

document.addEventListener('DOMContentLoaded', function() {
    // Initialize calendar
    calendar = LandscapingApp.initializeCalendar('job-calendar', {
        events: getJobEvents(),
        eventDrop: function(info) {
            updateJobSchedule(info.event.id, info.event.start);
        },
        eventClick: function(info) {
            showJobDetails(info.event.id);
        },
        dateClick: function(info) {
            showCreateJobModal(info.dateStr);
        }
    });
    
    // Calendar view change
    document.getElementById('calendar-view').addEventListener('change', function() {
        calendar.changeView(this.value);
    });
    
    // Crew filter
    document.getElementById('crew-filter').addEventListener('change', function() {
        filterJobsByCrew(this.value);
    });
});

function getJobEvents() {
    // Mock job events for calendar
    return [
        {
            id: '1',
            title: 'Lawn Mowing - Johnson Residence',
            start: '2024-08-13T09:00:00',
            end: '2024-08-13T11:00:00',
            backgroundColor: '#eab308',
            borderColor: '#ca8a04',
            extendedProps: {
                customer: 'John Johnson',
                property: 'Johnson Residence Front Yard',
                crew: 'Team Alpha',
                status: 'scheduled'
            }
        },
        {
            id: '2',
            title: 'Hedge Trimming - Smith Corp',
            start: '2024-08-13T13:00:00',
            end: '2024-08-13T16:00:00',
            backgroundColor: '#3b82f6',
            borderColor: '#2563eb',
            extendedProps: {
                customer: 'Smith Corp',
                property: 'Smith Commercial Property',
                crew: 'Team Beta',
                status: 'in_progress'
            }
        },
        {
            id: '3',
            title: 'Fertilizing - Wilson Garden',
            start: '2024-08-14T08:00:00',
            end: '2024-08-14T10:00:00',
            backgroundColor: '#10b981',
            borderColor: '#059669',
            extendedProps: {
                customer: 'Sarah Wilson',
                property: 'Wilson Backyard Garden',
                crew: 'Team Alpha',
                status: 'completed'
            }
        }
    ];
}

function showJobDetails(jobId) {
    // Show job details modal
    console.log('Show job details for:', jobId);
    window.location.href = '/admin/jobs/' + jobId;
}

function showCreateJobModal(dateStr) {
    // Show create job modal for specific date
    console.log('Create job for date:', dateStr);
    window.location.href = '/admin/jobs/new?date=' + dateStr;
}

function filterJobsByCrew(crewId) {
    // Filter calendar events by crew
    if (calendar) {
        const events = getJobEvents().filter(event => 
            !crewId || event.extendedProps.crew.includes(crewId)
        );
        calendar.removeAllEvents();
        calendar.addEventSource(events);
    }
}

function showRouteOptimization() {
    document.getElementById('route-optimization-modal').classList.remove('hidden');
    
    // Populate today's jobs
    const today = new Date().toISOString().split('T')[0];
    const todayJobs = getJobEvents().filter(event => 
        event.start.startsWith(today)
    );
    
    const jobsContainer = document.getElementById('route-jobs');
    jobsContainer.innerHTML = todayJobs.map(job => \`
        <div class="flex items-center space-x-3 p-2 border border-gray-200 rounded">
            <input type="checkbox" checked class="job-checkbox" data-job-id="\${job.id}">
            <div class="flex-1">
                <div class="text-sm font-medium">\${job.title}</div>
                <div class="text-xs text-gray-500">\${job.extendedProps.customer}</div>
            </div>
        </div>
    \`).join('');
    
    // Initialize route map
    if (!window.routeMap) {
        window.routeMap = LandscapingApp.initializeMap('route-map', {
            center: [39.7817, -89.6501],
            zoom: 12
        });
    }
}

function hideRouteOptimization() {
    document.getElementById('route-optimization-modal').classList.add('hidden');
}

function optimizeRoute() {
    const selectedJobs = Array.from(document.querySelectorAll('.job-checkbox:checked'))
        .map(checkbox => checkbox.dataset.jobId);
    
    if (selectedJobs.length === 0) {
        LandscapingApp.showNotification('Please select jobs to optimize', 'warning');
        return;
    }
    
    // Mock route optimization
    const routeInfo = document.getElementById('route-info');
    routeInfo.textContent = \`Optimized route for \${selectedJobs.length} jobs. Estimated time: 6.5 hours, Distance: 45.2 miles\`;
    
    // Show optimized route on map
    if (window.routeMap) {
        // Mock route points
        const routePoints = [
            [39.7817, -89.6501],
            [39.7990, -89.6441],
            [39.7654, -89.6732]
        ];
        
        // Clear existing markers
        window.routeMap.eachLayer(function(layer) {
            if (layer instanceof L.Marker) {
                window.routeMap.removeLayer(layer);
            }
        });
        
        // Add route markers
        routePoints.forEach((point, index) => {
            L.marker(point).addTo(window.routeMap)
                .bindPopup(\`Stop \${index + 1}\`);
        });
        
        // Fit map to show all points
        window.routeMap.fitBounds(routePoints);
    }
    
    LandscapingApp.showNotification('Route optimized successfully!', 'success');
}
</script>
{{end}}`,

		"jobs_table.html": `{{define "content"}}
<div class="overflow-hidden">
    <div class="space-y-4">
        {{range .Data.jobs}}
        <div class="border border-gray-200 rounded-lg p-4 hover:shadow-md transition-shadow duration-150 ease-in-out">
            <div class="flex items-start justify-between">
                <div class="flex-1">
                    <div class="flex items-center space-x-3 mb-2">
                        <h4 class="text-lg font-medium text-gray-900">
                            <a href="/admin/jobs/{{.ID}}" class="hover:text-blue-600">{{.Title}}</a>
                        </h4>
                        <span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium {{if eq .Status "scheduled"}}bg-yellow-100 text-yellow-800{{else if eq .Status "in_progress"}}bg-blue-100 text-blue-800{{else if eq .Status "completed"}}bg-green-100 text-green-800{{else if eq .Status "cancelled"}}bg-red-100 text-red-800{{else}}bg-gray-100 text-gray-800{{end}}">
                            {{if eq .Status "scheduled"}}Scheduled{{else if eq .Status "in_progress"}}In Progress{{else if eq .Status "completed"}}Completed{{else if eq .Status "cancelled"}}Cancelled{{else}}{{.Status}}{{end}}
                        </span>
                        <span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium {{if eq .ServiceType "lawn_mowing"}}bg-green-100 text-green-800{{else if eq .ServiceType "hedge_trimming"}}bg-blue-100 text-blue-800{{else if eq .ServiceType "fertilizing"}}bg-purple-100 text-purple-800{{else if eq .ServiceType "snow_removal"}}bg-cyan-100 text-cyan-800{{else if eq .ServiceType "landscaping"}}bg-indigo-100 text-indigo-800{{else}}bg-gray-100 text-gray-800{{end}}">
                            {{if eq .ServiceType "lawn_mowing"}}Lawn Mowing{{else if eq .ServiceType "hedge_trimming"}}Hedge Trimming{{else if eq .ServiceType "fertilizing"}}Fertilizing{{else if eq .ServiceType "snow_removal"}}Snow Removal{{else if eq .ServiceType "landscaping"}}Landscaping{{else}}{{.ServiceType}}{{end}}
                        </span>
                    </div>
                    
                    <div class="text-sm text-gray-600 mb-2">
                        <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-2">
                            <div class="flex items-center space-x-1">
                                <svg class="w-4 h-4 text-gray-400" fill="currentColor" viewBox="0 0 20 20">
                                    <path fill-rule="evenodd" d="M10 9a3 3 0 100-6 3 3 0 000 6zm-7 9a7 7 0 1114 0H3z" clip-rule="evenodd"/>
                                </svg>
                                <span>{{.CustomerName}}</span>
                            </div>
                            <div class="flex items-center space-x-1">
                                <svg class="w-4 h-4 text-gray-400" fill="currentColor" viewBox="0 0 20 20">
                                    <path fill-rule="evenodd" d="M5.05 4.05a7 7 0 119.9 9.9L10 18.9l-4.95-4.95a7 7 0 010-9.9zM10 11a2 2 0 100-4 2 2 0 000 4z" clip-rule="evenodd"/>
                                </svg>
                                <span>{{.PropertyName}}</span>
                            </div>
                            <div class="flex items-center space-x-1">
                                <svg class="w-4 h-4 text-gray-400" fill="currentColor" viewBox="0 0 20 20">
                                    <path fill-rule="evenodd" d="M6 2a1 1 0 00-1 1v1H4a2 2 0 00-2 2v10a2 2 0 002 2h12a2 2 0 002-2V6a2 2 0 00-2-2h-1V3a1 1 0 10-2 0v1H7V3a1 1 0 00-1-1zm0 5a1 1 0 000 2h8a1 1 0 100-2H6z" clip-rule="evenodd"/>
                                </svg>
                                <span>{{.ScheduledDate}} at {{.ScheduledTime}}</span>
                            </div>
                            <div class="flex items-center space-x-1">
                                <svg class="w-4 h-4 text-gray-400" fill="currentColor" viewBox="0 0 20 20">
                                    <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm1-12a1 1 0 10-2 0v4a1 1 0 00.293.707l2.828 2.829a1 1 0 101.415-1.415L11 9.586V6z" clip-rule="evenodd"/>
                                </svg>
                                <span>{{.Duration}}</span>
                            </div>
                            <div class="flex items-center space-x-1">
                                <svg class="w-4 h-4 text-gray-400" fill="currentColor" viewBox="0 0 20 20">
                                    <path d="M8 9a3 3 0 100-6 3 3 0 000 6zM8 11a6 6 0 016 6H2a6 6 0 016-6zM16 7a1 1 0 10-2 0v1h-1a1 1 0 100 2h1v1a1 1 0 102 0v-1h1a1 1 0 100-2h-1V7z"/>
                                </svg>
                                <span>{{.CrewName}}</span>
                            </div>
                            <div class="flex items-center space-x-1">
                                <svg class="w-4 h-4 text-gray-400" fill="currentColor" viewBox="0 0 20 20">
                                    <path d="M8.433 7.418c.155-.103.346-.196.567-.267v1.698a2.305 2.305 0 01-.567-.267C8.07 8.34 8 8.114 8 8c0-.114.07-.34.433-.582zM11 12.849v-1.698c.22.071.412.164.567.267.364.243.433.468.433.582 0 .114-.07.34-.433.582a2.305 2.305 0 01-.567.267z"/>
                                    <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm1-13a1 1 0 10-2 0v.092a4.535 4.535 0 00-1.676.662C6.602 6.234 6 7.009 6 8c0 .99.602 1.765 1.324 2.246.48.32 1.054.545 1.676.662v1.941c-.391-.127-.68-.317-.843-.504a1 1 0 10-1.51 1.31c.562.649 1.413 1.076 2.353 1.253V15a1 1 0 102 0v-.092a4.535 4.535 0 001.676-.662C13.398 13.766 14 12.991 14 12c0-.99-.602-1.765-1.324-2.246A4.535 4.535 0 0011 9.092V7.151c.391.127.68.317.843.504a1 1 0 101.511-1.31c-.563-.649-1.413-1.076-2.354-1.253V5z" clip-rule="evenodd"/>
                                </svg>
                                <span>${{printf "%.0f" .Price}}</span>
                            </div>
                        </div>
                    </div>
                    
                    <p class="text-sm text-gray-600 mb-2">{{.Description}}</p>
                    
                    {{if .Notes}}
                    <div class="bg-gray-50 rounded-md p-2 text-sm text-gray-600">
                        <span class="font-medium">Notes:</span> {{.Notes}}
                    </div>
                    {{end}}
                </div>
                
                <div class="flex-shrink-0 ml-4">
                    <div class="flex items-center space-x-2">
                        {{if eq .Status "scheduled"}}
                        <button hx-post="/admin/jobs/{{.ID}}/start"
                                hx-target="closest .border"
                                hx-swap="outerHTML"
                                class="p-2 text-gray-400 hover:text-blue-600 rounded-full hover:bg-gray-100"
                                title="Start Job">
                            <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
                                <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM9.555 7.168A1 1 0 008 8v4a1 1 0 001.555.832l3-2a1 1 0 000-1.664l-3-2z" clip-rule="evenodd"/>
                            </svg>
                        </button>
                        {{else if eq .Status "in_progress"}}
                        <button hx-post="/admin/jobs/{{.ID}}/complete"
                                hx-target="closest .border"
                                hx-swap="outerHTML"
                                class="p-2 text-gray-400 hover:text-green-600 rounded-full hover:bg-gray-100"
                                title="Complete Job">
                            <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
                                <path fill-rule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clip-rule="evenodd"/>
                            </svg>
                        </button>
                        {{end}}
                        
                        <button onclick="showJobOnCalendar({{.ID}})"
                                class="p-2 text-gray-400 hover:text-purple-600 rounded-full hover:bg-gray-100"
                                title="View on Calendar">
                            <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
                                <path fill-rule="evenodd" d="M6 2a1 1 0 00-1 1v1H4a2 2 0 00-2 2v10a2 2 0 002 2h12a2 2 0 002-2V6a2 2 0 00-2-2h-1V3a1 1 0 10-2 0v1H7V3a1 1 0 00-1-1zm0 5a1 1 0 000 2h8a1 1 0 100-2H6z" clip-rule="evenodd"/>
                            </svg>
                        </button>
                        
                        <a href="/admin/jobs/{{.ID}}" 
                           class="p-2 text-gray-400 hover:text-blue-600 rounded-full hover:bg-gray-100"
                           title="View Details">
                            <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
                                <path d="M10 12a2 2 0 100-4 2 2 0 000 4z"/>
                                <path fill-rule="evenodd" d="M.458 10C1.732 5.943 5.522 3 10 3s8.268 2.943 9.542 7c-1.274 4.057-5.064 7-9.542 7S1.732 14.057.458 10zM14 10a4 4 0 11-8 0 4 4 0 018 0z" clip-rule="evenodd"/>
                            </svg>
                        </a>
                        
                        <div class="relative" x-data="{ open: false }">
                            <button @click="open = !open" 
                                    class="p-2 text-gray-400 hover:text-gray-600 rounded-full hover:bg-gray-100"
                                    title="More Actions">
                                <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
                                    <path d="M10 6a2 2 0 110-4 2 2 0 010 4zM10 12a2 2 0 110-4 2 2 0 010 4zM10 18a2 2 0 110-4 2 2 0 010 4z"/>
                                </svg>
                            </button>
                            <div x-show="open" @click.away="open = false" x-cloak
                                 class="absolute right-0 mt-2 w-48 bg-white rounded-md shadow-lg py-1 z-10 border border-gray-200">
                                <a href="/admin/jobs/{{.ID}}/edit" class="block px-4 py-2 text-sm text-gray-700 hover:bg-gray-100">Edit Job</a>
                                <a href="/admin/jobs/{{.ID}}/duplicate" class="block px-4 py-2 text-sm text-gray-700 hover:bg-gray-100">Duplicate</a>
                                <button hx-post="/admin/jobs/{{.ID}}/reschedule"
                                        hx-target="#reschedule-modal"
                                        class="w-full text-left block px-4 py-2 text-sm text-gray-700 hover:bg-gray-100">Reschedule</button>
                                {{if ne .Status "cancelled"}}
                                <button hx-post="/admin/jobs/{{.ID}}/cancel"
                                        hx-confirm="Are you sure you want to cancel this job?"
                                        hx-target="closest .border"
                                        hx-swap="outerHTML"
                                        class="w-full text-left block px-4 py-2 text-sm text-red-700 hover:bg-red-50">Cancel Job</button>
                                {{end}}
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
        {{end}}
        
        {{if eq (len .Data.jobs) 0}}
        <div class="text-center py-12">
            <svg class="mx-auto h-12 w-12 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3a2 2 0 012-2h4a2 2 0 012 2v4m-6 9l2 2 4-4m6-7H2a1 1 0 00-1 1v10a1 1 0 001 1h20a1 1 0 001-1V8a1 1 0 00-1-1z"/>
            </svg>
            <h3 class="mt-2 text-sm font-medium text-gray-900">No jobs found</h3>
            <p class="mt-1 text-sm text-gray-500">Get started by scheduling your first job.</p>
            <div class="mt-6">
                <a href="/admin/jobs/new" 
                   class="inline-flex items-center px-4 py-2 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500">
                    <svg class="mr-2 h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"/>
                    </svg>
                    Schedule Job
                </a>
            </div>
        </div>
        {{end}}
    </div>
    
    <!-- Pagination -->
    {{if gt .Data.totalCount .Data.limit}}
    <div class="mt-6 flex items-center justify-between border-t border-gray-200 pt-6">
        <div class="flex-1 flex justify-between sm:hidden">
            {{if gt .Data.currentPage 1}}
            <button hx-get="/api/v1/jobs/table?page={{add .Data.currentPage -1}}"
                    hx-target="#jobs-content"
                    hx-include="[data-search-filter], #job-search"
                    class="relative inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50">
                Previous
            </button>
            {{end}}
            {{if lt .Data.currentPage .Data.totalPages}}
            <button hx-get="/api/v1/jobs/table?page={{add .Data.currentPage 1}}"
                    hx-target="#jobs-content"
                    hx-include="[data-search-filter], #job-search"
                    class="ml-3 relative inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50">
                Next
            </button>
            {{end}}
        </div>
        
        <div class="hidden sm:flex-1 sm:flex sm:items-center sm:justify-between">
            <div>
                <p class="text-sm text-gray-700">
                    Showing <span class="font-medium">{{.Data.offset}}</span> to 
                    <span class="font-medium">{{add .Data.offset (len .Data.jobs)}}</span> of 
                    <span class="font-medium">{{.Data.totalCount}}</span> jobs
                </p>
            </div>
            <div>
                <nav class="relative z-0 inline-flex rounded-md shadow-sm -space-x-px">
                    {{if gt .Data.currentPage 1}}
                    <button hx-get="/api/v1/jobs/table?page={{add .Data.currentPage -1}}"
                            hx-target="#jobs-content"
                            hx-include="[data-search-filter], #job-search"
                            class="relative inline-flex items-center px-2 py-2 rounded-l-md border border-gray-300 bg-white text-sm font-medium text-gray-500 hover:bg-gray-50">
                        Previous
                    </button>
                    {{end}}
                    
                    {{range $page := .Data.pageNumbers}}
                    <button hx-get="/api/v1/jobs/table?page={{$page}}"
                            hx-target="#jobs-content"
                            hx-include="[data-search-filter], #job-search"
                            class="relative inline-flex items-center px-4 py-2 border text-sm font-medium {{if eq $page $.Data.currentPage}}bg-blue-50 border-blue-500 text-blue-600{{else}}bg-white border-gray-300 text-gray-500 hover:bg-gray-50{{end}}">
                        {{$page}}
                    </button>
                    {{end}}
                    
                    {{if lt .Data.currentPage .Data.totalPages}}
                    <button hx-get="/api/v1/jobs/table?page={{add .Data.currentPage 1}}"
                            hx-target="#jobs-content"
                            hx-include="[data-search-filter], #job-search"
                            class="relative inline-flex items-center px-2 py-2 rounded-r-md border border-gray-300 bg-white text-sm font-medium text-gray-500 hover:bg-gray-50">
                        Next
                    </button>
                    {{end}}
                </nav>
            </div>
        </div>
    </div>
    {{end}}
</div>

<script>
function showJobOnCalendar(jobId) {
    // Redirect to calendar view and highlight the specific job
    window.location.href = '/admin/jobs/calendar?highlight=' + jobId;
}
</script>
{{end}}`,
	}

	for name, content := range templates {
		tmpl, err := template.New(name).Funcs(ts.funcs).Parse(content)
		if err != nil {
			return err
		}
		ts.templates[name] = tmpl
	}

	return nil
}

func (ts *TemplateService) Render(templateName string, data TemplateData) (string, error) {
	tmpl, exists := ts.templates[templateName]
	if !exists {
		// Try to find base template and the specific template
		if baseTemplate, ok := ts.templates["base.html"]; ok {
			if specificTemplate, ok := ts.templates[templateName]; ok {
				// Clone base template and add specific template
				combined, err := baseTemplate.Clone()
				if err != nil {
					return "", err
				}
				for _, t := range specificTemplate.Templates() {
					combined, err = combined.AddParseTree(t.Name(), t.Tree)
					if err != nil {
						return "", err
					}
				}
				tmpl = combined
			}
		}
		if tmpl == nil {
			return "", template.ErrNoSuchTemplate
		}
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func createTemplateFuncs() template.FuncMap {
	return template.FuncMap{
		"substr": func(s string, start, length int) string {
			if start >= len(s) {
				return ""
			}
			end := start + length
			if end > len(s) {
				end = len(s)
			}
			return s[start:end]
		},
		"add": func(a, b int) int {
			return a + b
		},
		"formatCurrency": func(amount float64) string {
			return fmt.Sprintf("$%.2f", amount)
		},
		"formatDate": func(date interface{}) string {
			if date == nil {
				return "N/A"
			}
			// Handle string dates
			if str, ok := date.(string); ok {
				if str == "" {
					return "N/A"
				}
				return str // Return as-is for now, could parse if needed
			}
			return "N/A"
		},
		"eq": func(a, b interface{}) bool {
			return a == b
		},
		"lt": func(a, b int) bool {
			return a < b
		},
		"gt": func(a, b int) bool {
			return a > b
		},
		"len": func(slice interface{}) int {
			if slice == nil {
				return 0
			}
			// Use reflection to get length
			return 0 // Simplified for now
		},
		"printf": func(format string, args ...interface{}) string {
			return fmt.Sprintf(format, args...)
		},
		"ne": func(a, b interface{}) bool {
			return a != b
		},
	}
}