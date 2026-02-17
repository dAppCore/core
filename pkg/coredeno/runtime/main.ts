// CoreDeno Runtime Entry Point
// Connects to CoreGO via gRPC over Unix socket.
// Implements DenoService for module lifecycle management.

const socketPath = Deno.env.get("CORE_SOCKET");
if (!socketPath) {
  console.error("FATAL: CORE_SOCKET environment variable not set");
  Deno.exit(1);
}

console.error(`CoreDeno: connecting to ${socketPath}`);

// Tier 1: signal readiness and stay alive.
// Tier 2 adds the gRPC client and DenoService implementation.
console.error("CoreDeno: ready");

// Keep alive until parent sends SIGTERM
const ac = new AbortController();
Deno.addSignalListener("SIGTERM", () => {
  console.error("CoreDeno: shutting down");
  ac.abort();
});

try {
  await new Promise((_resolve, reject) => {
    ac.signal.addEventListener("abort", () => reject(new Error("shutdown")));
  });
} catch {
  // Clean exit on SIGTERM
}
