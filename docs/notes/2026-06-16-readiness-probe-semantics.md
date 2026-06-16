# Readiness probe semantics — how FX should think about `/healthz`

Captured while triaging `httpserver/controllers/home.go:16` (add a built-in
`/healthz`). The decision needs k8s probe context to be defensible, so the
context lives here.

## The k8s probe model

Three probes, distinct jobs:

| Probe         | Answers                                  | On failure                                            |
|---------------|------------------------------------------|-------------------------------------------------------|
| **startup**   | "Is the app done initializing?"          | Gates liveness/readiness until it passes once         |
| **liveness**  | "Is the process dead/wedged?"            | kubelet **kills the pod**, ReplicaSet restarts it     |
| **readiness** | "Should I send traffic to this pod now?" | kubelet **removes pod from Service endpoints**; pod stays up |

Readiness is specifically about *Service routing*. Failing it doesn't kill
anything — it just stops new traffic. Pass again and the pod returns to
endpoints. "Ready" therefore means "right now, requests will succeed
end-to-end."

## Default cadence

kubelet probe defaults:

- `periodSeconds: 10` — probe every 10s.
- `timeoutSeconds: 1` — fail if no response in 1s.
- `successThreshold: 1` — one pass flips back to ready.
- `failureThreshold: 3` — three consecutive fails before marked unready.
- `initialDelaySeconds: 0` — starts immediately; use a startup probe for
  warmup.

A readiness handler runs **~6/min per pod** in steady state and must respond
in **<1s**. That's the constraint envelope. FX picked a **500ms** budget for
the DB ping — leaves slack so a structured 503 lands before kubelet gives up.

## What readiness should check

The checks whose failure means a real request *will* fail:

- DB connection pool reachable (yes — DB-down means 500s).
- Migrations applied (yes — un-migrated DB is broken).
- App-internal warmup, depends on the app.

What it should **not** include:

- Downstream services you don't own — failing them pulls your pod for someone
  else's outage.
- Slow checks (>~200ms budget).
- Anything flappy. Readiness is a routing decision, not a metric.

## Probe runs continuously — what happens under load

Three scenarios when a pod is up and traffic spikes past capacity:

1. **Probe handler responds fast and 200, app is wedged.** Pod stays in
   endpoints, keeps getting traffic, keeps timing out user requests. This is
   the failure mode of a too-cheap probe — it doesn't reflect what the app is
   actually doing.

2. **Probe handler shares the contested resource and fails or times out.**
   E.g. probe pings the same `*sqlx.DB` pool that's exhausted. The probe queues
   for a connection, exceeds `timeoutSeconds`, kubelet records a failure.
   After `failureThreshold` consecutive fails (~30s), pod is removed from
   Service endpoints. Traffic drains, pod recovers, probe passes, pod is
   re-added. This is k8s's natural backpressure — but it only works if the
   probe exercises the same path real requests use.

3. **All replicas saturated simultaneously.** All readiness probes fail,
   Service has no endpoints, clients get connection refused / 503 with no
   body. Worse than slow responses. Mitigations: HPA scales out *before*
   saturation, PodDisruptionBudget enforces a minimum-ready count, or
   readiness gets less strict.

## The design pitfall — "load-aware readiness" cascades

Self-saturation-aware readiness (queue depth, in-flight count, p99 latency)
easily causes the scenario-3 cascade. Production playbooks keep readiness
about **dep reachability** (DB up, migrations applied) rather than
**self-saturation** (load metrics). Saturation is HPA's domain; readiness
exists so a pod whose *dep* is broken stops wasting requests.

## FX's choice

For the `Home.Healthz` handler:

- **Dep-reachability only.** `db.PingContext(ctx)` with a 500ms deadline.
  No queue-depth check, no Redis (unless added later as a checked-dep).
- **If `data.FromContext` is nil, 200.** App has no DB to be unready against;
  nothing to check, nothing to fail. Same shape extends to other deps later:
  only report what's actually wired.
- **Scenario 1 (probe fine, app wedged) is accepted.** That's HPA / alerts /
  log-based detection's job, not readiness's.

If future work needs both signals: keep `/healthz` as readiness, add `/livez`
as a pure 200, surface saturation via a metric scraped by the autoscaler
(not a k8s probe). Two more endpoints' worth of design — not in scope today.

## References

- k8s probe docs:
  <https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/>
- Probe configuration field defaults are documented in the same page under
  "Configure Probes."
