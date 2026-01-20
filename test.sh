#!/bin/bash

echo "================================================"
echo "Diagnosing Jaeger Tracing Issues"
echo "================================================"
echo ""

echo "1. Checking if services are running..."
echo "   Gateway:"
curl -s http://localhost:8080/health && echo " ✓ UP" || echo " ✗ DOWN"
echo "   Products:"
curl -s http://localhost:8081/health && echo " ✓ UP" || echo " ✗ DOWN"
echo ""

echo "2. Making test requests to generate traces..."
for i in {1..5}; do
    echo -n "   Request $i: "
    curl -s http://localhost:8080/products > /dev/null && echo "✓" || echo "✗"
    sleep 0.5
done
echo ""

echo "3. Checking Jaeger for services..."
SERVICES=$(curl -s http://localhost:16686/api/services 2>/dev/null | jq -r '.data[]' 2>/dev/null)
echo "   Services found in Jaeger:"
if [ -n "$SERVICES" ]; then
    echo "$SERVICES" | while read service; do
        echo "      - $service"
    done
else
    echo "      ✗ NO SERVICES FOUND!"
fi
echo ""

echo "4. Checking gateway container logs for Jaeger..."
echo "   Last 15 lines from gateway:"
docker logs gateway 2>&1 | tail -15
echo ""

echo "5. Checking products container logs for Jaeger..."
echo "   Last 15 lines from products:"
docker logs products 2>&1 | tail -15
echo ""

echo "6. Checking Jaeger container logs..."
echo "   Looking for span receipts:"
docker logs jaeger 2>&1 | grep -i "span" | tail -10
echo ""

echo "7. Testing Jaeger agent connectivity from containers..."
echo "   Gateway can reach Jaeger agent?"
docker exec gateway sh -c 'nc -zv jaeger 6831 2>&1' || echo "   ✗ Cannot reach Jaeger from gateway"
echo ""
echo "   Products can reach Jaeger agent?"
docker exec products sh -c 'nc -zv jaeger 6831 2>&1' || echo "   ✗ Cannot reach Jaeger from products"
echo ""

echo "8. Checking environment variables in containers..."
echo "   Gateway env:"
docker exec gateway env | grep JAEGER
echo ""
echo "   Products env:"
docker exec products env | grep JAEGER
echo ""

echo "================================================"
echo "Diagnosis Complete"
echo "================================================"
echo ""
echo "If no services appear in Jaeger, the issue is likely:"
echo "1. ✗ Services not sending traces (check logs above)"
echo "2. ✗ Services can't reach Jaeger agent (check connectivity)"
echo "3. ✗ Jaeger client not initialized properly in code"
echo ""