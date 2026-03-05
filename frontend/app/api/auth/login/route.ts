import { NextResponse } from "next/server";

const BACKEND_URL = process.env.API_URL || "http://localhost:9090/api";

export async function POST(request: Request) {
  let body: { email?: string; password?: string };
  try {
    body = await request.json();
  } catch {
    return NextResponse.json({ error: "Invalid JSON body" }, { status: 400 });
  }

  if (!body.email || !body.password) {
    return NextResponse.json(
      { error: "email and password are required" },
      { status: 400 }
    );
  }

  const res = await fetch(`${BACKEND_URL}/auth/login`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ email: body.email, password: body.password }),
  });

  const data = await res.json().catch(() => ({}));

  if (!res.ok) {
    return NextResponse.json(data, { status: res.status });
  }

  // Forward the Set-Cookie header from Go to the browser.
  const setCookie = res.headers.get("set-cookie");
  const response = NextResponse.json({ ok: true });
  if (setCookie) {
    response.headers.set("set-cookie", setCookie);
  }
  return response;
}
