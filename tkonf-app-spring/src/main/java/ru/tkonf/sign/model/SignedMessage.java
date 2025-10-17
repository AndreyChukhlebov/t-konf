package ru.tkonf.sign.model;

// SignedMessage.java
public class SignedMessage {
    private String originalMessage;
    private String signature;
    private String algorithm;
    private String publicKey;

    public SignedMessage() {}

    public SignedMessage(String originalMessage, String signature, String algorithm, String publicKey) {
        this.originalMessage = originalMessage;
        this.signature = signature;
        this.algorithm = algorithm;
        this.publicKey = publicKey;
    }

    // Геттеры и сеттеры
    public String getOriginalMessage() { return originalMessage; }
    public void setOriginalMessage(String originalMessage) { this.originalMessage = originalMessage; }

    public String getSignature() { return signature; }
    public void setSignature(String signature) { this.signature = signature; }

    public String getAlgorithm() { return algorithm; }
    public void setAlgorithm(String algorithm) { this.algorithm = algorithm; }

    public String getPublicKey() { return publicKey; }
    public void setPublicKey(String publicKey) { this.publicKey = publicKey; }
}