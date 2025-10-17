package ru.tkonf.sign.model;

import io.micronaut.serde.annotation.Serdeable;

@Serdeable
public class HealthResponse {
    private String status;

    public HealthResponse() {}

    public HealthResponse(String status) {
        this.status = status;
    }

    public String getStatus() { return status; }
    public void setStatus(String status) { this.status = status; }
}