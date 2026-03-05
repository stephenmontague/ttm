import { NextResponse } from "next/server";
import { cookies } from "next/headers";

const BACKEND_URL = process.env.API_URL || "http://localhost:9090/api";
const COOKIE_NAME = process.env.SESSION_COOKIE_NAME || "ttm_session";

export async function POST() {
  const cookieStore = await cookies();
  const sessionCookie = cookieStore.get(COOKIE_NAME);

  await fetch(`${BACKEND_URL}/auth/logout`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      ...(sessionCookie
        ? { Cookie: `${COOKIE_NAME}=${sessionCookie.value}` }
        : {}),
    },
  });

  // Clear cookie in browser regardless of backend response.
  const response = NextResponse.json({ ok: true });
  response.cookies.set(COOKIE_NAME, "", { maxAge: -1, path: "/" });
  return response;
}
