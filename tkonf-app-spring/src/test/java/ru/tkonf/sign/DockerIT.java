package ru.tkonf.sign;

import com.fasterxml.jackson.databind.ObjectMapper;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import ru.tkonf.sign.model.*;

import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;

import static org.junit.jupiter.api.Assertions.*;

public class DockerIT {

    private static final String BASE_URL = "http://localhost:8080/api/crypto";
    private HttpClient httpClient;
    private ObjectMapper objectMapper;

    @BeforeEach
    public void setUp() {
        httpClient = HttpClient.newHttpClient();
        objectMapper = new ObjectMapper();
    }

    @Test
    @DisplayName("Тест здоровья сервиса")
    public void testHealthCheck() throws Exception {
        HttpRequest request = HttpRequest.newBuilder()
                .uri(URI.create(BASE_URL + "/health"))
                .GET()
                .build();

        HttpResponse<String> response = httpClient.send(request, HttpResponse.BodyHandlers.ofString());

        assertEquals(200, response.statusCode());
        HealthResponse healthResponse = objectMapper.readValue(response.body(), HealthResponse.class);
        assertEquals("Spring Crypto Service is running", healthResponse.getStatus());
    }

    @Test
    @DisplayName("Тест получения публичного ключа")
    public void testGetPublicKey() throws Exception {
        HttpRequest request = HttpRequest.newBuilder()
                .uri(URI.create(BASE_URL + "/public-key"))
                .GET()
                .build();

        HttpResponse<String> response = httpClient.send(request, HttpResponse.BodyHandlers.ofString());

        assertEquals(200, response.statusCode());
        assertNotNull(response.body());
        assertFalse(response.body().isEmpty());
    }

    @Test
    @DisplayName("Тест подписи сообщения")
    public void testSignMessage() throws Exception {
        SignatureRequest request = new SignatureRequest("Test message", null);

        String requestBody = objectMapper.writeValueAsString(request);

        HttpRequest httpRequest = HttpRequest.newBuilder()
                .uri(URI.create(BASE_URL + "/sign"))
                .header("Content-Type", "application/json")
                .POST(HttpRequest.BodyPublishers.ofString(requestBody))
                .build();

        HttpResponse<String> httpResponse = httpClient.send(httpRequest, HttpResponse.BodyHandlers.ofString());

        assertEquals(200, httpResponse.statusCode());

        SignedMessage signedMessage = objectMapper.readValue(httpResponse.body(), SignedMessage.class);
        assertEquals("Test message", signedMessage.getOriginalMessage());
        assertNotNull(signedMessage.getSignature());
        assertNotNull(signedMessage.getAlgorithm());
        assertNotNull(signedMessage.getPublicKey());
        assertFalse(signedMessage.getSignature().isEmpty());
    }

    @Test
    @DisplayName("Тест проверки валидной подписи")
    public void testVerifyValidSignature() throws Exception {
        // Сначала получаем подпись
        SignatureRequest signRequest = new SignatureRequest("Test verify message", null);
        String signRequestBody = objectMapper.writeValueAsString(signRequest);

        HttpRequest signHttpRequest = HttpRequest.newBuilder()
                .uri(URI.create(BASE_URL + "/sign"))
                .header("Content-Type", "application/json")
                .POST(HttpRequest.BodyPublishers.ofString(signRequestBody))
                .build();

        HttpResponse<String> signResponse = httpClient.send(signHttpRequest, HttpResponse.BodyHandlers.ofString());
        SignedMessage signedMessage = objectMapper.readValue(signResponse.body(), SignedMessage.class);

        // Проверяем подпись
        VerificationRequest verifyRequest = new VerificationRequest(
                "Test verify message",
                signedMessage.getSignature()
        );

        String verifyRequestBody = objectMapper.writeValueAsString(verifyRequest);

        HttpRequest verifyHttpRequest = HttpRequest.newBuilder()
                .uri(URI.create(BASE_URL + "/verify"))
                .header("Content-Type", "application/json")
                .POST(HttpRequest.BodyPublishers.ofString(verifyRequestBody))
                .build();

        HttpResponse<String> verifyResponse = httpClient.send(verifyHttpRequest, HttpResponse.BodyHandlers.ofString());

        assertEquals(200, verifyResponse.statusCode());
        VerificationResponse verificationResponse = objectMapper.readValue(verifyResponse.body(), VerificationResponse.class);
        assertTrue(verificationResponse.isValid());
        assertEquals("Signature is VALID", verificationResponse.getMessage());
    }

