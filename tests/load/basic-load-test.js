import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

// Custom metrics
const errorRate = new Rate('errors');

// Test configuration
export const options = {
  stages: [
    { duration: '2m', target: 10 }, // Ramp up to 10 users
    { duration: '5m', target: 10 }, // Stay at 10 users
    { duration: '2m', target: 20 }, // Ramp up to 20 users
    { duration: '5m', target: 20 }, // Stay at 20 users
    { duration: '2m', target: 0 },  // Ramp down to 0 users
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'], // 95% of requests must complete below 500ms
    http_req_failed: ['rate<0.1'],    // Error rate must be below 10%
    errors: ['rate<0.1'],             // Custom error rate must be below 10%
  },
};

const BASE_URL = 'http://localhost:8080';

export default function () {
  // Test health endpoint
  let response = http.get(`${BASE_URL}/health`);
  
  let result = check(response, {
    'health check status is 200': (r) => r.status === 200,
    'health check response time < 200ms': (r) => r.timings.duration < 200,
    'health check has correct content': (r) => r.json('status') === 'healthy',
  });
  
  errorRate.add(!result);
  
  sleep(1);
  
  // Test API endpoints (when authentication is implemented)
  // This is a placeholder for future API testing
  /*
  response = http.get(`${BASE_URL}/api/v1/domains`, {
    headers: {
      'Authorization': 'Bearer ' + token,
    },
  });
  
  check(response, {
    'domains API status is 200': (r) => r.status === 200,
    'domains API response time < 500ms': (r) => r.timings.duration < 500,
  });
  */
  
  sleep(1);
}
