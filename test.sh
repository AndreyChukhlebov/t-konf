# 1. Проверка здоровья (должно работать)
curl  -v -X GET http://0.0.0.0:8090/api/crypto/health

# 2. Получение публичного ключа (должно работать)
curl  -v -X GET http://localhost:8090/api/crypto/public-key

# 3. Подпись сообщения - ВАЖНО: используйте точный формат JSON
curl  -v -X POST http://localhost:8090/api/crypto/sign \
  -H "Content-Type: application/json" \
  -H "Accept: application/json" \
  -d '{"message":"Hello World","algorithm":"RSA"}'

# 4. Если все еще есть проблемы, попробуйте с явным указанием charset
curl  -v -X POST http://localhost:8090/api/crypto/sign \
  -H "Content-Type: application/json; charset=utf-8" \
  -d '{"message":"Test","algorithm":"RSA"}'

# 5. Альтернатива с использованием файла
echo '{"message":"From file","algorithm":"RSA"}' > request.json
curl  -v -X POST http://localhost:8090/api/crypto/sign \
  -H "Content-Type: application/json" \
  -d @request.json