import { NextResponse } from "next/server";
import { cookies } from "next/headers";
import { backendPost } from "@/lib/backend";

const COOKIE_NAME = process.env.SESSION_COOKIE_NAME || "ttm_session";

export async function POST(
  _request: Request,
  { params }: { params: Promise<{ slug: string }> }
) {
  const { slug } = await params;

  if (!slug) {
    return NextResponse.json({ error: "Slug is required" }, { status: 400 });
  }

  const cookieStore = await cookies();
  const session = cookieStore.get(COOKIE_NAME);
  const cookieHeader = session ? `${COOKIE_NAME}=${session.value}` : "";

  const { data, status, ok } = await backendPost(`/admin/companies/${slug}/reconcile`, "{}", cookieHeader);

  if (!ok) {
    return NextResponse.json(
      { error: "Failed to reconcile workflow status" },
      { status }
    );
  }

  return NextResponse.json(data);
}
