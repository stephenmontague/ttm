import { NextResponse } from "next/server";
import { backendGet, backendPost } from "@/lib/backend";

export async function GET() {
  const { data, status, ok } = await backendGet("/admin/companies");

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

  const { data, status, ok } = await backendPost(
    "/admin/companies",
    JSON.stringify({ companyName: body.companyName.trim() })
  );

  if (!ok) {
    return NextResponse.json(
      { error: "Failed to create workflow" },
      { status }
    );
  }

  return NextResponse.json(data, { status: 201 });
}
