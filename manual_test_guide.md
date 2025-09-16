# 🧪 Guía de Pruebas Manuales - TrustlessWork Indexer

## 🚀 Preparación

1. **Inicia el servidor:**
   ```bash
   docker compose up -d postgres
   go build -o indexer cmd/indexer/main.go
   ./indexer
   ```

2. **Verifica que esté funcionando:**
   ```bash
   curl http://localhost:8080/health || echo "API is running on port 8080"
   ```

## 📋 Casos de Prueba

### 🎯 **Test 1: Single Release - Básico**
```bash
# Crear escrow
curl -X POST http://localhost:8080/escrows/single \
  -H 'Content-Type: application/json' \
  -d @tests/single_basic.json

# Obtener escrow
curl -X GET http://localhost:8080/escrows/SINGLE001BASICTEST123456789ABCDEF | jq '.'

# Verificar campos importantes
curl -X GET http://localhost:8080/escrows/SINGLE001BASICTEST123456789ABCDEF | jq '.amount.raw, .platformFee, .milestones[0].description'
```

### 🎯 **Test 2: Single Release - Cantidades Altas**
```bash
# Crear escrow con cantidades muy altas
curl -X POST http://localhost:8080/escrows/single \
  -H 'Content-Type: application/json' \
  -d @tests/single_high_amount.json

# Verificar que maneja números grandes correctamente
curl -X GET http://localhost:8080/escrows/SINGLE002HIGHAMOUNT123456789ABCDEF | jq '.amount.raw'

# Debe mostrar: "999999999999999999"
```

### 🎯 **Test 3: Single Release - Valores Mínimos**
```bash
# Crear escrow con valores mínimos
curl -X POST http://localhost:8080/escrows/single \
  -H 'Content-Type: application/json' \
  -d @tests/single_min_values.json

# Verificar valores mínimos
curl -X GET http://localhost:8080/escrows/SINGLE003MINVALUES123456789ABCDEF | jq '.amount.raw, .platformFee'

# Debe mostrar: "1", 0
```

### 🎯 **Test 4: Multi Release - Básico**
```bash
# Crear escrow multi-milestone
curl -X POST http://localhost:8080/escrows/multi \
  -H 'Content-Type: application/json' \
  -d @tests/multi_basic.json

# Obtener y verificar milestones
curl -X GET http://localhost:8080/escrows/MULTI001BASICTEST123456789ABCDEF | jq '.milestones | length'

# Debe mostrar: 3

# Ver detalles de milestones
curl -X GET http://localhost:8080/escrows/MULTI001BASICTEST123456789ABCDEF | jq '.milestones[] | {description, amount}'
```

### 🎯 **Test 5: Multi Release - Complejo**
```bash
# Crear escrow complejo con 7 milestones
curl -X POST http://localhost:8080/escrows/multi \
  -H 'Content-Type: application/json' \
  -d @tests/multi_complex.json

# Verificar número de milestones
curl -X GET http://localhost:8080/escrows/MULTI002COMPLEX123456789ABCDEF | jq '.milestones | length'

# Debe mostrar: 7

# Ver total calculado
curl -X GET http://localhost:8080/escrows/MULTI002COMPLEX123456789ABCDEF | jq '.totalAmount.raw'

# Debe mostrar: "1550000000" (suma de todos los milestones)
```

### 🎯 **Test 6: Multi Release - Mínimo**
```bash
# Crear escrow con configuración mínima
curl -X POST http://localhost:8080/escrows/multi \
  -H 'Content-Type: application/json' \
  -d @tests/multi_minimal.json

# Verificar configuración mínima
curl -X GET http://localhost:8080/escrows/MULTI003MINIMAL123456789ABCDEF | jq '.milestones[].amount, .totalAmount.raw'

# Debe mostrar: 50, 50, "100"
```

## 🔄 **Test 7: Operaciones CRUD Completas**

### Crear, Leer, Actualizar, Eliminar:
```bash
# 1. CREAR
curl -X POST http://localhost:8080/escrows/single \
  -H 'Content-Type: application/json' \
  -d @tests/single_basic.json

# 2. LEER
curl -X GET http://localhost:8080/escrows/SINGLE001BASICTEST123456789ABCDEF

# 3. ACTUALIZAR (mismo POST con datos modificados)
# Modifica el archivo JSON y ejecuta de nuevo:
curl -X POST http://localhost:8080/escrows/single \
  -H 'Content-Type: application/json' \
  -d @tests/single_basic.json

# 4. ELIMINAR  
curl -X DELETE http://localhost:8080/escrows/SINGLE001BASICTEST123456789ABCDEF
```

## 💰 **Test 8: Indexación de Depósitos**

```bash
# Crear un escrow primero
curl -X POST http://localhost:8080/escrows/multi -H 'Content-Type: application/json' -d @tests/multi_basic.json

# Indexar depósitos (usa datos mock)
curl -X POST http://localhost:8080/index/funder-deposits/MULTI001BASICTEST123456789ABCDEF

# Verificar respuesta de depósitos
curl -X POST http://localhost:8080/index/funder-deposits/MULTI001BASICTEST123456789ABCDEF | jq '.deposits | length'

# Debe mostrar: 2 (datos mock)
```

## 🧪 **Test 9: Casos Extremos y Validaciones**

