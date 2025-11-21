import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

// Custom metrics
const errorRate = new Rate('errors');

// Test configuration
export const options = {
  stages: [
    { duration: '30s', target: 5 },    // Ramp up to 5 users
    { duration: '1m', target: 5 },     // Stay at 5 users
    { duration: '30s', target: 10 },   // Ramp up to 10 users
    { duration: '1m', target: 10 },    // Stay at 10 users
    { duration: '30s', target: 0 },     // Ramp down
  ],
  thresholds: {
    http_req_duration: ['p(95)<1000', 'p(99)<2000'], // Payment flow is slower
    http_req_failed: ['rate<0.02'],                  // Error rate < 2%
    errors: ['rate<0.02'],
  },
};

// Service URLs
const USER_SERVICE = __ENV.USER_SERVICE_URL || 'http://localhost:8181';
const EXPENSES_SERVICE = __ENV.EXPENSES_SERVICE_URL || 'http://localhost:8282';
const PAYMENT_SERVICE = __ENV.PAYMENT_SERVICE_URL || 'http://localhost:8989';

// Pre-authenticated user (should exist in test environment)
const TEST_USER = {
  username: __ENV.TEST_USERNAME || 'testuser',
  password: __ENV.TEST_PASSWORD || 'testpass123',
};

export default function () {
  // Step 1: Login
  const loginRes = http.post(`${USER_SERVICE}/user/login`, JSON.stringify({
    username: TEST_USER.username,
    password: TEST_USER.password,
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
    return;
  }

  const token = JSON.parse(loginRes.body).token;
  const userId = JSON.parse(loginRes.body).userId;
  sleep(1);

  // Step 2: Create expense
  const expenseRes = http.post(`${EXPENSES_SERVICE}/expenses`, JSON.stringify({
    name: `Test Expense ${Date.now()}`,
    description: 'Load test expense',
    amount: Math.random() * 100 + 10, // Random amount between 10-110
    category: 'test',
    tags: ['load-test'],
  }), {
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`,
    },
    tags: { name: 'CreateExpense' },
  });

  const expenseSuccess = check(expenseRes, {
    'create expense status is 201': (r) => r.status === 201,
    'create expense returns expenseId': (r) => {
      if (r.status === 201) {
        const body = JSON.parse(r.body);
        return body.expenseId !== undefined;
      }
      return false;
    },
  });

  if (!expenseSuccess) {
    errorRate.add(1);
    return;
  }

  const expenseId = JSON.parse(expenseRes.body).expenseId;
  sleep(1);

  // Step 3: Process payment
  const paymentRes = http.post(`${PAYMENT_SERVICE}/payment`, JSON.stringify({
    expenseId: expenseId,
    userId: userId,
  }), {
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`,
    },
    tags: { name: 'ProcessPayment' },
  });

  const paymentSuccess = check(paymentRes, {
    'process payment status is 201 or 200': (r) => r.status === 201 || r.status === 200,
    'process payment returns transactionId': (r) => {
      if (r.status === 201 || r.status === 200) {
        const body = JSON.parse(r.body);
        return body.transactionId !== undefined;
      }
      return false;
    },
  });

  if (!paymentSuccess) {
    errorRate.add(1);
    return;
  }

  const transactionId = JSON.parse(paymentRes.body).transactionId;
  sleep(1);

  // Step 4: Get transaction status
  const transactionRes = http.get(`${PAYMENT_SERVICE}/payment/${transactionId}`, {
    headers: {
      'Authorization': `Bearer ${token}`,
    },
    tags: { name: 'GetTransaction' },
  });

  check(transactionRes, {
    'get transaction status is 200': (r) => r.status === 200,
    'get transaction returns status': (r) => {
      if (r.status === 200) {
        const body = JSON.parse(r.body);
        return body.status !== undefined;
      }
      return false;
    },
  });

  sleep(1);
}

