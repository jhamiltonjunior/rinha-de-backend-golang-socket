wrk.method = "POST"
wrk.body   = '{"correlationId":"704fa1b6-b5e3-4efc-a6c1-31955d3f4aa4","amount": 19.90}'
wrk.headers["Content-Type"] = "application/json"

-- wrk -t2 -c10 -d10s --latency -s /scripts/post.lua http://rinha-api-1:3000/payments

-- wrk -t4 -c100 -d10s --latency -s /scripts/post.lua http://nginx:80/payments
