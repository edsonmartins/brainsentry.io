package com.integraltech.brainsentry.service;

import com.integraltech.brainsentry.domain.User;
import com.integraltech.brainsentry.dto.request.LoginRequest;
import com.integraltech.brainsentry.dto.response.LoginResponse;
import com.integraltech.brainsentry.repository.UserJpaRepository;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.security.crypto.password.PasswordEncoder;
import org.springframework.stereotype.Service;

import java.time.Instant;

/**
 * Authentication service.
 */
@Slf4j
@Service
@RequiredArgsConstructor
public class AuthService {

    private final UserJpaRepository userRepo;
    private final PasswordEncoder passwordEncoder;
    private final JwtService jwtService;

    /**
     * Authenticate user with email and password.
     */
    public LoginResponse login(LoginRequest request) {
        log.info("Login attempt for email: {}", request.getEmail());

        User user = userRepo.findByEmail(request.getEmail().toLowerCase())
                .orElseThrow(() -> {
                    log.warn("User not found: {}", request.getEmail());
                    return new IllegalArgumentException("Invalid email or password");
                });

        log.info("User found: {}, active: {}, passwordHash: {}", user.getEmail(), user.getActive(), user.getPasswordHash());

        if (!user.getActive()) {
            throw new IllegalArgumentException("User account is disabled");
        }

        boolean passwordMatches = passwordEncoder.matches(request.getPassword(), user.getPasswordHash());
        log.info("Password matches: {}", passwordMatches);

        if (!passwordMatches) {
            throw new IllegalArgumentException("Invalid email or password");
        }

        // Update last login
        user.setLastLoginAt(Instant.now());
        userRepo.save(user);

        // Generate JWT token
        String token = jwtService.generateToken(
                user.getId(),
                user.getEmail(),
                user.getTenantId(),
                user.getRoles()
        );

        log.info("User logged in successfully: {}", user.getEmail());

        return LoginResponse.builder()
                .token(token)
                .tenantId(user.getTenantId())
                .user(LoginResponse.User.builder()
                        .id(user.getId())
                        .email(user.getEmail())
                        .name(user.getName())
                        .roles(user.getRoles())
                        .build())
                .build();
    }
}
