package ru.tkonf.sign;


import io.micronaut.http.HttpResponse;
import io.micronaut.http.HttpStatus;
import io.micronaut.http.MediaType;
import io.micronaut.http.annotation.*;
import jakarta.inject.Inject;
import ru.tkonf.sign.model.*;
import ru.tkonf.sign.service.SignatureService;

@Controller("/api/crypto")
@Produces(MediaType.APPLICATION_JSON)
@Consumes(MediaType.APPLICATION_JSON)
public class MicronautCryptoController {

    @Inject
    SignatureService signatureService;

    @Post("/sign")
    public HttpResponse<?> signMessage(@Body SignatureRequest request) {
        try {
            System.out.println("Received sign request: " + request.getMessage());

            String signature = signatureService.signMessage(request.getMessage());
            SignedMessage signedMessage = new SignedMessage(
                    request.getMessage(),
                    signature,
                    signatureService.getAlgorithm(),
                    signatureService.getPublicKeyBase64()
            );
            return HttpResponse.ok(signedMessage);
        } catch (Exception e) {
            e.printStackTrace();
            return HttpResponse.status(HttpStatus.BAD_REQUEST)
                    .body(new ErrorResponse("Error signing message: " + e.getMessage()));
        }
    }

    @Post("/verify")
    public HttpResponse<?> verifySignature(@Body VerificationRequest request) {
        try {
            System.out.println("Received verify request: " + request.getMessage());

            boolean isValid = signatureService.verifySignature(request.getMessage(), request.getSignature());
            String message = isValid ? "Signature is VALID" : "Signature is INVALID";
            return HttpResponse.ok(new VerificationResponse(isValid, message));
        } catch (Exception e) {
            e.printStackTrace();
            return HttpResponse.status(HttpStatus.BAD_REQUEST)
                    .body(new ErrorResponse("Error verifying signature: " + e.getMessage()));
        }
    }

    @Post("/encrypt")
    public HttpResponse<?> encryptMessage(@Body SignatureRequest request) {
        try {
            System.out.println("Received encrypt request: " + request.getMessage());

            String encrypted = signatureService.encrypt(request.getMessage());
            return HttpResponse.ok(new CryptoResponse(encrypted));
        } catch (Exception e) {
            e.printStackTrace();
            return HttpResponse.status(HttpStatus.BAD_REQUEST)
                    .body(new ErrorResponse("Error encrypting message: " + e.getMessage()));
        }
    }

    @Post("/decrypt")
    public HttpResponse<?> decryptMessage(@Body SignatureRequest request) {
        try {
            System.out.println("Received decrypt request: " + request.getMessage());

            String decrypted = signatureService.decrypt(request.getMessage());
            return HttpResponse.ok(new CryptoResponse(decrypted));
        } catch (Exception e) {
            e.printStackTrace();
            return HttpResponse.status(HttpStatus.BAD_REQUEST)
                    .body(new ErrorResponse("Error decrypting message: " + e.getMessage()));
        }
    }

    @Get("/public-key")
    @Produces(MediaType.TEXT_PLAIN)
    public HttpResponse<String> getPublicKey() {
        try {
            String publicKey = signatureService.getPublicKeyBase64();
            return HttpResponse.ok(publicKey);
        } catch (Exception e) {
            return HttpResponse.status(HttpStatus.INTERNAL_SERVER_ERROR)
                    .body("Error getting public key: " + e.getMessage());
        }
    }

    @Get("/health")
    public HttpResponse<HealthResponse> health() {
        return HttpResponse.ok(new HealthResponse("Micronaut Crypto Service is running"));
    }
}