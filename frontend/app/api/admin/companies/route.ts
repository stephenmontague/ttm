import { NextResponse } from "next/server";
import { cookies } from "next/headers";
import { backendGet, backendPost } from "@/lib/backend";

const COOKIE_NAME = process.env.SESSION_COOKIE_NAME || "ttm_session";

async function getSessionCookie(): Promise<string> {
  const cookieStore = await cookies();
  const session = cookieStore.get(COOKIE_NAME);
  return session ? `${COOKIE_NAME}=${session.value}` : "";
}

export async function GET() {
  const cookieHeader = await getSessionCookie();
  const { data, status, ok } = await backendGet("/admin/companies", cookieHeader);

  if (!ok) {
    return NextResponse.json(
      { error: "Failed to fetch companies" },
      { status }
    );
  }

  return NextResponse.json(data);
}

export async function POST(request: Request) {
  let body: { companyName?: string };
  try {
    body = await request.json();
  } catch {
    return NextResponse.json({ error: "Invalid JSON body" }, { status: 400 });
  }

  if (!body.companyName || typeof body.companyName !== "string" || !body.companyName.trim()) {
    return NextResponse.json(
      { error: "companyName is required" },
      { status: 400 }
    );
  }

  const cookieHeader = await getSessionCookie();
  const { data, status, ok } = await backendPost(
    "/admin/companies",
    JSON.stringify({ companyName: body.companyName.trim() }),
    cookieHeader
  );

  if (!ok) {
    return NextResponse.json(
      { error: "Failed to create workflow" },
      { status }
    );
  }

  return NextResponse.json(data, { status: 201 });
}
