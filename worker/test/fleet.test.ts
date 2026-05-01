import { describe, expect, it } from "vitest";

import { FleetDurableObject } from "../src/fleet";
import type { Env, LeaseRecord } from "../src/types";

class MemoryStorage {
  private readonly values = new Map<string, unknown>();

  async get<T>(key: string): Promise<T | undefined> {
    return this.values.get(key) as T | undefined;
  }

  async put<T>(key: string, value: T): Promise<void> {
    this.values.set(key, value);
  }

  async deleteAlarm(): Promise<void> {}

  async setAlarm(_time: number): Promise<void> {}

  async list<T>({ prefix = "" }: { prefix?: string } = {}): Promise<Map<string, T>> {
    const matches = new Map<string, T>();
    for (const [key, value] of this.values) {
      if (key.startsWith(prefix)) {
        matches.set(key, value as T);
      }
    }
    return matches;
  }

  seed<T>(key: string, value: T): void {
    this.values.set(key, value);
  }

  value<T>(key: string): T | undefined {
    return this.values.get(key) as T | undefined;
  }
}

describe("fleet lease identity and idle", () => {
  it("resolves owner-scoped slugs and heartbeat extends idle expiry", async () => {
    const storage = new MemoryStorage();
    const fleet = testFleet(storage);
    const touchedAt = new Date(Date.now() - 10 * 60 * 1000);
    const expiresAt = new Date(touchedAt.getTime() + 1800 * 1000);
    storage.seed(
      "lease:cbx_000000000001",
      testLease({
        id: "cbx_000000000001",
        slug: "blue-lobster",
        owner: "peter@example.com",
        org: "openclaw",
        createdAt: touchedAt.toISOString(),
        updatedAt: touchedAt.toISOString(),
        lastTouchedAt: touchedAt.toISOString(),
        ttlSeconds: 5400,
        idleTimeoutSeconds: 1800,
        expiresAt: expiresAt.toISOString(),
      }),
    );

    const heartbeat = await fleet.fetch(
      request("POST", "/v1/leases/blue-lobster/heartbeat", {
        headers: {
          "cf-access-authenticated-user-email": "peter@example.com",
          "x-crabbox-org": "openclaw",
        },
        body: { idleTimeoutSeconds: 2400 },
      }),
    );
    expect(heartbeat.status).toBe(200);
    const { lease } = (await heartbeat.json()) as { lease: LeaseRecord };
    expect(lease.id).toBe("cbx_000000000001");
    expect(lease.slug).toBe("blue-lobster");
    expect(lease.idleTimeoutSeconds).toBe(2400);
    expect(Date.parse(lease.expiresAt)).toBeGreaterThan(expiresAt.getTime());
  });
});

