import { NextResponse } from "next/server";
import { APIError } from "better-auth";
import { auth } from "@/lib/auth/auth";
import { verifyInternalSecret } from "@/lib/auth/verifyInternalSecret";

interface GithubUser {
  id: number;
  login: string;
  name: string | null;
  avatar_url: string;
}

export async function POST(request: Request) {
  if (!verifyInternalSecret(request)) {
    return new NextResponse(null, { status: 404 });
  }

  const body = await request.json().catch(() => null);
  const githubAccessToken = body?.githubAccessToken;
  if (typeof githubAccessToken !== "string" || !githubAccessToken) {
    return NextResponse.json({ error: "githubAccessToken is required" }, { status: 400 });
  }

  const profileRes = await fetch("https://api.github.com/user", {
    headers: {
      Authorization: `Bearer ${githubAccessToken}`,
      "User-Agent": "coop",
    },
  });
  if (!profileRes.ok) {
    return NextResponse.json({ error: "invalid githubAccessToken" }, { status: 401 });
  }
  const profile = (await profileRes.json()) as GithubUser;

  try {
    const result = await auth.api.signInSocial({
      body: {
        provider: "github",
        idToken: {
          token: githubAccessToken,
          accessToken: githubAccessToken,
        },
      },
    });

    if (!("user" in result) || !result.user) {
      return NextResponse.json({ error: "failed to resolve github user" }, { status: 502 });
    }

    return NextResponse.json({
      userId: result.user.id,
      username: profile.login,
      displayName: profile.name || profile.login,
      avatarUrl: profile.avatar_url,
    });
  } catch (err) {
    if (err instanceof APIError) {
      return NextResponse.json({ error: err.body?.message ?? err.message }, { status: err.statusCode });
    }
    throw err;
  }
}
