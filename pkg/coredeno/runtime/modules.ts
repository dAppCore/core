// Module registry — tracks loaded modules and their lifecycle status.
// Tier 2: status tracking only. Tier 3 adds real Deno worker isolates.

export type ModuleStatus = "UNKNOWN" | "LOADING" | "RUNNING" | "STOPPED" | "ERRORED";

// Status enum values matching the proto definition.
export const StatusEnum: Record<ModuleStatus, number> = {
  UNKNOWN: 0,
  LOADING: 1,
  RUNNING: 2,
  STOPPED: 3,
  ERRORED: 4,
};

export interface Module {
  code: string;
  entryPoint: string;
  permissions: string[];
  status: ModuleStatus;
}

export class ModuleRegistry {
  private modules = new Map<string, Module>();

  load(code: string, entryPoint: string, permissions: string[]): void {
    this.modules.set(code, {
      code,
      entryPoint,
      permissions,
      status: "RUNNING",
    });
    console.error(`CoreDeno: module loaded: ${code}`);
  }

  unload(code: string): boolean {
    const mod = this.modules.get(code);
    if (!mod) return false;
    mod.status = "STOPPED";
    console.error(`CoreDeno: module unloaded: ${code}`);
    return true;
  }

  status(code: string): ModuleStatus {
    return this.modules.get(code)?.status ?? "UNKNOWN";
  }

  list(): Module[] {
    return Array.from(this.modules.values());
  }
}
