import http from 'k6/http';
import { check, sleep, group } from 'k6';
import { Rate, Trend } from 'k6/metrics';
import { htmlReport } from 'https://raw.githubusercontent.com/benc-uk/k6-reporter/main/dist/bundle.js';

// Custom metrics
const errorRate = new Rate('error_rate');
const responseTimeThreshold = new Trend('response_time_threshold', true);

// Configuration for different test scenarios
export let options = {
  scenarios: {
    // Smoke test - basic functionality
    smoke: {
      executor: 'constant-vus',
      vus: 1,
      duration: '30s',
      tags: { test_type: 'smoke' },
      env: { SCENARIO: 'smoke' },
    },
    
    // Load test - normal expected load
    load: {
      executor: 'constant-vus',
      vus: 50,
      duration: '5m',
      tags: { test_type: 'load' },
      env: { SCENARIO: 'load' },
    },
    
    // Stress test - beyond normal capacity
    stress: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '2m', target: 100 },
        { duration: '5m', target: 200 },
        { duration: '2m', target: 300 },
        { duration: '5m', target: 300 },
        { duration: '2m', target: 0 },
      ],
      tags: { test_type: 'stress' },
      env: { SCENARIO: 'stress' },
    },
    
    // Spike test - sudden traffic spikes
    spike: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '10s', target: 100 },
        { duration: '1m', target: 100 },
        { duration: '10s', target: 1000 }, // Sudden spike
        { duration: '1m', target: 1000 },
        { duration: '10s', target: 100 },
        { duration: '1m', target: 100 },
        { duration: '10s', target: 0 },
      ],
      tags: { test_type: 'spike' },
      env: { SCENARIO: 'spike' },
    },
    
    // Volume test - large amount of data
    volume: {
      executor: 'constant-vus',
      vus: 20,
      duration: '10m',
      tags: { test_type: 'volume' },
      env: { SCENARIO: 'volume' },
    },
    
    // Soak test - extended duration
    soak: {
      executor: 'constant-vus',
      vus: 30,
      duration: '1h',
      tags: { test_type: 'soak' },
      env: { SCENARIO: 'soak' },
    }
  },
  
  thresholds: {
    // HTTP request duration should be less than 2s for 95% of requests
    'http_req_duration': ['p(95)<2000'],
    
    // HTTP request duration should be less than 500ms for 90% of requests
    'http_req_duration{scenario:load}': ['p(90)<500'],
    
    // Error rate should be less than 5%
    'error_rate': ['rate<0.05'],
    
    // At least 95% of requests should complete successfully
    'http_req_failed': ['rate<0.05'],
    
    // API endpoints should respond within SLA
    'response_time_threshold': ['p(95)<1000'],
    
    // Specific thresholds for different scenarios
    'http_req_duration{scenario:smoke}': ['p(95)<1000'],
    'http_req_duration{scenario:stress}': ['p(95)<5000'],
    'http_req_duration{scenario:spike}': ['p(95)<3000'],
  }
};

// Test configuration
const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const API_BASE = `${BASE_URL}/api/v1`;

// Test data
const TEST_USERS = [
  { email: 'test1@example.com', password: 'password123', role: 'admin' },
  { email: 'test2@example.com', password: 'password123', role: 'user' },
  { email: 'test3@example.com', password: 'password123', role: 'user' },
  { email: 'customer1@example.com', password: 'password123', role: 'customer' },
  { email: 'crew1@example.com', password: 'password123', role: 'crew' },
];

// Authentication token cache
let authTokens = new Map();

// Utility functions
function getRandomUser() {
  return TEST_USERS[Math.floor(Math.random() * TEST_USERS.length)];
}

function getAuthToken(user) {
  const key = `${user.email}:${user.role}`;
  
  if (!authTokens.has(key)) {
    const loginResponse = http.post(`${API_BASE}/auth/login`, JSON.stringify({
      email: user.email,
      password: user.password,
    }), {
      headers: { 'Content-Type': 'application/json' },
    });
    
    if (loginResponse.status === 200) {
      const token = JSON.parse(loginResponse.body).access_token;
      authTokens.set(key, token);
      return token;
    }
    return null;
  }
  
  return authTokens.get(key);
}

function createAuthHeaders(user) {
  const token = getAuthToken(user);
  return token ? {
    'Authorization': `Bearer ${token}`,
    'Content-Type': 'application/json',
  } : { 'Content-Type': 'application/json' };
}

function generateCustomerData() {
  return {
    email: `customer.${Date.now()}.${Math.random()}@example.com`,
    name: `Customer ${Math.floor(Math.random() * 10000)}`,
    phone: `+1${Math.floor(Math.random() * 9000000000) + 1000000000}`,
    address: `${Math.floor(Math.random() * 9999) + 1} Main St`,
    city: 'Test City',
    state: 'TS',
    zip_code: `${Math.floor(Math.random() * 90000) + 10000}`,
    country: 'USA',
  };
}

