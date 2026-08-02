"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { api } from "@/lib/api";
import { useData } from "@/lib/hooks";

export default function LoginPage() {
  const router = useRouter();
  const { data: setup, loading } = useData(
    () => api.get<{ needsSetup: boolean }>("/api/setup"),
    []
  );
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [confirm, setConfirm] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  const needsSetup = setup?.needsSetup ?? false;

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    setError(null);
    if (needsSetup && password !== confirm) {
      setError("As passwords não coincidem.");
      return;
    }
    setBusy(true);
    try {
      if (needsSetup) {
        await api.post("/api/setup", { username, password });
      } else {
        await api.post("/api/login", { username, password });
      }
      router.replace("/");
    } catch (err) {
      setError((err as Error).message);
      setBusy(false);
    }
  }

  if (loading) return null;

  return (
    <main className="content" style={{ display: "flex", alignItems: "center" }}>
      <form onSubmit={submit} style={{ width: "100%", maxWidth: 380, margin: "0 auto" }}>
        <div style={{ textAlign: "center", marginBottom: 24 }}>
          {/* eslint-disable-next-line @next/next/no-img-element */}
          <img src="/icon.svg" alt="" width={64} height={64} style={{ borderRadius: 16 }} />
          <h1 style={{ fontSize: 24, margin: "12px 0 4px" }}>Bolso</h1>
          <p style={{ color: "var(--muted)", margin: 0, fontSize: 14 }}>
            {needsSetup
              ? "Primeira utilização: cria o teu acesso."
              : "Contas da casa"}
          </p>
        </div>
        <div className="card" style={{ padding: 16 }}>
          <div className="field">
            <label htmlFor="lg-user">Utilizador</label>
            <input
              id="lg-user"
              autoComplete="username"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              placeholder={needsSetup ? "por exemplo, loureiro" : ""}
            />
          </div>
          <div className="field">
            <label htmlFor="lg-pass">Password</label>
            <input
              id="lg-pass"
              type="password"
              autoComplete={needsSetup ? "new-password" : "current-password"}
              value={password}
              onChange={(e) => setPassword(e.target.value)}
            />
          </div>
          {needsSetup && (
            <div className="field">
              <label htmlFor="lg-confirm">Repetir password</label>
              <input
                id="lg-confirm"
                type="password"
                autoComplete="new-password"
                value={confirm}
                onChange={(e) => setConfirm(e.target.value)}
              />
            </div>
          )}
          {error && <p className="error">{error}</p>}
          <button className="btn" style={{ width: "100%", marginTop: 6 }} disabled={busy}>
            {needsSetup ? "Criar acesso" : "Entrar"}
          </button>
        </div>
      </form>
    </main>
  );
}
