const BACKEND_URL = process.env.API_URL || "http://localhost:9090/api";

interface BackendResponse<T = unknown> {
  data: T;
  status: number;
  ok: boolean;
}

export async function backendGet<T = unknown>(path: string): Promise<BackendResponse<T>> {
  const res = await fetch(`${BACKEND_URL}${path}`, {
    headers: { "Content-Type": "application/json" },
    cache: "no-store",
  });

  let data: T;
  try {
    data = await res.json();
  } catch {
    data = { error: "Invalid response from backend" } as T;
  }

  return { data, status: res.status, ok: res.ok };
}

export async function backendPost<T = unknown>(path: string, body: string): Promise<BackendResponse<T>> {
  const res = await fetch(`${BACKEND_URL}${path}`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body,
  });

  let data: T;
  try {
    data = await res.json();
  } catch {
    data = { error: "Invalid response from backend" } as T;
  }

  return { data, status: res.status, ok: res.ok };
}
