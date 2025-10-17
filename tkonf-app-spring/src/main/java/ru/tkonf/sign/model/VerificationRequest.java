package ru.tkonf.sign.model;

public class VerificationRequest {
    private String message;
    private String signature;

    public VerificationRequest() {}

    public VerificationRequest(String message, String signature) {
        this.message = message;
        this.signature = signature;
    }

    // Геттеры и сеттеры
    public String getMessage() { return message; }
    public void setMessage(String message) { this.message = message; }

    public String getSignature() { return signature; }
    public void setSignature(String signature) { this.signature = signature; }
}