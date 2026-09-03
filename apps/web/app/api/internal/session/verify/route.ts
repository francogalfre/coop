import { NextResponse } from "next/server";
import { getCookies } from "better-auth/cookies";
import { auth } from "@/lib/auth/auth";
import { verifyInternalSecret } from "@/lib/auth/verifyInternalSecret";

export async function POST(request: Request) {
  if (!verifyInternalSecret(request)) {
    return new NextResponse(null, { status: 404 });
  }

  const body = await request.json().catch(() => null);
  const cookie = body?.cookie;
  if (typeof cookie !== "string" || !cookie) {
    return NextResponse.json({ error: "cookie is required" }, { status: 400 });
  }

  const { sessionToken } = getCookies(auth.options);
  const headers = new Headers();
  headers.set("cookie", `${sessionToken.name}=${cookie}`);

  const result = await auth.api.getSession({ headers });
  if (!result?.user) {
    return new NextResponse(null, { status: 404 });
  }

  return NextResponse.json({ userId: result.user.id, name: result.user.name, image: result.user.image ?? "" });
}
