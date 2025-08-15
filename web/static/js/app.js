// Main JavaScript application for the Landscaping App Web Interface

// Initialize the application when DOM is loaded
document.addEventListener('DOMContentLoaded', function() {
    console.log('Landscaping App Web Interface loaded');
    
    // Initialize HTMX event listeners
    initializeHTMXListeners();
    
    // Initialize Alpine.js components
    initializeAlpineComponents();
    
    // Initialize WebSocket connections
    initializeWebSocket();
    
    // Initialize notification system
    initializeNotifications();
    
    // Initialize navigation
    initializeNavigation();
    
    // Initialize data visualization
    initializeDataVisualization();
});

// HTMX Event Listeners
function initializeHTMXListeners() {
    // Show loading indicators
    document.addEventListener('htmx:beforeRequest', function(event) {
        showLoadingIndicator(event.detail.elt);
    });
    
    // Hide loading indicators
    document.addEventListener('htmx:afterRequest', function(event) {
        hideLoadingIndicator(event.detail.elt);
    });
    
    // Handle errors
    document.addEventListener('htmx:responseError', function(event) {
        console.error('HTMX Error:', event.detail);
        showNotification('An error occurred. Please try again.', 'error');
    });
    
    // Handle successful updates
    document.addEventListener('htmx:afterSwap', function(event) {
        // Re-initialize any components in the swapped content
        initializeSwappedContent(event.detail.target);
    });
}

// Alpine.js Components
function initializeAlpineComponents() {
    // Global Alpine.js data and functions
    window.Alpine = window.Alpine || {};
    
    // Mobile sidebar toggle
    window.Alpine.data('sidebar', () => ({
        open: false,
        toggle() {
            this.open = !this.open;
        },
        close() {
            this.open = false;
        }
    }));
    
    // Notification component
    window.Alpine.data('notification', () => ({
        visible: false,
        message: '',
        type: 'info',
        show(message, type = 'info') {
            this.message = message;
            this.type = type;
            this.visible = true;
            setTimeout(() => {
                this.visible = false;
            }, 5000);
        },
        hide() {
            this.visible = false;
        }
    }));
    
    // Search component
    window.Alpine.data('search', () => ({
        query: '',
        results: [],
        loading: false,
        debounceTimer: null,
        
        search() {
            clearTimeout(this.debounceTimer);
            this.debounceTimer = setTimeout(() => {
                this.performSearch();
            }, 300);
        },
        
        performSearch() {
            if (this.query.length < 2) {
                this.results = [];
                return;
            }
            
            this.loading = true;
            // Search logic here
            setTimeout(() => {
                this.loading = false;
            }, 500);
        }
    }));
}

// WebSocket Initialization
function initializeWebSocket() {
    if (typeof WebSocket === 'undefined') {
        console.warn('WebSocket not supported');
        return;
    }
    
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${protocol}//${window.location.host}/ws`;
    
    let ws = null;
    let reconnectAttempts = 0;
    const maxReconnectAttempts = 5;
    
    function connect() {
        try {
            ws = new WebSocket(wsUrl);
            
            ws.onopen = function() {
                console.log('WebSocket connected');
                reconnectAttempts = 0;
                // Send authentication if needed
                ws.send(JSON.stringify({
                    type: 'auth',
                    token: getAuthToken()
                }));
            };
            
            ws.onmessage = function(event) {
                try {
                    const data = JSON.parse(event.data);
                    handleWebSocketMessage(data);
                } catch (e) {
                    console.error('Error parsing WebSocket message:', e);
                }
            };
            
            ws.onclose = function() {
                console.log('WebSocket disconnected');
                // Attempt to reconnect
                if (reconnectAttempts < maxReconnectAttempts) {
                    setTimeout(() => {
                        reconnectAttempts++;
                        connect();
                    }, 1000 * Math.pow(2, reconnectAttempts));
                }
            };
            
            ws.onerror = function(error) {
                console.error('WebSocket error:', error);
            };
            
        } catch (error) {
            console.error('Failed to connect WebSocket:', error);
        }
    }
    
    // Initialize connection
    connect();
    
    // Store WebSocket instance globally
    window.ws = ws;
}

// Handle WebSocket Messages
function handleWebSocketMessage(data) {
    switch (data.type) {
        case 'notification':
            showNotification(data.data.message, data.data.type || 'info');
            break;
            
        case 'job_update':
            handleJobUpdate(data.data);
            break;
            
        case 'dashboard_update':
            handleDashboardUpdate(data.data);
            break;
            
        case 'chat_response':
            handleChatResponse(data.data);
            break;
            
        default:
            console.log('Unknown WebSocket message type:', data.type);
    }
}

