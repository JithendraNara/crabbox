import type { LeaseConfig } from "./config";
import { normalizeLeaseSlug } from "./slug";

export function leaseProviderLabels(
  config: LeaseConfig,
  leaseID: string,
  slug: string,
  owner: string,
  provider: string,
  now: Date,
  extra: Record<string, string> = {},
): Record<string, string> {
  return {
    class: config.class,
    crabbox: "true",
    created_by: "crabbox",
    keep: String(config.keep),
    lease: leaseID,
    slug: normalizeLeaseSlug(slug),
    owner: sanitizeLabel(owner),
    profile: config.profile,
    provider_key: config.providerKey,
    provider,
    server_type: config.serverType,
    state: "leased",
    created_at: now.toISOString(),
    last_touched_at: now.toISOString(),
    idle_timeout_secs: String(config.idleTimeoutSeconds),
    expires_at: new Date(
      now.getTime() + Math.min(config.ttlSeconds, config.idleTimeoutSeconds) * 1000,
    ).toISOString(),
    ...extra,
  };
}

function sanitizeLabel(value: string): string {
  return value.replaceAll(/[^a-zA-Z0-9_.@-]/g, "_").slice(0, 63) || "unknown";
}
