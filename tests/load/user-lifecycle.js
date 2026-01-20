import http from 'k6/http';
import { check, sleep } from 'k6';
import { uuidv4 } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';

export const options = {
  insecureSkipTLSVerify: true,
  stages: [
    { duration: '1m', target: 10 }, // Ramp-up
    { duration: '3m', target: 50 }, // Heavy load: 50 concurrent users
    { duration: '1m', target: 0 },  // Ramp-down
  ],
  thresholds: {
    // We expect < 1% of requests to fail (non-2xx/409)
    http_req_failed: ['rate<0.01'], 
    // 95% of transactions should be under 400ms (Transactions are slower than Puts!)
    http_req_duration: ['p(95)<400'], 
  },
};

const BASE_URL = 'https://api.develop.smartcash.danielrive.site';

export default function () {
  // 1. Generate unique identity for this specific iteration
  const randomId = uuidv4().substring(0, 8);
  const username = `user${__VU}x${__ITER}`;
  const email = `${username}@example.com`;
  const password = "SecurePassword123";

  const params = {
    headers: { 'Content-Type': 'application/json' },
  };

  // --- STEP 1: CREATE USER ---
  // This triggers your new Go Transaction (3 writes in one)
  const registerPayload = JSON.stringify({
    firstName: "Load",
    lastName: "Test",
    username: username,
    email: email,
    password: password,
  });

  const resRegister = http.post(`${BASE_URL}/user`, registerPayload, params);

  check(resRegister, {
    'is status 201': (r) => r.status === 201,
  });

  // Simulate "thinking time" between registration and login
  sleep(3);

  // --- STEP 2: LOGIN ---
  // This tests your GSI lookup and password hashing performance
  const loginPayload = JSON.stringify({
    username: username,
    password: password,
  });

  const resLogin = http.post(`${BASE_URL}/user/login`, loginPayload, params);

  check(resLogin, {
    'is status 200': (r) => r.status === 200,
    'has token': (r) => r.json('token') !== undefined,
  });

  sleep(2);
}