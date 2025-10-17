package ru.tkonf.sign.model;

import io.micronaut.serde.annotation.Serdeable;

// CryptoResponse.java
@Serdeable
public class CryptoResponse {
    private String result;

    public CryptoResponse() {}

    public CryptoResponse(String result) {
        this.result = result;
    }

    public String getResult() { return result; }
    public void setResult(String result) { this.result = result; }
}
