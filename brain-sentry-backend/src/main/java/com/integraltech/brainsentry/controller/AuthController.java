package com.integraltech.brainsentry.controller;

import com.integraltech.brainsentry.dto.request.LoginRequest;
import com.integraltech.brainsentry.dto.response.LoginResponse;
import com.integraltech.brainsentry.service.AuthService;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.tags.Tag;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.http.ResponseEntity;
import java.util.Map;
import org.springframework.web.bind.annotation.*;

/**
 * Authentication controller.
 * Handles login, logout, and token refresh operations.
 */
@Slf4j
@RestController
@RequestMapping("/v1/auth")
@RequiredArgsConstructor
@Tag(name = "Authentication", description = "Authentication endpoints")
public class AuthController {

    private final AuthService authService;

    @PostMapping("/login")
    @Operation(summary = "Login with email and password")
    public ResponseEntity<LoginResponse> login(@Valid @RequestBody LoginRequest request) {
        log.info("Login request for email: {}", request.getEmail());
        try {
            LoginResponse response = authService.login(request);
            return ResponseEntity.ok(response);
        } catch (IllegalArgumentException e) {
            log.warn("Login failed for email: {}", request.getEmail());
            return ResponseEntity.status(401).build();
        }
    }

    @PostMapping("/logout")
    @Operation(summary = "Logout current user")
    public ResponseEntity<Void> logout() {
        // In a stateless JWT setup, logout is handled client-side by removing the token
        // This endpoint exists for future refresh token invalidation
        return ResponseEntity.ok().build();
    }

    @PostMapping("/refresh")
    @Operation(summary = "Refresh access token")
    public ResponseEntity<LoginResponse> refresh(@RequestHeader("Authorization") String authHeader) {
        // For future implementation with refresh tokens
        return ResponseEntity.status(501).build();
    }
}
