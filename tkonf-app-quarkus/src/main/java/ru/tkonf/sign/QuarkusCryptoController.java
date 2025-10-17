package ru.tkonf.sign;



import jakarta.inject.Inject;
import jakarta.ws.rs.*;
import jakarta.ws.rs.core.MediaType;
import jakarta.ws.rs.core.Response;
import ru.tkonf.sign.model.*;
import ru.tkonf.sign.service.SignatureService;

@Path("/api/crypto")
@Produces(MediaType.APPLICATION_JSON)
@Consumes(MediaType.APPLICATION_JSON)
public class QuarkusCryptoController {

    @Inject
    SignatureService signatureService;

    @POST
    @Path("/sign")
    public Response signMessage(SignatureRequest request) {
        try {

            String signature = signatureService.signMessage(request.getMessage());
            SignedMessage signedMessage = new SignedMessage(
                    request.getMessage(),
                    signature,
                    signatureService.getAlgorithm(),
                    signatureService.getPublicKeyBase64()
            );
            return Response.ok(signedMessage).build();
        } catch (Exception e) {
            e.printStackTrace();
            return Response.status(Response.Status.BAD_REQUEST)
                    .entity(new ErrorResponse("Error signing message: " + e.getMessage()))
                    .build();
        }
    }

    @POST
    @Path("/verify")
    public Response verifySignature(VerificationRequest request) {
        try {

            boolean isValid = signatureService.verifySignature(request.getMessage(), request.getSignature());
            String message = isValid ? "Signature is VALID" : "Signature is INVALID";
            return Response.ok(new VerificationResponse(isValid, message)).build();
        } catch (Exception e) {
            e.printStackTrace();
            return Response.status(Response.Status.BAD_REQUEST)
                    .entity(new ErrorResponse("Error verifying signature: " + e.getMessage()))
                    .build();
        }
    }

    @POST
    @Path("/encrypt")
    public Response encryptMessage(SignatureRequest request) {
        try {

            String encrypted = signatureService.encrypt(request.getMessage());
            return Response.ok(new CryptoResponse(encrypted)).build();
        } catch (Exception e) {
            e.printStackTrace();
            return Response.status(Response.Status.BAD_REQUEST)
                    .entity(new ErrorResponse("Error encrypting message: " + e.getMessage()))
                    .build();
        }
    }

    @POST
    @Path("/decrypt")
    public Response decryptMessage(SignatureRequest request) {
        try {

            String decrypted = signatureService.decrypt(request.getMessage());
            return Response.ok(new CryptoResponse(decrypted)).build();
        } catch (Exception e) {
            e.printStackTrace();
            return Response.status(Response.Status.BAD_REQUEST)
                    .entity(new ErrorResponse("Error decrypting message: " + e.getMessage()))
                    .build();
        }
    }

    @GET
    @Path("/public-key")
    @Produces(MediaType.TEXT_PLAIN)
    public Response getPublicKey() {
        try {
            String publicKey = signatureService.getPublicKeyBase64();
            return Response.ok(publicKey).build();
        } catch (Exception e) {
            return Response.status(Response.Status.INTERNAL_SERVER_ERROR)
                    .entity("Error getting public key: " + e.getMessage())
                    .build();
        }
    }

    @GET
    @Path("/health")
    public Response health() {
        return Response.ok(new HealthResponse("Quarkus Crypto Service is running")).build();
    }
}