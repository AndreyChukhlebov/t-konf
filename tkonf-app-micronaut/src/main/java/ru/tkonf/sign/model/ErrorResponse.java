package ru.tkonf.sign.model;

import io.micronaut.serde.annotation.Serdeable;

@Serdeable
public class ErrorResponse {
    private String error;

    public ErrorResponse() {
    }

    public ErrorResponse(String error) {
        this.error = error;
    }

    public String getError() {
        return error;
    }

    public void setError(String error) {
        this.error = error;
    }
}