    @Test
    @DisplayName("Тест проверки невалидной подписи")
    public void testVerifyInvalidSignature() throws Exception {
        VerificationRequest verifyRequest = new VerificationRequest(
                "Test message",
                "invalid_signature_here"
        );

        String verifyRequestBody = objectMapper.writeValueAsString(verifyRequest);

        HttpRequest verifyHttpRequest = HttpRequest.newBuilder()
                .uri(URI.create(BASE_URL + "/verify"))
                .header("Content-Type", "application/json")
                .POST(HttpRequest.BodyPublishers.ofString(verifyRequestBody))
                .build();

        HttpResponse<String> verifyResponse = httpClient.send(verifyHttpRequest, HttpResponse.BodyHandlers.ofString());

        assertEquals(400, verifyResponse.statusCode());
        ErrorResponse verificationResponse = objectMapper.readValue(verifyResponse.body(), ErrorResponse.class);
        assertEquals("Error verifying signature: Illegal base64 character 5f", verificationResponse.getError());
    }

    @Test
    @DisplayName("Тест шифрования и дешифрования сообщения")
    public void testEncryptDecryptRoundTrip() throws Exception {
        String originalMessage = "Secret message for encryption";

        // Шифруем сообщение
        SignatureRequest encryptRequest = new SignatureRequest(originalMessage, null);
        String encryptRequestBody = objectMapper.writeValueAsString(encryptRequest);

        HttpRequest encryptHttpRequest = HttpRequest.newBuilder()
                .uri(URI.create(BASE_URL + "/encrypt"))
                .header("Content-Type", "application/json")
                .POST(HttpRequest.BodyPublishers.ofString(encryptRequestBody))
                .build();

        HttpResponse<String> encryptResponse = httpClient.send(encryptHttpRequest, HttpResponse.BodyHandlers.ofString());
        assertEquals(200, encryptResponse.statusCode());

        CryptoResponse encryptCryptoResponse = objectMapper.readValue(encryptResponse.body(), CryptoResponse.class);
        String encryptedMessage = encryptCryptoResponse.getResult();
        assertNotNull(encryptedMessage);
        assertNotEquals(originalMessage, encryptedMessage);

        // Дешифруем сообщение
        SignatureRequest decryptRequest = new SignatureRequest(encryptedMessage, null);
        String decryptRequestBody = objectMapper.writeValueAsString(decryptRequest);

        HttpRequest decryptHttpRequest = HttpRequest.newBuilder()
                .uri(URI.create(BASE_URL + "/decrypt"))
                .header("Content-Type", "application/json")
                .POST(HttpRequest.BodyPublishers.ofString(decryptRequestBody))
                .build();

        HttpResponse<String> decryptResponse = httpClient.send(decryptHttpRequest, HttpResponse.BodyHandlers.ofString());
        assertEquals(200, decryptResponse.statusCode());

        CryptoResponse decryptCryptoResponse = objectMapper.readValue(decryptResponse.body(), CryptoResponse.class);
        String decryptedMessage = decryptCryptoResponse.getResult();
        assertEquals(originalMessage, decryptedMessage);
    }

    @Test
    @DisplayName("Тест ошибки при неверных данных")
    public void testErrorHandling() throws Exception {
        String invalidJson = "{ invalid json }";

        HttpRequest request = HttpRequest.newBuilder()
                .uri(URI.create(BASE_URL + "/sign"))
                .header("Content-Type", "application/json")
                .POST(HttpRequest.BodyPublishers.ofString(invalidJson))
                .build();

        HttpResponse<String> response = httpClient.send(request, HttpResponse.BodyHandlers.ofString());

        assertEquals(400, response.statusCode());

        // Проверяем что тело ответа не пустое и содержит информацию об ошибке
        assertNotNull(response.body());
        assertFalse(response.body().isEmpty());


        // Проверяем что в ответе есть упоминание об ошибке (может быть в разных форматах)
        assertTrue(response.body().toLowerCase().contains("error") ||
                response.body().toLowerCase().contains("unable") ||
                response.body().toLowerCase().contains("deserialize"));
    }

    @Test
    @DisplayName("Тест подписи с указанием алгоритма")
    public void testSignWithAlgorithm() throws Exception {
        SignatureRequest request = new SignatureRequest("Test message with algorithm", "RSA");

        String requestBody = objectMapper.writeValueAsString(request);

        HttpRequest httpRequest = HttpRequest.newBuilder()
                .uri(URI.create(BASE_URL + "/sign"))
                .header("Content-Type", "application/json")
                .POST(HttpRequest.BodyPublishers.ofString(requestBody))
                .build();

        HttpResponse<String> httpResponse = httpClient.send(httpRequest, HttpResponse.BodyHandlers.ofString());

        assertEquals(200, httpResponse.statusCode());

        SignedMessage signedMessage = objectMapper.readValue(httpResponse.body(), SignedMessage.class);
        assertEquals("Test message with algorithm", signedMessage.getOriginalMessage());
        assertNotNull(signedMessage.getAlgorithm());
    }
}