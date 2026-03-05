import { NextResponse } from "next/server";
import { cookies } from "next/headers";
import { backendPost } from "@/lib/backend";
import { VALID_SIGNAL_ACTIONS, type SignalAction } from "@/lib/constants";

const COOKIE_NAME = process.env.SESSION_COOKIE_NAME || "ttm_session";

function validateBody(action: SignalAction, body: Record<string, unknown>): string | null {
  switch (action) {
    case "outreach":
      if (!body.channel || typeof body.channel !== "string") return "channel is required";
      if (!body.notes || typeof body.notes !== "string") return "notes is required";
      return null;
    case "contact":
      if (!body.name || typeof body.name !== "string") return "name is required";
      if (!body.role || typeof body.role !== "string") return "role is required";
      return null;
    case "contact_remove":
      if (!body.name || typeof body.name !== "string") return "name is required";
      return null;
    case "agent":
      if (!body.task_type || typeof body.task_type !== "string") return "task_type is required";
      return null;
    case "booked":
      if (!body.notes || typeof body.notes !== "string") return "notes is required";
      return null;
    default:
      return "Invalid action";
  }
}

export async function POST(
  request: Request,
  { params }: { params: Promise<{ slug: string; action: string }> }
) {
  const { slug, action } = await params;

  if (!slug) {
    return NextResponse.json({ error: "Slug is required" }, { status: 400 });
  }

  if (!VALID_SIGNAL_ACTIONS.includes(action as SignalAction)) {
    return NextResponse.json(
      { error: `Invalid signal action: ${action}. Valid actions: ${VALID_SIGNAL_ACTIONS.join(", ")}` },
      { status: 400 }
    );
  }

  let body: Record<string, unknown> = {};
  try {
    const text = await request.text();
    if (text) {
      body = JSON.parse(text);
    }
  } catch {
    return NextResponse.json({ error: "Invalid JSON body" }, { status: 400 });
  }

  const validationError = validateBody(action as SignalAction, body);
  if (validationError) {
    return NextResponse.json({ error: validationError }, { status: 400 });
  }

  // Map frontend action names to backend paths.
  const backendAction = action === "contact_remove" ? "contact/remove" : action;

  const cookieStore = await cookies();
  const session = cookieStore.get(COOKIE_NAME);
  const cookieHeader = session ? `${COOKIE_NAME}=${session.value}` : "";

  const { data, status, ok } = await backendPost(
    `/admin/companies/${slug}/signal/${backendAction}`,
    JSON.stringify(body),
    cookieHeader
  );

  if (!ok) {
    return NextResponse.json(
      { error: "Failed to send signal" },
      { status }
    );
  }

  return NextResponse.json({ sent: true, signal: action, ...data as object });
}
