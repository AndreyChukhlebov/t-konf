package ru.tkonf.sign.controller;

import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;
import ru.tkonf.sign.service.SignatureService;
import ru.tkonf.sign.model.*;

@RestController
@RequestMapping("/api/crypto")
public class SpringCryptoController {

    private final SignatureService signatureService;

    // Конструктор с инъекцией зависимости
    public SpringCryptoController(SignatureService signatureService) {
        this.signatureService = signatureService;
    }

    @PostMapping("/sign")
    public ResponseEntity<?> signMessage(@RequestBody SignatureRequest request) {
        try {

            String signature = signatureService.signMessage(request.getMessage());
            SignedMessage signedMessage = new SignedMessage(
                    request.getMessage(),
                    signature,
                    signatureService.getAlgorithm(),
                    signatureService.getPublicKeyBase64()
            );
            return ResponseEntity.ok(signedMessage);
        } catch (Exception e) {
            e.printStackTrace();
            return ResponseEntity.badRequest()
                    .body(new ErrorResponse("Error signing message: " + e.getMessage()));
        }
    }

    @PostMapping("/verify")
    public ResponseEntity<?> verifySignature(@RequestBody VerificationRequest request) {
        try {

            boolean isValid = signatureService.verifySignature(request.getMessage(), request.getSignature());
            String message = isValid ? "Signature is VALID" : "Signature is INVALID";
            return ResponseEntity.ok(new VerificationResponse(isValid, message));
        } catch (Exception e) {
            e.printStackTrace();
            return ResponseEntity.badRequest()
                    .body(new ErrorResponse("Error verifying signature: " + e.getMessage()));
        }
    }

    @PostMapping("/encrypt")
    public ResponseEntity<?> encryptMessage(@RequestBody SignatureRequest request) {
        try {

            String encrypted = signatureService.encrypt(request.getMessage());
            return ResponseEntity.ok(new CryptoResponse(encrypted));
        } catch (Exception e) {
            e.printStackTrace();
            return ResponseEntity.badRequest()
                    .body(new ErrorResponse("Error encrypting message: " + e.getMessage()));
        }
    }

    @PostMapping("/decrypt")
    public ResponseEntity<?> decryptMessage(@RequestBody SignatureRequest request) {
        try {

            String decrypted = signatureService.decrypt(request.getMessage());
            return ResponseEntity.ok(new CryptoResponse(decrypted));
        } catch (Exception e) {
            e.printStackTrace();
            return ResponseEntity.badRequest()
                    .body(new ErrorResponse("Error decrypting message: " + e.getMessage()));
        }
    }

    @GetMapping("/public-key")
    public ResponseEntity<String> getPublicKey() {
        try {
            String publicKey = signatureService.getPublicKeyBase64();
            return ResponseEntity.ok(publicKey);
        } catch (Exception e) {
            return ResponseEntity.internalServerError()
                    .body("Error getting public key: " + e.getMessage());
        }
    }

    @GetMapping("/health")
    public ResponseEntity<HealthResponse> health() {
        return ResponseEntity.ok(new HealthResponse("Spring Crypto Service is running"));
    }
}