import http from 'k6/http';
import { check, sleep } from 'k6';
import encoding from 'k6/encoding';

export const options = {
    insecureSkipTLSVerify: true,
    stages: [
        { duration: '30s', target: 20 }, 
        { duration: '2m', target: 20 },  
        { duration: '30s', target: 0 },  
    ],
    thresholds: {
        http_req_failed: ['rate<0.01'],   
        http_req_duration: ['p(95)<500'], 
    },
};

const BASE_URL = 'https://api.develop.smartcash.danielrive.site';

/**
 * Decodes JWT using k6 native encoding
 */
function getUserIdFromToken(token) {
    if (!token) return null;
    try {
        const base64Url = token.split('.')[1];
        let base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/');
        
        // Add padding if necessary for the decoder
        while (base64.length % 4 !== 0) {
            base64 += '=';
        }

        const decoded = encoding.b64decode(base64, 'std', 's');
        const payload = JSON.parse(decoded);
        
        // Check if your JWT claim is 'user_id' or 'userId'
        return payload.user_id || payload.userId; 
    } catch (e) {
        console.error("Failed to decode JWT: " + e.message);
        return null;
    }
}

export default function () {
    const headers = { 'Content-Type': 'application/json' };

    // --- STEP 1: LOGIN ---
    const loginPayload = JSON.stringify({
        username: "johndoe",
        password: "SecurePassword123"
    });

    const loginRes = http.post(`${BASE_URL}/user/login`, loginPayload, { headers });

    // Check if login was successful
    if (!check(loginRes, { 'Login successful': (r) => r.status === 200 })) {
        console.error(`Login failed: ${loginRes.status} - ${loginRes.body}`);
        return; // Stop this VU iteration
    }

    const token = loginRes.json('token');
    const userId = getUserIdFromToken(token);

    if (!userId) {
        console.error("Could not extract userId from token claims");
        return;
    }

    // --- STEP 2: CREATE EXPENSE ---
    const expensePayload = JSON.stringify({
        name: `LoadTest-${__VU}-${__ITER}`,
        amount: 45.99,
        description: "Automated test expense",
        category: "Testing",
        date: new Date().toISOString().split('T')[0]
    });

    const expenseRes = http.post(`${BASE_URL}/expenses`, expensePayload, {
        headers: { ...headers, 'Authorization': `Bearer ${token}` }
    });

    const expenseId = expenseRes.json('expenseId');

    if (!check(expenseRes, { 'Expense created': (r) => r.status === 201 })) {
        console.error(`Expense creation failed: ${expenseRes.body}`);
        return;
    }

    // --- STEP 3: PAY EXPENSE ---
    console.log(`DEBUG: Sending Payment - User: ${userId}, Expense: ${expenseId}`);
    const paymentPayload = JSON.stringify({
        expenseId: expenseId,
        amount: 45.99
    });

    const paymentRes = http.post(`${BASE_URL}/payment`, paymentPayload, {
        headers: { ...headers, 'Authorization': `Bearer ${token}` }
    });

    check(paymentRes, {
        'Payment accepted': (r) => r.status === 200 || r.status === 201 || r.status === 202,
    });

    sleep(1);
}