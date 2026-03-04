import { NextResponse } from "next/server";
import { backendGet } from "@/lib/backend";

export async function GET() {
  const { data, status, ok } = await backendGet("/companies");

  if (!ok) {
    return NextResponse.json(
      { error: "Failed to fetch companies" },
      { status }
    );
  }

  return NextResponse.json(data);
}