function generateJobData(customerId, propertyId) {
  const services = ['mowing', 'trimming', 'edging', 'fertilizing', 'irrigation'];
  const randomServices = services.sort(() => 0.5 - Math.random()).slice(0, Math.floor(Math.random() * 3) + 1);
  
  return {
    customer_id: customerId,
    property_id: propertyId,
    title: `Landscaping Service ${Math.floor(Math.random() * 1000)}`,
    description: 'Regular landscaping maintenance service',
    services: randomServices,
    priority: ['low', 'medium', 'high'][Math.floor(Math.random() * 3)],
    estimated_duration: Math.floor(Math.random() * 240) + 60, // 1-4 hours
    scheduled_date: new Date(Date.now() + Math.floor(Math.random() * 30) * 24 * 60 * 60 * 1000).toISOString(),
  };
}

// Test scenarios
export default function() {
  const scenario = __ENV.SCENARIO || 'load';
  
  switch(scenario) {
    case 'smoke':
      runSmokeTest();
      break;
    case 'load':
      runLoadTest();
      break;
    case 'stress':
      runStressTest();
      break;
    case 'spike':
      runSpikeTest();
      break;
    case 'volume':
      runVolumeTest();
      break;
    case 'soak':
      runSoakTest();
      break;
    default:
      runLoadTest();
  }
  
  sleep(1);
}

function runSmokeTest() {
  group('Smoke Test - Basic API Health', function() {
    // Health check
    group('Health Check', function() {
      const response = http.get(`${BASE_URL}/health`);
      check(response, {
        'health check status is 200': (r) => r.status === 200,
        'health check response time < 100ms': (r) => r.timings.duration < 100,
      });
      errorRate.add(response.status !== 200);
    });
    
    // Authentication
    group('Authentication', function() {
      const user = getRandomUser();
      const response = http.post(`${API_BASE}/auth/login`, JSON.stringify({
        email: user.email,
        password: user.password,
      }), {
        headers: { 'Content-Type': 'application/json' },
      });
      
      check(response, {
        'login status is 200': (r) => r.status === 200,
        'login response has token': (r) => JSON.parse(r.body).access_token !== undefined,
      });
      errorRate.add(response.status !== 200);
    });
    
    // Basic API endpoint
    group('Basic API Access', function() {
      const user = getRandomUser();
      const headers = createAuthHeaders(user);
      
      const response = http.get(`${API_BASE}/customers`, { headers });
      check(response, {
        'customers endpoint accessible': (r) => r.status === 200 || r.status === 401,
      });
    });
  });
}

function runLoadTest() {
  group('Load Test - Normal Operations', function() {
    const user = getRandomUser();
    const headers = createAuthHeaders(user);
    
    if (!headers.Authorization) {
      console.log('Failed to authenticate user for load test');
      return;
    }
    
    // Customer operations (70% of traffic)
    if (Math.random() < 0.7) {
      group('Customer Operations', function() {
        // List customers
        const listResponse = http.get(`${API_BASE}/customers`, { headers });
        check(listResponse, {
          'list customers success': (r) => r.status === 200,
          'list customers response time acceptable': (r) => r.timings.duration < 1000,
        });
        errorRate.add(listResponse.status !== 200);
        responseTimeThreshold.add(listResponse.timings.duration);
        
        // Create customer (20% of customer operations)
        if (Math.random() < 0.2) {
          const customerData = generateCustomerData();
          const createResponse = http.post(`${API_BASE}/customers`, JSON.stringify(customerData), { headers });
          
          check(createResponse, {
            'create customer success': (r) => r.status === 201,
            'create customer response time acceptable': (r) => r.timings.duration < 2000,
          });
          errorRate.add(createResponse.status !== 201);
          responseTimeThreshold.add(createResponse.timings.duration);
        }
      });
    }
    
    // Job operations (20% of traffic)
    else if (Math.random() < 0.9) {
      group('Job Operations', function() {
        const listResponse = http.get(`${API_BASE}/jobs`, { headers });
        check(listResponse, {
          'list jobs success': (r) => r.status === 200,
          'list jobs response time acceptable': (r) => r.timings.duration < 1500,
        });
        errorRate.add(listResponse.status !== 200);
        responseTimeThreshold.add(listResponse.timings.duration);
      });
    }
    
    // Invoice/Payment operations (10% of traffic)
    else {
      group('Financial Operations', function() {
        const invoicesResponse = http.get(`${API_BASE}/invoices`, { headers });
        check(invoicesResponse, {
          'list invoices success': (r) => r.status === 200,
        });
        errorRate.add(invoicesResponse.status !== 200);
        
        const paymentsResponse = http.get(`${API_BASE}/payments`, { headers });
        check(paymentsResponse, {
          'list payments success': (r) => r.status === 200,
        });
        errorRate.add(paymentsResponse.status !== 200);
      });
    }
  });
}

