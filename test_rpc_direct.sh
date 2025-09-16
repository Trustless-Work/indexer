#!/bin/bash

echo "Testing getLatestLedger..."
curl -X POST https://soroban-testnet.stellar.org \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "getLatestLedger"
  }'

echo -e "\n\nTesting getEvents with our contract ID..."
curl -X POST https://soroban-testnet.stellar.org \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 2,
    "method": "getEvents",
    "params": {
      "startLedger": 1000000,
      "endLedger": 1000100,
      "filters": [
        {
          "type": "contract",
          "contractIds": ["CACDYF3CYMJEJTIVFESQYZTN67GO2R5D5IUABTCUG3HXQSRXCSOROBAN"]
        }
      ],
      "pagination": {
        "limit": 10
      }
    }
  }'
echo