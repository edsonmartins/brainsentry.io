package com.integraltech.brainsentry.config;

import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Component;

/**
 * Simple tenant identifier resolver.
 *
 * TODO: Implement full Hibernate 6 multi-tenancy after basic tests pass.
 */
@Slf4j
@Component
public class TenantIdentifierResolver {

    private static final String DEFAULT_TENANT = "default";

    public String getTenantId() {
        // TODO: Return tenant from context when multi-tenancy is implemented
        return DEFAULT_TENANT;
    }
}
