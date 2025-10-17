package ru.tkonf.sign.model;

import io.micronaut.serde.annotation.Serdeable;

@Serdeable
public class SignatureRequest {
    private String message;
    private String algorithm;

    public SignatureRequest() {}

    public SignatureRequest(String message, String algorithm) {
        this.message = message;
        this.algorithm = algorithm;
    }

    // Геттеры и сеттеры
    public String getMessage() { return message; }
    public void setMessage(String message) { this.message = message; }

    public String getAlgorithm() { return algorithm; }
    public void setAlgorithm(String algorithm) { this.algorithm = algorithm; }
}