// Notification System
function initializeNotifications() {
    // Create notification container if it doesn't exist
    if (!document.getElementById('notification-container')) {
        const container = document.createElement('div');
        container.id = 'notification-container';
        container.className = 'fixed top-4 right-4 z-50 space-y-2';
        document.body.appendChild(container);
    }
}

function showNotification(message, type = 'info') {
    const container = document.getElementById('notification-container');
    if (!container) return;
    
    const notification = document.createElement('div');
    notification.className = `notification ${type} px-4 py-3 rounded-md shadow-lg max-w-sm`;
    notification.innerHTML = `
        <div class="flex items-center">
            <div class="flex-1">
                <p class="text-sm font-medium">${message}</p>
            </div>
            <button class="ml-3 text-sm" onclick="this.parentElement.parentElement.remove()">
                ×
            </button>
        </div>
    `;
    
    container.appendChild(notification);
    
    // Auto-remove after 5 seconds
    setTimeout(() => {
        if (notification.parentElement) {
            notification.remove();
        }
    }, 5000);
}

// Loading Indicators
function showLoadingIndicator(element) {
    const indicator = element.querySelector('.htmx-indicator');
    if (indicator) {
        indicator.style.opacity = '1';
    } else {
        // Create a simple spinner
        const spinner = document.createElement('div');
        spinner.className = 'htmx-indicator spinner';
        element.appendChild(spinner);
    }
}

function hideLoadingIndicator(element) {
    const indicator = element.querySelector('.htmx-indicator');
    if (indicator) {
        indicator.style.opacity = '0';
        setTimeout(() => {
            if (indicator.parentElement) {
                indicator.remove();
            }
        }, 500);
    }
}

// Utility Functions
function getAuthToken() {
    // Get auth token from cookie or localStorage
    const cookies = document.cookie.split(';');
    for (let cookie of cookies) {
        const [name, value] = cookie.trim().split('=');
        if (name === 'session_token') {
            return value;
        }
    }
    return null;
}

function initializeSwappedContent(element) {
    // Re-initialize any JavaScript components in swapped content
    // This is called after HTMX swaps content
    
    // Re-initialize Alpine.js components
    if (window.Alpine) {
        window.Alpine.initTree(element);
    }
    
    // Re-initialize any other components as needed
}

// Job Update Handlers
function handleJobUpdate(data) {
    // Update job status in real-time
    const jobElements = document.querySelectorAll(`[data-job-id="${data.id}"]`);
    jobElements.forEach(element => {
        const statusElement = element.querySelector('.job-status');
        if (statusElement) {
            statusElement.textContent = data.status;
            statusElement.className = `job-status status-${data.status}`;
        }
    });
    
    // Show notification for status changes
    showNotification(`Job ${data.title} status updated to ${data.status}`, 'info');
}

// Dashboard Update Handlers
function handleDashboardUpdate(data) {
    // Update dashboard statistics
    Object.keys(data).forEach(key => {
        const element = document.querySelector(`[data-stat="${key}"]`);
        if (element) {
            element.textContent = data[key];
        }
    });
}

// Chat Response Handlers
function handleChatResponse(data) {
    // Handle AI chat responses
    const chatContainer = document.getElementById('chat-messages');
    if (chatContainer) {
        const messageElement = document.createElement('div');
        messageElement.className = 'chat-message ai-message';
        messageElement.innerHTML = `
            <div class="message-content">
                <p>${data.message}</p>
                <span class="message-time">${new Date(data.timestamp).toLocaleTimeString()}</span>
            </div>
        `;
        chatContainer.appendChild(messageElement);
        chatContainer.scrollTop = chatContainer.scrollHeight;
    }
}

// Form Validation
function validateForm(formElement) {
    const inputs = formElement.querySelectorAll('input[required], select[required], textarea[required]');
    let isValid = true;
    
    inputs.forEach(input => {
        if (!input.value.trim()) {
            isValid = false;
            input.classList.add('border-red-500');
            showFieldError(input, 'This field is required');
        } else {
            input.classList.remove('border-red-500');
            hideFieldError(input);
        }
    });
    
    return isValid;
}

function showFieldError(input, message) {
    hideFieldError(input); // Remove existing error
    
    const error = document.createElement('div');
    error.className = 'field-error text-red-500 text-sm mt-1';
    error.textContent = message;
    
    input.parentElement.appendChild(error);
}

function hideFieldError(input) {
    const error = input.parentElement.querySelector('.field-error');
    if (error) {
        error.remove();
    }
}

// Navigation Management
function initializeNavigation() {
    // Set active navigation state
    setActiveNavigation();
    
    // Add keyboard shortcuts
    initializeKeyboardShortcuts();
}

