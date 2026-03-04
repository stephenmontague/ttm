import { NextResponse } from "next/server";
import { backendGet } from "@/lib/backend";

export async function GET(
  _request: Request,
  { params }: { params: Promise<{ slug: string }> }
) {
  const { slug } = await params;

  if (!slug) {
    return NextResponse.json({ error: "Slug is required" }, { status: 400 });
  }

  const { data, status, ok } = await backendGet(`/companies/${slug}`);

  if (!ok) {
    return NextResponse.json(
      { error: status === 404 ? "Company not found" : "Failed to fetch company" },
      { status }
    );
  }

  return NextResponse.json(data);
}
