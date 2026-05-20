package keel

// ManagedProxyHeaders returns headers used by provider wrappers for proxy idempotency.
func ManagedProxyHeaders(idempotencyKey string) map[string]string {
	if idempotencyKey == "" {
		return nil
	}
	return map[string]string{"Idempotency-Key": idempotencyKey}
}

// BuildManagedProxyPayload adds Keel governance fields to a provider-shaped proxy payload.
func BuildManagedProxyPayload(providerPayload map[string]any, projectID string, subject *PermitSubject, parentPermitID, sessionID *string) map[string]any {
	payload := map[string]any{}
	metadata := map[string]any{}

	for key, value := range providerPayload {
		if key == "" || value == nil {
			continue
		}
		if key == "metadata" {
			if valueMap, ok := value.(map[string]any); ok {
				for metadataKey, metadataValue := range valueMap {
					if metadataKey != "" && metadataValue != nil {
						metadata[metadataKey] = metadataValue
					}
				}
				continue
			}
		}
		payload[key] = value
	}

	payload["project_id"] = projectID
	if parentPermitID != nil && *parentPermitID != "" {
		metadata["keel_parent_permit_id"] = *parentPermitID
	}
	if sessionID != nil && *sessionID != "" {
		metadata["keel_session_id"] = *sessionID
	}
	if len(metadata) > 0 {
		payload["metadata"] = metadata
	}
	if identity := identityFromSubject(subject); identity != nil {
		payload["identity"] = identity
	}
	return payload
}

func identityFromSubject(subject *PermitSubject) *Identity {
	if subject == nil || subject.ID == "" {
		return nil
	}

	id := subject.ID
	confidence := ConfidenceExact
	switch subject.Type {
	case "user", "service_account", "agent", "system", "anonymous":
		subjectType := subject.Type
		return &Identity{
			Subject: &IdentitySubject{
				Type:       &subjectType,
				ID:         &id,
				Confidence: confidence,
			},
		}
	case "api_key":
		credentialType := subject.Type
		return &Identity{
			Credential: &IdentityCredential{
				Type:       &credentialType,
				ID:         &id,
				Confidence: confidence,
			},
		}
	default:
		externalID := subject.Type + ":" + subject.ID
		return &Identity{
			Subject: &IdentitySubject{
				ID:         &id,
				ExternalID: &externalID,
				Confidence: confidence,
			},
		}
	}
}