function runStressTest() {
  group('Stress Test - High Load', function() {
    const user = getRandomUser();
    const headers = createAuthHeaders(user);
    
    // Concurrent operations to stress the system
    const operations = [
      () => http.get(`${API_BASE}/customers`, { headers }),
      () => http.get(`${API_BASE}/jobs`, { headers }),
      () => http.get(`${API_BASE}/properties`, { headers }),
      () => http.get(`${API_BASE}/quotes`, { headers }),
      () => http.get(`${API_BASE}/invoices`, { headers }),
    ];
    
    // Execute multiple operations rapidly
    operations.forEach(operation => {
      const response = operation();
      check(response, {
        'stress test endpoint responds': (r) => r.status < 500,
        'stress test response time under threshold': (r) => r.timings.duration < 5000,
      });
      errorRate.add(response.status >= 500);
    });
  });
}

function runSpikeTest() {
  group('Spike Test - Sudden Load Increase', function() {
    const user = getRandomUser();
    const headers = createAuthHeaders(user);
    
    // Simulate spike behavior - rapid requests
    for (let i = 0; i < 5; i++) {
      const response = http.get(`${API_BASE}/customers`, { headers });
      check(response, {
        'spike test handles request': (r) => r.status < 500,
        'spike test reasonable response time': (r) => r.timings.duration < 3000,
      });
      errorRate.add(response.status >= 500);
    }
  });
}

function runVolumeTest() {
  group('Volume Test - Large Data Operations', function() {
    const user = getRandomUser();
    const headers = createAuthHeaders(user);
    
    // Create multiple customers to test data volume handling
    for (let i = 0; i < 3; i++) {
      const customerData = generateCustomerData();
      const response = http.post(`${API_BASE}/customers`, JSON.stringify(customerData), { headers });
      
      check(response, {
        'volume test customer creation': (r) => r.status === 201,
      });
      errorRate.add(response.status !== 201);
    }
    
    // List with pagination
    const listResponse = http.get(`${API_BASE}/customers?limit=50&offset=0`, { headers });
    check(listResponse, {
      'volume test pagination works': (r) => r.status === 200,
      'volume test large list reasonable time': (r) => r.timings.duration < 2000,
    });
    errorRate.add(listResponse.status !== 200);
  });
}

function runSoakTest() {
  group('Soak Test - Extended Duration', function() {
    const user = getRandomUser();
    const headers = createAuthHeaders(user);
    
    // Typical user workflow over extended period
    const workflows = [
      () => {
        // Check dashboard
        http.get(`${API_BASE}/dashboard`, { headers });
        sleep(2);
        
        // View customers
        http.get(`${API_BASE}/customers`, { headers });
        sleep(1);
        
        // View jobs
        http.get(`${API_BASE}/jobs`, { headers });
        sleep(1);
      },
      
      () => {
        // Check invoices
        http.get(`${API_BASE}/invoices`, { headers });
        sleep(2);
        
        // Check payments
        http.get(`${API_BASE}/payments`, { headers });
        sleep(1);
      }
    ];
    
    const workflow = workflows[Math.floor(Math.random() * workflows.length)];
    workflow();
  });
}

// Report generation
export function handleSummary(data) {
  return {
    'load-test-results.html': htmlReport(data),
    'load-test-results.json': JSON.stringify(data, null, 2),
  };
}

// Setup function - runs once per VU
export function setup() {
  console.log('Setting up load test environment...');
  
  // Pre-authenticate test users
  TEST_USERS.forEach(user => {
    const response = http.post(`${API_BASE}/auth/login`, JSON.stringify({
      email: user.email,
      password: user.password,
    }), {
      headers: { 'Content-Type': 'application/json' },
    });
    
    if (response.status === 200) {
      const token = JSON.parse(response.body).access_token;
      authTokens.set(`${user.email}:${user.role}`, token);
      console.log(`Pre-authenticated user: ${user.email}`);
    } else {
      console.log(`Failed to pre-authenticate user: ${user.email}`);
    }
  });
  
  return { timestamp: Date.now() };
}

// Teardown function - runs once after all VUs finish
export function teardown(data) {
  console.log(`Load test completed. Duration: ${(Date.now() - data.timestamp) / 1000}s`);
  
  // Clear authentication cache
  authTokens.clear();
}