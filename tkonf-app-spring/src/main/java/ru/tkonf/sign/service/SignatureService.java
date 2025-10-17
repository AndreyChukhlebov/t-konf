package ru.tkonf.sign.service;




import org.springframework.stereotype.Service;

import javax.crypto.Cipher;
import java.security.KeyPair;
import java.security.KeyPairGenerator;
import java.security.NoSuchAlgorithmException;
import java.security.Signature;
import java.util.Base64;

@Service
public class SignatureService {
    private KeyPair keyPair;
    private final String algorithm;

    public SignatureService() throws NoSuchAlgorithmException {
        this.algorithm = "RSA";
        generateKeyPair();
    }

    public SignatureService(String algorithm) throws NoSuchAlgorithmException {
        this.algorithm = algorithm;
        generateKeyPair();
    }

    private void generateKeyPair() throws NoSuchAlgorithmException {
        KeyPairGenerator keyGen = KeyPairGenerator.getInstance(algorithm);
        keyGen.initialize(2048); // Размер ключа
        this.keyPair = keyGen.generateKeyPair();
    }

    /**
     * Подписывает сообщение с использованием приватного ключа
     */
    public String signMessage(String message) throws Exception {
        Signature signature = Signature.getInstance("SHA256withRSA");
        signature.initSign(keyPair.getPrivate());
        signature.update(message.getBytes());

        byte[] digitalSignature = signature.sign();
        return Base64.getEncoder().encodeToString(digitalSignature);
    }

    /**
     * Проверяет подпись сообщения с использованием публичного ключа
     */
    public boolean verifySignature(String message, String signatureBase64) throws Exception {
        Signature signature = Signature.getInstance("SHA256withRSA");
        signature.initVerify(keyPair.getPublic());
        signature.update(message.getBytes());

        byte[] digitalSignature = Base64.getDecoder().decode(signatureBase64);
        return signature.verify(digitalSignature);
    }

    /**
     * Шифрует сообщение с использованием публичного ключа
     */
    public String encrypt(String message) throws Exception {
        Cipher cipher = Cipher.getInstance(algorithm);
        cipher.init(Cipher.ENCRYPT_MODE, keyPair.getPublic());
        byte[] encryptedBytes = cipher.doFinal(message.getBytes());
        return Base64.getEncoder().encodeToString(encryptedBytes);
    }

    /**
     * Расшифровывает сообщение с использованием приватного ключа
     */
    public String decrypt(String encryptedMessage) throws Exception {
        Cipher cipher = Cipher.getInstance(algorithm);
        cipher.init(Cipher.DECRYPT_MODE, keyPair.getPrivate());
        byte[] decryptedBytes = cipher.doFinal(Base64.getDecoder().decode(encryptedMessage));
        return new String(decryptedBytes);
    }

    public String getPublicKeyBase64() {
        return Base64.getEncoder().encodeToString(keyPair.getPublic().getEncoded());
    }

    public String getPrivateKeyBase64() {
        return Base64.getEncoder().encodeToString(keyPair.getPrivate().getEncoded());
    }

    public String getAlgorithm() {
        return algorithm;
    }
}