function setActiveNavigation() {
    const currentPath = window.location.pathname;
    const navItems = document.querySelectorAll('.nav-item');
    
    navItems.forEach(item => {
        item.classList.remove('bg-gray-700', 'text-white');
        item.classList.add('text-gray-300');
        
        if (item.getAttribute('href') === currentPath || 
            (currentPath.startsWith(item.getAttribute('href')) && item.getAttribute('href') !== '/admin')) {
            item.classList.remove('text-gray-300');
            item.classList.add('bg-gray-700', 'text-white');
        }
    });
}

function initializeKeyboardShortcuts() {
    document.addEventListener('keydown', function(e) {
        // Ctrl/Cmd + shortcuts
        if (e.ctrlKey || e.metaKey) {
            switch(e.key) {
                case '1':
                    e.preventDefault();
                    window.location.href = '/admin';
                    break;
                case '2':
                    e.preventDefault();
                    window.location.href = '/admin/customers';
                    break;
                case '3':
                    e.preventDefault();
                    window.location.href = '/admin/properties';
                    break;
                case '4':
                    e.preventDefault();
                    window.location.href = '/admin/jobs';
                    break;
                case '5':
                    e.preventDefault();
                    window.location.href = '/admin/jobs/calendar';
                    break;
                case '/':
                    e.preventDefault();
                    const searchInput = document.querySelector('#global-search');
                    if (searchInput) {
                        searchInput.focus();
                    }
                    break;
            }
        }
    });
}

// Data Visualization
function initializeDataVisualization() {
    // Initialize Chart.js if available
    if (typeof Chart !== 'undefined') {
        initializeCharts();
    } else {
        // Load Chart.js dynamically
        loadScript('https://cdn.jsdelivr.net/npm/chart.js', initializeCharts);
    }
}

function loadScript(src, callback) {
    const script = document.createElement('script');
    script.src = src;
    script.onload = callback;
    script.onerror = function() {
        console.error('Failed to load script:', src);
    };
    document.head.appendChild(script);
}

function initializeCharts() {
    // Revenue Chart
    const revenueCanvas = document.getElementById('revenue-chart');
    if (revenueCanvas) {
        createRevenueChart(revenueCanvas);
    }
    
    // Jobs Progress Chart
    const jobsCanvas = document.getElementById('jobs-chart');
    if (jobsCanvas) {
        createJobsChart(jobsCanvas);
    }
}

function createRevenueChart(canvas) {
    new Chart(canvas, {
        type: 'line',
        data: {
            labels: ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun'],
            datasets: [{
                label: 'Revenue',
                data: [12000, 19000, 15000, 25000, 22000, 30000],
                borderColor: 'rgb(34, 197, 94)',
                backgroundColor: 'rgba(34, 197, 94, 0.1)',
                tension: 0.1
            }]
        },
        options: {
            responsive: true,
            plugins: {
                legend: {
                    display: false
                }
            },
            scales: {
                y: {
                    beginAtZero: true,
                    ticks: {
                        callback: function(value) {
                            return '$' + value.toLocaleString();
                        }
                    }
                }
            }
        }
    });
}

function createJobsChart(canvas) {
    new Chart(canvas, {
        type: 'doughnut',
        data: {
            labels: ['Completed', 'In Progress', 'Scheduled', 'Pending'],
            datasets: [{
                data: [65, 20, 10, 5],
                backgroundColor: [
                    'rgb(34, 197, 94)',
                    'rgb(59, 130, 246)',
                    'rgb(245, 158, 11)',
                    'rgb(239, 68, 68)'
                ]
            }]
        },
        options: {
            responsive: true,
            plugins: {
                legend: {
                    position: 'bottom'
                }
            }
        }
    });
}

// Map Integration
function initializeMap(containerId, options = {}) {
    if (typeof L !== 'undefined') {
        return createLeafletMap(containerId, options);
    } else {
        // Load Leaflet dynamically
        loadCSS('https://unpkg.com/leaflet@1.9.4/dist/leaflet.css');
        loadScript('https://unpkg.com/leaflet@1.9.4/dist/leaflet.js', function() {
            createLeafletMap(containerId, options);
        });
    }
}

function loadCSS(href) {
    const link = document.createElement('link');
    link.rel = 'stylesheet';
    link.href = href;
    document.head.appendChild(link);
}

function createLeafletMap(containerId, options) {
    const container = document.getElementById(containerId);
    if (!container) return null;
    
    const defaultOptions = {
        center: [40.7128, -74.0060], // New York City
        zoom: 13,
        ...options
    };
    
    const map = L.map(containerId).setView(defaultOptions.center, defaultOptions.zoom);
    
    L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
        attribution: '© OpenStreetMap contributors'
    }).addTo(map);
    
    return map;
}

