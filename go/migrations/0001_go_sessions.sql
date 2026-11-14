CREATE TABLE IF NOT EXISTS "SantaGoSessions" (
 "Id" uuid PRIMARY KEY,
 "UserId" uuid NOT NULL REFERENCES "AspNetUsers"("Id") ON DELETE CASCADE,
 "TokenHash" text NOT NULL UNIQUE,
 "CsrfHash" text NOT NULL,
 "ExpiresAt" timestamptz NOT NULL,
 "CreatedAt" timestamptz NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS "IX_SantaGoSessions_ExpiresAt" ON "SantaGoSessions" ("ExpiresAt");