### Test de Validación de Platform Fee:
```bash
# Crear JSON con platform fee inválido (>10000)
cat > tests/invalid_fee.json << 'EOF'
{
  "contractId": "INVALID001TEST123456789ABCDEF",
  "contractBaseId": "BASE999EXAMPLE123456789ABCDEF",
  "signer": "GASIGNER999999999999999999999999999999999999999999999999",
  "engagementId": "engagement_invalid_001",
  "title": "Invalid Fee Test",
  "description": "Testing platform fee validation",
  "roles": {
    "approver": "GAAPPROVER999999999999999999999999999999999999999999999",
    "serviceProvider": "GASERVICE999999999999999999999999999999999999999999999",
    "platformAddress": "GAPLATFORM999999999999999999999999999999999999999999999",
    "releaseSigner": "GARELEASE999999999999999999999999999999999999999999999",
    "disputeResolver": "GADISPUTE999999999999999999999999999999999999999999999",
    "receiver": "GARECEIVER999999999999999999999999999999999999999999999"
  },
  "amountRaw": 1000000,
  "balanceRaw": 1000000,
  "platformFee": 15000,
  "milestones": [{"description": "Test milestone"}],
  "trustline": {
    "address": "CTEST3STELLARORG123456789ABCDEF123456789ABCDEF123456789",
    "decimals": 6,
    "name": "Test Token"
  },
  "receiverMemo": 9001
}
EOF

# Probar validación (debería fallar)
curl -X POST http://localhost:8080/escrows/single \
  -H 'Content-Type: application/json' \
  -d @tests/invalid_fee.json
```

### Test de Contract ID Duplicado:
```bash
# Crear escrow
curl -X POST http://localhost:8080/escrows/single -H 'Content-Type: application/json' -d @tests/single_basic.json

# Intentar crear el mismo ID de nuevo (debería actualizar, no fallar)
curl -X POST http://localhost:8080/escrows/single -H 'Content-Type: application/json' -d @tests/single_basic.json
```

## 🔍 **Test 10: Verificación de Base de Datos**

```bash
# Verificar datos en la base de datos
docker exec trustlesswork-postgres psql -U indexer -d indexer -c "
SELECT contract_id, platform_fee, signer 
FROM single_release_escrow 
ORDER BY created_at DESC 
LIMIT 3;"

docker exec trustlesswork-postgres psql -U indexer -d indexer -c "
SELECT contract_id, platform_fee, signer 
FROM multi_release_escrow 
ORDER BY created_at DESC 
LIMIT 3;"

# Ver depósitos
docker exec trustlesswork-postgres psql -U indexer -d indexer -c "
SELECT contract_id, depositor, amount_raw 
FROM escrow_funder_deposits 
ORDER BY occurred_at DESC 
LIMIT 5;"
```

## ✅ **Checklist de Validación**

Después de cada test, verifica:

- [ ] **Status Code 200** para operaciones exitosas
- [ ] **JSON válido** en las respuestas
- [ ] **Datos correctos** en los campos principales
- [ ] **Persistencia** en la base de datos
- [ ] **Manejo de errores** para casos inválidos

## 🎯 **Casos de Prueba Avanzados**

### Test Performance con Milestones Múltiples:
```bash
# Crear JSON con 20 milestones para probar performance
cat > tests/multi_many_milestones.json << 'EOF'
{
  "contractId": "MULTI999MANYTEST123456789ABCDEF",
  "contractBaseId": "BASE999EXAMPLE123456789ABCDEF",
  "signer": "GASIGNER999999999999999999999999999999999999999999999999",
  "engagementId": "engagement_many_999",
  "title": "Many Milestones Test",
  "description": "Testing with many milestones",
  "roles": {
    "approver": "GAAPPROVER999999999999999999999999999999999999999999999",
    "serviceProvider": "GASERVICE999999999999999999999999999999999999999999999",
    "platformAddress": "GAPLATFORM999999999999999999999999999999999999999999999",
    "releaseSigner": "GARELEASE999999999999999999999999999999999999999999999",
    "disputeResolver": "GADISPUTE999999999999999999999999999999999999999999999",
    "receiver": "GARECEIVER999999999999999999999999999999999999999999999"
  },
  "platformFee": 100,
  "milestones": [
    {"description": "Milestone 1", "amount": 50000000},
    {"description": "Milestone 2", "amount": 50000000},
    {"description": "Milestone 3", "amount": 50000000},
    {"description": "Milestone 4", "amount": 50000000},
    {"description": "Milestone 5", "amount": 50000000},
    {"description": "Milestone 6", "amount": 50000000},
    {"description": "Milestone 7", "amount": 50000000},
    {"description": "Milestone 8", "amount": 50000000},
    {"description": "Milestone 9", "amount": 50000000},
    {"description": "Milestone 10", "amount": 50000000}
  ],
  "trustline": {
    "address": "CTEST9STELLARORG123456789ABCDEF123456789ABCDEF123456789",
    "decimals": 6,
    "name": "Test Many"
  },
  "receiverMemo": 9999
}
EOF

# Crear y medir tiempo
time curl -X POST http://localhost:8080/escrows/multi \
  -H 'Content-Type: application/json' \
  -d @tests/multi_many_milestones.json
```

## 📊 **Monitoreo Durante Pruebas**

Mientras ejecutas las pruebas, puedes monitorear:

```bash
# Ver logs del servidor
tail -f indexer.log

# Monitorear conexiones de base de datos
docker exec trustlesswork-postgres psql -U indexer -d indexer -c "SELECT count(*) FROM pg_stat_activity WHERE datname='indexer';"

# Ver uso de memoria
ps aux | grep indexer
```

---

## 🎯 **Resultado Esperado**

Todas las pruebas deberían:
- ✅ **Crear escrows correctamente**
- ✅ **Devolver datos completos en GET**
- ✅ **Manejar diferentes tipos de datos**
- ✅ **Persistir en la base de datos**
- ✅ **Eliminar correctamente**
- ✅ **Indexar depósitos**

¡Con esta guía puedes hacer pruebas exhaustivas de tu API!