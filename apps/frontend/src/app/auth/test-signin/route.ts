import { clerkClient } from "@clerk/nextjs/server";
import { NextResponse } from "next/server";
import { z } from "zod";

const BodySchema = z.object({
    email: z.string().email(),
    password: z.string().min(1),
});

function assertDevEnabled() {
    // 開発環境でのみ有効化し、さらに明示的な環境変数が無い限りは拒否する
    if (process.env.NODE_ENV !== "development") return false;
    if (process.env.CINETAG_DEV_TEST_SIGNIN_ENABLED !== "1") return false;
    return true;
}

function toClerkErrorShape(err: unknown): {
    status?: number;
    clerkTraceId?: string;
    errors?: unknown;
} {
    if (!err || typeof err !== "object") return {};
    const anyErr = err as Record<string, unknown>;
    return {
        status: typeof anyErr.status === "number" ? anyErr.status : undefined,
        clerkTraceId:
            typeof anyErr.clerkTraceId === "string" ? anyErr.clerkTraceId : undefined,
        errors: anyErr.errors,
    };
}

export async function POST(req: Request) {
    if (!assertDevEnabled()) {
        return NextResponse.json(
            {
                error:
                    "This endpoint is disabled. Set CINETAG_DEV_TEST_SIGNIN_ENABLED=1 in development only.",
            },
            { status: 403 }
        );
    }

    let body: unknown;
    try {
        body = await req.json();
    } catch {
        return NextResponse.json({ error: "Invalid JSON body" }, { status: 400 });
    }

    const parsed = BodySchema.safeParse(body);
    if (!parsed.success) {
        return NextResponse.json(
            { error: "Invalid request body", details: parsed.error.flatten() },
            { status: 400 }
        );
    }

    const { email, password } = parsed.data;

    try {
        const client = await clerkClient();

        const userList = await client.users.getUserList({
            emailAddress: [email],
            limit: 1,
        });
        const user = userList.data[0];
        // console.log(user);
        if (!user) {
            return NextResponse.json({ error: "Invalid credentials" }, { status: 401 });
        }

        await client.users.verifyPassword({ userId: user.id, password });

        const session = await client.sessions.createSession({ userId: user.id });
        const token = await client.sessions.getToken(session.id, "cinetag-backend");

        return NextResponse.json(
            { accessToken: token.jwt },
            { status: 200, headers: { "Cache-Control": "no-store" } }
        );
    } catch (error) {
        // 開発用途なので、Clerk側の422などの原因が分かるように最小限の詳細を返す
        const shape = toClerkErrorShape(error);
        return NextResponse.json(
            {
                error: "Invalid credentials",
                ...(shape.status ? { clerkStatus: shape.status } : {}),
                ...(shape.clerkTraceId ? { clerkTraceId: shape.clerkTraceId } : {}),
                ...(shape.errors ? { clerkErrors: shape.errors } : {}),
            },
            { status: 401, headers: { "Cache-Control": "no-store" } }
        );
    }
}

export async function GET() {
    return NextResponse.json({ error: "Method Not Allowed" }, { status: 405 });
}


