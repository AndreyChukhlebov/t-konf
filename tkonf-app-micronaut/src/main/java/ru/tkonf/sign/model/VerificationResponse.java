package ru.tkonf.sign.model;

import io.micronaut.serde.annotation.Serdeable;

// VerificationResponse.java
@Serdeable
public class VerificationResponse {
    private boolean valid;
    private String message;

    public VerificationResponse() {}

    public VerificationResponse(boolean valid, String message) {
        this.valid = valid;
        this.message = message;
    }

    // Геттеры и сеттеры
    public boolean isValid() { return valid; }
    public void setValid(boolean valid) { this.valid = valid; }

    public String getMessage() { return message; }
    public void setMessage(String message) { this.message = message; }
}