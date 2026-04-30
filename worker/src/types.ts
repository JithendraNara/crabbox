export interface Env {
  FLEET: DurableObjectNamespace;
  HETZNER_TOKEN: string;
  CRABBOX_SHARED_TOKEN?: string;
}

export interface LeaseRequest {
  profile?: string;
  class?: string;
  serverType?: string;
  location?: string;
  image?: string;
  sshUser?: string;
  sshPort?: string;
  providerKey?: string;
  workRoot?: string;
  ttlSeconds?: number;
  keep?: boolean;
  sshPublicKey?: string;
}

export interface LeaseRecord {
  id: string;
  owner: string;
  profile: string;
  class: string;
  serverType: string;
  serverID: number;
  serverName: string;
  host: string;
  sshUser: string;
  sshPort: string;
  workRoot: string;
  keep: boolean;
  state: "active" | "released" | "expired" | "failed";
  createdAt: string;
  updatedAt: string;
  expiresAt: string;
}

export interface HetznerServer {
  id: number;
  name: string;
  status: string;
  labels: Record<string, string>;
  public_net: {
    ipv4: {
      ip: string;
    };
  };
  server_type: {
    name: string;
  };
}

export interface HetznerSSHKey {
  id: number;
  name: string;
  fingerprint: string;
  public_key: string;
}

export interface MachineView {
  id: number;
  name: string;
  status: string;
  serverType: string;
  host: string;
  labels: Record<string, string>;
}
