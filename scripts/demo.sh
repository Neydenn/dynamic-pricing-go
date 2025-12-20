set -euo pipefail

base_catalog="http://localhost:8081"
base_order="http://localhost:8082"
base_pricing="http://localhost:8083"

wait_http() {
  url="$1"
  for i in $(seq 1 60); do
    if curl -fsS "$url" >/dev/null 2>&1; then
      return 0
    fi
    sleep 1
  done
  echo "service not ready: $url"
  exit 1
}

wait_http "$base_catalog/health"
wait_http "$base_order/health"
wait_http "$base_pricing/health"

py_get() {
  python3 - <<'PY'
import json,sys
data=json.load(sys.stdin)
print(data[sys.argv[1]])
PY
}

echo "create product"
product_json=$(curl -fsS -X POST "$base_catalog/products" -H "Content-Type: application/json" -d '{"name":"Notebook","base_price":100.0,"stock":10}')
product_id=$(echo "$product_json" | python3 -c "import json,sys;print(json.load(sys.stdin)['id'])")
echo "product_id=$product_id"

echo "create user"
user_json=$(curl -fsS -X POST "$base_order/users" -H "Content-Type: application/json" -d '{"email":"test@example.com"}')
user_id=$(echo "$user_json" | python3 -c "import json,sys;print(json.load(sys.stdin)['id'])")
echo "user_id=$user_id"

echo "initial price (may be 404 until first order)"
curl -sS "$base_pricing/prices/$product_id" || true
echo

echo "place 5 orders"
for i in $(seq 1 5); do
  curl -fsS -X POST "$base_order/orders" -H "Content-Type: application/json" -d "{"user_id":"$user_id","product_id":"$product_id","qty":1}" >/dev/null
done

sleep 2

echo "price after demand"
curl -fsS "$base_pricing/prices/$product_id"
echo

echo "set low stock"
curl -fsS -X PATCH "$base_catalog/products/$product_id/stock" -H "Content-Type: application/json" -d '{"stock":3}' >/dev/null

sleep 2

echo "place 3 more orders"
for i in $(seq 1 3); do
  curl -fsS -X POST "$base_order/orders" -H "Content-Type: application/json" -d "{"user_id":"$user_id","product_id":"$product_id","qty":1}" >/dev/null
done

sleep 2

echo "price after demand + low stock"
curl -fsS "$base_pricing/prices/$product_id"
echo
