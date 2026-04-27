/**
 * k6 smoke / light load against public API health endpoints.
 *
 * Prerequisites: brew install k6 (or https://k6.io/docs/getting-started/installation/)
 *
 * Usage:
 *   BASE_URL=https://api.develop.smartcash.danielrive.site k6 run tests/load/smoke.js
 *
 * After JWT CSI fix and deploy, validate pods Ready then run quick-test.sh, then this script.
 */

import http from "k6/http";
import { check, sleep } from "k6";

export const options = {
  stages: [
    { duration: "30s", target: 5 },
    { duration: "1m", target: 10 },
    { duration: "30s", target: 0 },
  ],
  thresholds: {
    http_req_failed: ["rate<0.05"],
    http_req_duration: ["p(95)<3000"],
  },
};

const BASE_URL = (__ENV.BASE_URL || "https://api.develop.smartcash.danielrive.site").replace(/\/$/, "");

// Paths match Gin routes (see app/*/main.go); ingress uses host api.<env>.smartcash.danielrive.site
const paths = [
  "/user/health",
  "/expenses/health",
  "/bank/health",
  "/payment/health",
];

export default function () {
  const path = paths[Math.floor(Math.random() * paths.length)];
  const res = http.get(`${BASE_URL}${path}`, { tags: { name: path } });
  check(res, {
    "status is 200": (r) => r.status === 200,
  });
  sleep(0.3 + Math.random() * 0.5);
}
