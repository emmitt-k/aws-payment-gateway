## Context

The auth service currently has several critical security and performance issues that need to be addressed to achieve full compliance with openspec specifications. The main problems are inefficient API key validation using brute-force scanning, missing TTL implementation, and lack of proper idempotency management.

## Goals / Non-Goals

### Goals
- Implement secure and efficient API key validation
- Add automatic key expiration with TTL
- Implement idempotency for all auth operations
- Enhance security with proper rate limiting
- Improve audit logging with DynamoDB integration
- Add comprehensive error handling with specific error codes

### Non-Goals
- Complete redesign of the auth service architecture
- Migration away from the hybrid PostgreSQL/DynamoDB approach
- Addition of new authentication methods (OAuth, JWT, etc.)

## Decisions

### Decision 1: API Key Validation Flow
- **What**: Update ValidateApiKey use case to accept raw API keys and perform hashing internally
- **Why**: Current implementation has a mismatch where middleware passes raw keys but use case expects hashes
- **Alternatives considered**: 
  - Update middleware to hash keys before passing to use case
  - Create separate validation endpoints for raw vs hashed keys
- **Chosen approach**: Modify use case to maintain backward compatibility and centralize validation logic

### Decision 2: DynamoDB Key Structure
- **What**: Implement composite key structure with PK=ACCOUNT#id and SK=APIKEY#id, plus GSI for key hash lookup
- **Why**: Current scan-based approach is inefficient and exposes all active keys during validation
- **Alternatives considered**:
  - Use only key hash as primary key (loses account grouping)
  - Implement separate lookup table (adds complexity)
- **Chosen approach**: Composite key with GSI provides both efficient lookup and account grouping

### Decision 3: TTL Implementation
- **What**: Add TTL attribute to DynamoDB table based on expires_at field
- **Why**: Automatic cleanup of expired keys reduces operational overhead
- **Alternatives considered**:
  - Manual cleanup via Lambda function
  - Database-level triggers
- **Chosen approach**: Native DynamoDB TTL is most efficient and cost-effective

### Decision 4: Idempotency Storage
- **What**: Use DynamoDB for idempotency keys with 24-hour TTL
- **Why**: Prevents duplicate processing of auth operations
- **Alternatives considered**:
  - Redis (requires additional infrastructure)
  - PostgreSQL (adds load to primary database)
- **Chosen approach**: DynamoDB provides required TTL and scalability

## Risks / Trade-offs

### Risk 1: Breaking Changes
- **Risk**: API key validation flow changes could break existing clients
- **Mitigation**: Maintain backward compatibility in middleware, gradual rollout

### Risk 2: Performance Impact
- **Risk**: New validation logic might impact authentication performance
- **Mitigation**: Implement comprehensive performance testing and monitoring

### Risk 3: Data Migration
- **Risk**: DynamoDB schema changes require careful migration
- **Mitigation**: Implement backward-compatible migration with rollback procedures

### Trade-off 1: Storage Cost
- **Trade-off**: Additional indexes and TTL attributes increase storage costs
- **Justification**: Improved security and performance outweigh minimal cost increase

### Trade-off 2: Complexity
- **Trade-off**: More sophisticated validation logic increases code complexity
- **Justification**: Security and compliance requirements necessitate this complexity

## Migration Plan

### Phase 1: Preparation
1. Create new DynamoDB table with updated schema
2. Implement backward-compatible migration script
3. Add comprehensive monitoring and alerting

### Phase 2: Implementation
1. Deploy updated auth service with feature flags
2. Migrate existing API keys to new format
3. Enable new validation logic gradually

### Phase 3: Cleanup
1. Remove old validation code
2. Clean up deprecated DynamoDB attributes
3. Update documentation and runbooks

### Rollback Procedures
1. Feature flags to immediately revert to old validation
2. Database backup restoration procedures
3. Emergency communication plan for service degradation

## Open Questions

1. Should we implement a grace period for expired keys to prevent service disruption?
2. What rate limits should be applied to different auth operations?
3. Should we implement key rotation policies automatically?
4. How should we handle audit log retention beyond 90 days for compliance?
5. Should we implement different TTL values for different types of audit events?

## Security Considerations

1. **Constant-time comparison**: Implement constant-time operations for key validation
2. **Information leakage**: Ensure error messages don't reveal system information
3. **Attack vectors**: Protect against timing attacks and brute force attempts
4. **Data encryption**: Ensure all sensitive data is encrypted at rest and in transit
5. **Access patterns**: Design DynamoDB access patterns to minimize data exposure

## Performance Requirements

1. **Authentication latency**: Target <50ms for 95th percentile
2. **Throughput**: Support 1000+ auth requests per second
3. **Scalability**: Auto-scale to handle peak loads without degradation
4. **Efficiency**: Minimize DynamoDB read/write operations

## Monitoring and Observability

1. **Key metrics**: Authentication success/failure rates, validation latency
2. **Business metrics**: API key creation/revocation rates, account growth
3. **System metrics**: DynamoDB performance, error rates, TTL operations
4. **Security metrics**: Failed authentication patterns, rate limit violations