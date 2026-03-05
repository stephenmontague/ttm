const BACKEND_URL = process.env.API_URL || "http://localhost:9090/api";

interface BackendResponse<T = unknown> {
  data: T;
  status: number;
  ok: boolean;
}

export async function backendGet<T = unknown>(path: string, cookieHeader?: string): Promise<BackendResponse<T>> {
  const headers: Record<string, string> = { "Content-Type": "application/json" };
  if (cookieHeader) headers["Cookie"] = cookieHeader;

  const res = await fetch(`${BACKEND_URL}${path}`, {
    headers,
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

export async function backendPost<T = unknown>(path: string, body: string, cookieHeader?: string): Promise<BackendResponse<T>> {
  const headers: Record<string, string> = { "Content-Type": "application/json" };
  if (cookieHeader) headers["Cookie"] = cookieHeader;

  const res = await fetch(`${BACKEND_URL}${path}`, {
    method: "POST",
    headers,
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
