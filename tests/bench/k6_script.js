import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

const errorRate = new Rate('errors');
const canaryPassRate = new Rate('canary_pass');

export const options = {
  stages: [
    { duration: '30s', target: 2 },  // 2 workers
    { duration: '2m', target: 2 },   // Stay at 2 workers
    { duration: '30s', target: 0 },  // Ramp down
  ],
  thresholds: {
    'http_req_duration': ['p(95)<2500'], // p95 < 2.5s
    'errors': ['rate<0.01'],             // Error rate < 1%
    'canary_pass': ['rate>0.99'],        // Canary pass > 99%
  },
};

export default function () {
  const payload = JSON.stringify({
    prompt: generatePrompt(),
    model: 'llama2',
    token: `bench-${__VU}-${__ITER}`,
  });

  const params = {
    headers: { 'Content-Type': 'application/json' },
    timeout: '30s',
  };

  const res = http.post('http://localhost:8080/generate', payload, params);
  
  // Check response
  const success = check(res, {
    'status is 200': (r) => r.status === 200,
    'has completion': (r) => r.json('completion') !== undefined,
    'has receipt': (r) => r.json('receipt') !== undefined,
  });

  errorRate.add(!success);

  // Check canary if present
  if (res.status === 200) {
    const receipt = res.json('receipt');
    if (receipt && receipt.canary && receipt.canary.id) {
      canaryPassRate.add(receipt.canary.passed ? 1 : 0);
    }
  }

  sleep(0.1); // 100ms between requests
}

function generatePrompt() {
  const prompts = [
    'What is the weather today?',
    'Calculate the sum of 15 and 27',
    'Tell me a short joke',
    'What is the capital of Japan?',
    'How do I make coffee?',
    'What is 2 + 2?',
    'Name three colors',
    'What day is it?',
  ];
  
  // Short prompts <= 256 tokens
  return prompts[Math.floor(Math.random() * prompts.length)];
}