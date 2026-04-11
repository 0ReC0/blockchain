#!/bin/bash

echo "=== Sending test transaction ==="
curl -s -X POST "http://localhost:8081/transactions" \
  -H "Content-Type: application/json" \
  -d '{"ID":"tx_test_123","From":"sender1","To":"receiver1","Amount":50,"Fee":0.001,"Timestamp":1775923500,"Signature":"30440220","IsPrivate":false}'

echo ""
echo ""
echo "=== Waiting 12 seconds for PoS block creation ==="
sleep 12

echo ""
echo "=== Blocks after 12s ==="
curl -s http://localhost:8081/blocks 2>&1 | python3 -c "import sys,json; b=json.load(sys.stdin); print(f'Total blocks: {len(b)}'); [print(f'  Block #{x[\"index\"]}: {len(x.get(\"transactions\",[]))} txs') for x in b]"

echo ""
echo "=== Node logs ==="
tail -30 /tmp/node.log 2>&1 | grep -E "(Block|PoS|⛏️|✅)"
