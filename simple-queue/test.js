import http from 'k6/http';

export const options = {
    scenarios: {
        enqueue_test: {
            executor: 'constant-arrival-rate',
            rate: 100, // 100 RPS
            timeUnit: '1s',
            duration: '30s',
            preAllocatedVUs: 50,
            maxVUs: 200,
        },
    },
};

export default function () {
    http.post('http://localhost:8080/queue', JSON.stringify({
        ID: Math.floor(Math.random() * 100000),
        GroupID: "grupo-1",
        Body: "teste"
    }), {
        headers: { 'Content-Type': 'application/json' },
    });
}