// File Upload Handler
function initializeFileUpload() {
    const uploadAreas = document.querySelectorAll('.file-upload-area');
    
    uploadAreas.forEach(area => {
        const input = area.querySelector('input[type="file"]');
        
        area.addEventListener('dragover', function(e) {
            e.preventDefault();
            area.classList.add('dragover');
        });
        
        area.addEventListener('dragleave', function(e) {
            e.preventDefault();
            area.classList.remove('dragover');
        });
        
        area.addEventListener('drop', function(e) {
            e.preventDefault();
            area.classList.remove('dragover');
            
            const files = e.dataTransfer.files;
            if (input && files.length > 0) {
                input.files = files;
                handleFileUpload(input);
            }
        });
        
        if (input) {
            input.addEventListener('change', function() {
                handleFileUpload(input);
            });
        }
    });
}

function handleFileUpload(input) {
    const files = input.files;
    const preview = input.closest('.file-upload-area').querySelector('.file-preview');
    
    if (preview) {
        preview.innerHTML = '';
        
        Array.from(files).forEach(file => {
            const fileItem = document.createElement('div');
            fileItem.className = 'file-item flex items-center p-2 bg-gray-50 rounded';
            
            if (file.type.startsWith('image/')) {
                const img = document.createElement('img');
                img.src = URL.createObjectURL(file);
                img.className = 'w-12 h-12 object-cover rounded mr-3';
                fileItem.appendChild(img);
            }
            
            const info = document.createElement('div');
            info.innerHTML = `
                <div class="text-sm font-medium">${file.name}</div>
                <div class="text-xs text-gray-500">${formatFileSize(file.size)}</div>
            `;
            fileItem.appendChild(info);
            
            preview.appendChild(fileItem);
        });
    }
}

function formatFileSize(bytes) {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}

// Calendar Integration
function initializeCalendar(containerId, options = {}) {
    if (typeof FullCalendar !== 'undefined') {
        return createFullCalendar(containerId, options);
    } else {
        // Load FullCalendar dynamically
        loadCSS('https://cdn.jsdelivr.net/npm/fullcalendar@6.1.10/index.global.min.css');
        loadScript('https://cdn.jsdelivr.net/npm/fullcalendar@6.1.10/index.global.min.js', function() {
            createFullCalendar(containerId, options);
        });
    }
}

function createFullCalendar(containerId, options) {
    const container = document.getElementById(containerId);
    if (!container) return null;
    
    const calendar = new FullCalendar.Calendar(container, {
        initialView: 'dayGridMonth',
        headerToolbar: {
            left: 'prev,next today',
            center: 'title',
            right: 'dayGridMonth,timeGridWeek,timeGridDay'
        },
        editable: true,
        droppable: true,
        eventDrop: function(info) {
            // Handle job reschedule
            updateJobSchedule(info.event.id, info.event.start);
        },
        ...options
    });
    
    calendar.render();
    return calendar;
}

function updateJobSchedule(jobId, newDate) {
    // Send HTMX request to update job schedule
    htmx.ajax('PUT', `/api/v1/jobs/${jobId}/schedule`, {
        values: {
            scheduled_date: newDate.toISOString()
        },
        swap: 'none',
        target: null
    }).then(function() {
        showNotification('Job schedule updated successfully', 'success');
    }).catch(function() {
        showNotification('Failed to update job schedule', 'error');
    });
}

// Real-time Features
function enableRealTimeUpdates() {
    // Set up periodic updates for dashboard
    setInterval(updateDashboardStats, 30000); // Every 30 seconds
    
    // Set up notification polling as fallback
    if (!window.ws || window.ws.readyState !== WebSocket.OPEN) {
        setInterval(checkNotifications, 60000); // Every minute
    }
}

function updateDashboardStats() {
    const statsContainer = document.getElementById('stats-container');
    if (statsContainer) {
        htmx.ajax('GET', '/api/v1/dashboard/stats', {
            target: '#stats-container',
            swap: 'innerHTML'
        });
    }
}

function checkNotifications() {
    htmx.ajax('GET', '/api/v1/notifications', {
        target: null,
        swap: 'none'
    }).then(function(response) {
        if (response.notifications && response.notifications.length > 0) {
            response.notifications.forEach(notification => {
                showNotification(notification.message, notification.type);
            });
        }
    });
}

// Export functions for global use
window.LandscapingApp = {
    showNotification,
    validateForm,
    getAuthToken,
    initializeMap,
    initializeCalendar,
    initializeFileUpload,
    enableRealTimeUpdates,
    setActiveNavigation
};