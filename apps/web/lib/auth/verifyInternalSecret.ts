import { timingSafeEqual } from "node:crypto";

export function verifyInternalSecret(request: Request): boolean {
  const expected = process.env.COOP_INTERNAL_SECRET;
  if (!expected) return false;

  const provided = request.headers.get("x-coop-internal-secret");
  if (!provided) return false;

  const expectedBuf = Buffer.from(expected);
  const providedBuf = Buffer.from(provided);
  if (expectedBuf.length !== providedBuf.length) return false;

  return timingSafeEqual(expectedBuf, providedBuf);
}