describe("fleet run history", () => {
  it("records finished runs and serves logs", async () => {
    const fleet = testFleet();
    const ownerHeaders = {
      "cf-access-authenticated-user-email": "peter@example.com",
      "x-crabbox-org": "openclaw",
    };
    const create = await fleet.fetch(
      request("POST", "/v1/runs", {
        headers: ownerHeaders,
        body: {
          leaseID: "cbx_000000000001",
          provider: "aws",
          class: "beast",
          serverType: "c7a.48xlarge",
          command: ["go", "test", "./..."],
        },
      }),
    );
    expect(create.status).toBe(201);
    const { run } = (await create.json()) as { run: { id: string } };

    const finish = await fleet.fetch(
      request("POST", `/v1/runs/${run.id}/finish`, {
        body: {
          exitCode: 0,
          syncMs: 12,
          commandMs: 34,
          log: "ok\n",
          results: {
            format: "junit",
            files: ["junit.xml"],
            suites: 1,
            tests: 2,
            failures: 1,
            errors: 0,
            skipped: 0,
            timeSeconds: 1.2,
            failed: [{ suite: "pkg", name: "fails", kind: "failure" }],
          },
        },
      }),
    );
    expect(finish.status).toBe(200);
    const finished = (await finish.json()) as {
      run: { state: string; logBytes: number; results?: { tests: number } };
    };
    expect(finished.run.state).toBe("succeeded");
    expect(finished.run.logBytes).toBe(3);
    expect(finished.run.results?.tests).toBe(2);

    const listed = await fleet.fetch(request("GET", "/v1/runs?leaseID=cbx_000000000001"));
    const listBody = (await listed.json()) as { runs: Array<{ id: string; owner: string }> };
    expect(listBody.runs).toHaveLength(1);
    expect(listBody.runs[0]?.id).toBe(run.id);
    expect(listBody.runs[0]?.owner).toBe("peter@example.com");

    const logs = await fleet.fetch(request("GET", `/v1/runs/${run.id}/logs`));
    expect(await logs.text()).toBe("ok\n");
  });

  it("bounds stored result summaries", async () => {
    const fleet = testFleet();
    const create = await fleet.fetch(
      request("POST", "/v1/runs", {
        body: {
          leaseID: "cbx_000000000001",
          provider: "aws",
          class: "beast",
          serverType: "c7a.48xlarge",
          command: ["go", "test", "./..."],
        },
      }),
    );
    expect(create.status).toBe(201);
    const { run } = (await create.json()) as { run: { id: string } };
    const failed = Array.from({ length: 150 }, (_, index) => ({
      suite: "pkg",
      name: `fails-${index}`,
      kind: "failure" as const,
      message: "x".repeat(5000),
    }));

    const finish = await fleet.fetch(
      request("POST", `/v1/runs/${run.id}/finish`, {
        body: {
          exitCode: 1,
          log: "",
          results: {
            format: "junit",
            files: Array.from({ length: 80 }, (_, index) => `junit-${index}.xml`),
            suites: 1,
            tests: 150,
            failures: 150,
            errors: 0,
            skipped: 0,
            timeSeconds: 1.2,
            failed,
          },
        },
      }),
    );
    expect(finish.status).toBe(200);
    const finished = (await finish.json()) as {
      run: { results?: { files: string[]; failed: Array<{ message?: string }> } };
    };
    expect(finished.run.results?.files).toHaveLength(50);
    expect(finished.run.results?.failed).toHaveLength(100);
    expect(
      new TextEncoder().encode(finished.run.results?.failed[0]?.message ?? "").byteLength,
    ).toBe(4096);
  });
});

describe("fleet identity", () => {
  it("reports owner and org from request context", async () => {
    const fleet = testFleet();
    const response = await fleet.fetch(
      request("GET", "/v1/whoami", {
        headers: {
          "cf-access-authenticated-user-email": "peter@example.com",
          "x-crabbox-org": "openclaw",
        },
      }),
    );
    expect(await response.json()).toEqual({
      owner: "peter@example.com",
      org: "openclaw",
      auth: "bearer",
    });
  });
});

function testFleet(storage = new MemoryStorage()): FleetDurableObject {
  return new FleetDurableObject(
    { storage } as unknown as DurableObjectState,
    { CRABBOX_DEFAULT_ORG: "default-org" } as Env,
  );
}

function testLease(overrides: Partial<LeaseRecord>): LeaseRecord {
  return {
    id: "cbx_000000000000",
    provider: "hetzner",
    cloudID: "123",
    owner: "peter@example.com",
    org: "openclaw",
    profile: "default",
    class: "beast",
    serverType: "ccx63",
    serverID: 123,
    serverName: "crabbox-blue-lobster",
    providerKey: "crabbox-cbx-000000000000",
    host: "192.0.2.1",
    sshUser: "crabbox",
    sshPort: "2222",
    workRoot: "/work/crabbox",
    keep: true,
    ttlSeconds: 5400,
    estimatedHourlyUSD: 1,
    maxEstimatedUSD: 1.5,
    state: "active",
    createdAt: "2026-05-01T00:00:00.000Z",
    updatedAt: "2026-05-01T00:00:00.000Z",
    expiresAt: "2026-05-01T01:30:00.000Z",
    ...overrides,
  };
}

function request(
  method: string,
  path: string,
  init: { headers?: Record<string, string>; body?: unknown } = {},
): Request {
  return new Request(`https://crabbox.test${path}`, {
    method,
    headers: {
      ...(init.body === undefined ? {} : { "content-type": "application/json" }),
      ...init.headers,
    },
    body: init.body === undefined ? undefined : JSON.stringify(init.body),
  });
}
