import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

// Custom metrics
const errorRate = new Rate('errors');

// Test configuration
export const options = {
  stages: [
    { duration: '30s', target: 10 },   // Ramp up to 10 users
    { duration: '1m', target: 10 },    // Stay at 10 users
    { duration: '30s', target: 20 },    // Ramp up to 20 users
    { duration: '1m', target: 20 },     // Stay at 20 users
    { duration: '30s', target: 0 },     // Ramp down
  ],
  thresholds: {
    http_req_duration: ['p(95)<500', 'p(99)<1000'], // 95% of requests < 500ms, 99% < 1s
    http_req_failed: ['rate<0.01'],                 // Error rate < 1%
    errors: ['rate<0.01'],
  },
};

// Service URLs
const USER_SERVICE = __ENV.USER_SERVICE_URL || 'http://localhost:8181';

// Test data
const testUsers = [
  { username: 'testuser1', email: 'test1@example.com', password: 'testpass123' },
  { username: 'testuser2', email: 'test2@example.com', password: 'testpass123' },
  { username: 'testuser3', email: 'test3@example.com', password: 'testpass123' },
];

export default function () {
  // Select random test user
  const user = testUsers[Math.floor(Math.random() * testUsers.length)];
  const timestamp = Date.now();
  const uniqueUser = {
    username: `${user.username}_${timestamp}`,
    email: `${timestamp}_${user.email}`,
    password: user.password,
  };

  // Step 1: Register user
  const registerRes = http.post(`${USER_SERVICE}/user`, JSON.stringify({
    firstName: 'Test',
    lastName: 'User',
    username: uniqueUser.username,
    email: uniqueUser.email,
    password: uniqueUser.password,
  }), {
    headers: { 'Content-Type': 'application/json' },
    tags: { name: 'RegisterUser' },
  });

  const registerSuccess = check(registerRes, {
    'register status is 201': (r) => r.status === 201,
    'register has userId': (r) => {
      if (r.status === 201) {
        const body = JSON.parse(r.body);
        return body.userId !== undefined;
      }
      return false;
    },
  });

  if (!registerSuccess) {
    errorRate.add(1);
    sleep(1);
    return;
  }

  const userId = JSON.parse(registerRes.body).userId;
  sleep(1);

  // Step 2: Login
  const loginRes = http.post(`${USER_SERVICE}/user/login`, JSON.stringify({
    username: uniqueUser.username,
    password: uniqueUser.password,
  }), {
    headers: { 'Content-Type': 'application/json' },
    tags: { name: 'Login' },
  });

  const loginSuccess = check(loginRes, {
    'login status is 200': (r) => r.status === 200,
    'login returns token': (r) => {
      if (r.status === 200) {
        const body = JSON.parse(r.body);
        return body.token !== undefined;
      }
      return false;
    },
  });

  if (!loginSuccess) {
    errorRate.add(1);
    sleep(1);
    return;
  }

  const token = JSON.parse(loginRes.body).token;
  sleep(1);

  // Step 3: Get user by ID (authenticated)
  const getUserRes = http.get(`${USER_SERVICE}/user/${userId}`, {
    headers: {
      'Authorization': `Bearer ${token}`,
    },
    tags: { name: 'GetUserById' },
  });

  check(getUserRes, {
    'get user status is 200': (r) => r.status === 200,
    'get user returns correct userId': (r) => {
      if (r.status === 200) {
        const body = JSON.parse(r.body);
        return body.userId === userId;
      }
      return false;
    },
  });

  sleep(1);

  // Step 4: Get user by email
  const getUserByEmailRes = http.get(`${USER_SERVICE}/user?email=${encodeURIComponent(uniqueUser.email)}`, {
    tags: { name: 'GetUserByEmail' },
  });

  check(getUserByEmailRes, {
    'get user by email status is 200': (r) => r.status === 200,
  });

  sleep(1);
}

// Summary will be printed automatically by k6

