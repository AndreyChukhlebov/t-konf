package ru.tkonf.sign;

import io.quarkus.test.junit.QuarkusTest;
import io.restassured.http.ContentType;
import org.junit.jupiter.api.Test;

import static io.restassured.RestAssured.given;
import static org.hamcrest.CoreMatchers.is;
import static org.hamcrest.CoreMatchers.notNullValue;

@QuarkusTest
class GreetingResourceTest {

    @Test
    public void testHealthEndpoint() {
        given()
                .when().get("/api/crypto/health")
                .then()
                .statusCode(200)
                .body("status", is("Quarkus Crypto Service is running"));
    }


    @Test
    public void testSignWithSpecialCharacters() {
        String requestBody = "{\"message\":\"Hello @#$%^&*() World!\",\"algorithm\":\"RSA\"}";

        given()
                .contentType(ContentType.JSON)
                .body(requestBody)
                .when().post("/api/crypto/sign")
                .then()
                .statusCode(200)
                .body("originalMessage", is("Hello @#$%^&*() World!"))
                .body("signature", notNullValue());
    }



    @Test
    public void testSignWithLongMessage() {
        String longMessage = "A".repeat(1000);
        String requestBody = "{\"message\":\"" + longMessage + "\",\"algorithm\":\"RSA\"}";

        given()
                .contentType(ContentType.JSON)
                .body(requestBody)
                .when().post("/api/crypto/sign")
                .then()
                .statusCode(200)
                .body("originalMessage", is(longMessage))
                .body("signature", notNullValue());
    }

    @Test
    public void testSignWithoutAlgorithm() {
        String requestBody = "{\"message\":\"Test without algorithm\"}";

        given()
                .contentType(ContentType.JSON)
                .body(requestBody)
                .when().post("/api/crypto/sign")
                .then()
                .statusCode(200)
                .body("originalMessage", is("Test without algorithm"))
                .body("signature", notNullValue());
    }

    @Test
    public void testVerifyWithNullMessage() {
        String requestBody = "{\"message\":null,\"signature\":\"some_signature\"}";

        given()
                .contentType(ContentType.JSON)
                .body(requestBody)
                .when().post("/api/crypto/verify")
                .then()
                .statusCode(400);
    }
}