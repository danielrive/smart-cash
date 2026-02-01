import http from 'k6/http';
import { check, sleep } from 'k6';
import { uuidv4 } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';

export const options = {
    insecureSkipTLSVerify: true,
    stages: [
        { duration: '30s', target: 20 }, // Ramp-up to 20 users
        { duration: '2m', target: 20 },  // Stay at 20 users
        { duration: '30s', target: 0 },  // Ramp-down
    ],
    thresholds: {
        http_req_failed: ['rate<0.01'], // Fail if more than 1% errors
        http_req_duration: ['p(95)<500'], // 95% of requests should be < 500ms
    },
};

const BASE_URL = 'https://api.develop.smartcash.danielrive.site';

// --- SETUP PHASE ---
// This runs ONCE before the load test starts to prepare data
export function setup() {
    const users = [];
    const password = "SecurePassword123";

    // Create 20 unique users for our 20 VUs
    for (let i = 0; i < 20; i++) {
        const id = uuidv4().substring(0, 8);
        const userPayload = {
            firstName: "Load",
            lastName: "Tester",
            username: `user${id}`,
            email: `test${id}@example.com`,
            password: password,
        };

        const res = http.post(`${BASE_URL}/user`, JSON.stringify(userPayload), {
            headers: { 'Content-Type': 'application/json' },
        });

        if (res.status === 201) {
            users.push({ username: userPayload.username, password: password });
        }
    }
    return { testUsers: users };
}

// --- VIRTUAL USER LOGIC ---
export default function (data) {
    // Pick a user from the setup data based on the VU ID
    const user = data.testUsers[(__VU - 1) % data.testUsers.length];
    const params = { headers: { 'Content-Type': 'application/json' } };

    // 1. LOGIN (Get the Token)
    const loginRes = http.post(`${BASE_URL}/user/login`, JSON.stringify({
        username: user.username,
        password: user.password
    }), params);

    const token = loginRes.json('token');
    
    check(loginRes, {
        'login successful': (r) => r.status === 200,
        'has token': (t) => token !== undefined,
    });

    // 2. CREATE EXPENSE (Using the Token)
    const authParams = {
        headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${token}`,
        },
    };

    const expensePayload = JSON.stringify({
        name: `Expense ${__ITER}`,
        amount: Math.random() * 100,
        description: "Automated Load Test Entry",
        category: "Testing",
        tags: ["load-test", "k6"]
    });

    const expenseRes = http.post(`${BASE_URL}/expenses`, expensePayload, authParams);

    check(expenseRes, {
        'expense created 201': (r) => r.status === 201,
    });

    // Short sleep to simulate real user behavior
    sleep(1);
}