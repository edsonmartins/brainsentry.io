package com.integraltech.brainsentry.config;

import io.swagger.v3.oas.models.Components;
import io.swagger.v3.oas.models.OpenAPI;
import io.swagger.v3.oas.models.info.Contact;
import io.swagger.v3.oas.models.info.Info;
import io.swagger.v3.oas.models.info.License;
import io.swagger.v3.oas.models.media.Content;
import io.swagger.v3.oas.models.media.MediaType;
import io.swagger.v3.oas.models.media.Schema;
import io.swagger.v3.oas.models.responses.ApiResponse;
import io.swagger.v3.oas.models.security.SecurityRequirement;
import io.swagger.v3.oas.models.security.SecurityScheme;
import io.swagger.v3.oas.models.servers.Server;
import org.springdoc.core.customizers.OpenApiCustomizer;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

import java.util.List;

/**
 * OpenAPI/Swagger configuration for Brain Sentry API documentation.
 *
 * Provides interactive API documentation at /swagger-ui.html
 * and OpenAPI spec at /v3/api-docs
 */
@Configuration
public class OpenApiConfig {

    private static final String SECURITY_SCHEME_NAME = "bearerAuth";

    @Bean
    public OpenAPI brainSentryOpenAPI() {
        return new OpenAPI()
                .info(new Info()
                        .title("Brain Sentry API")
                        .description("""
                                # Brain Sentry - Agent Memory System

                                Sistema de memória para agentes de IA que armazena, recupera e injeta contexto relevante em prompts.

                                ## Autenticação

                                A maioria dos endpoints requer autenticação via JWT Bearer Token.

                                ## Multi-tenancy

                                Use o header `X-Tenant-ID` para especificar o tenant em requisições multi-tenant.

                                ## Categorias de Memória

                                - **PATTERN**: Padrões de design e código
                                - **DECISION**: Decisões arquiteturais
                                - **ANTIPATTERN**: Anti-padrões identificados
                                - **DOMAIN**: Conhecimento de domínio
                                - **BUG**: Bugs conhecidos e soluções
                                - **OPTIMIZATION**: Otimizações de performance
                                - **INTEGRATION**: Integrações com sistemas externos
                                """)
                        .version("1.0.0")
                        .contact(new Contact()
                                .name("IntegrAllTech")
                                .email("support@integraltech.com")
                                .url("https://brainsentry.io"))
                        .license(new License()
                                .name("Apache 2.0")
                                .url("https://www.apache.org/licenses/LICENSE-2.0.html")))
                .servers(List.of(
                        new Server().url("http://localhost:8080").description("Desenvolvimento"),
                        new Server().url("https://api.brainsentry.io").description("Produção")
                ))
                .components(new Components()
                        .addSecuritySchemes(SECURITY_SCHEME_NAME,
                                new SecurityScheme()
                                        .name(SECURITY_SCHEME_NAME)
                                        .type(SecurityScheme.Type.HTTP)
                                        .scheme("bearer")
                                        .bearerFormat("JWT")
                                        .description("Token JWT de autenticação"))
                        .addSchemas("TenantId", new Schema<>()
                                .type("string")
                                .description("ID do Tenant para multi-tenancy")
                                .example("default"))
                        .addSchemas("ErrorResponse", new Schema<>()
                                .type("object")
                                .description("Resposta de erro padrão")
                                .addProperty("timestamp", new Schema<>().type("string").format("date-time"))
                                .addProperty("status", new Schema<>().type("integer").example(400))
                                .addProperty("error", new Schema<>().type("string").example("Bad Request"))
                                .addProperty("message", new Schema<>().type("string"))
                                .addProperty("path", new Schema<>().type("string")))
                )
                .addSecurityItem(new SecurityRequirement().addList(SECURITY_SCHEME_NAME));
    }

    @Bean
    public OpenApiCustomizer brainSentryOpenApiCustomizer() {
        return openApi -> {
            // Add common responses globally
            openApi.getPaths().values().forEach(pathItem -> pathItem.readOperations().forEach(operation -> {
                // Add 400 Bad Request
                operation.getResponses().addApiResponse("400", new ApiResponse()
                        .description("Requisição inválida")
                        .content(new Content()
                                .addMediaType("application/json", new MediaType()
                                        .schema(new Schema<>().$ref("#/components/schemas/ErrorResponse")))));

                // Add 401 Unauthorized
                operation.getResponses().addApiResponse("401", new ApiResponse()
                        .description("Não autorizado - Token JWT inválido ou ausente")
                        .content(new Content()
                                .addMediaType("application/json", new MediaType()
                                        .schema(new Schema<>().$ref("#/components/schemas/ErrorResponse")))));

                // Add 403 Forbidden
                operation.getResponses().addApiResponse("403", new ApiResponse()
                        .description("Acesso negado")
                        .content(new Content()
                                .addMediaType("application/json", new MediaType()
                                        .schema(new Schema<>().$ref("#/components/schemas/ErrorResponse")))));

                // Add 500 Internal Server Error
                operation.getResponses().addApiResponse("500", new ApiResponse()
                        .description("Erro interno do servidor")
                        .content(new Content()
                                .addMediaType("application/json", new MediaType()
                                        .schema(new Schema<>().$ref("#/components/schemas/ErrorResponse")))));
            }));
        };
    }
}
