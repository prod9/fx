#!/usr/bin/env python3

import anyio
import sys
import os
import dagger

PLATFORM = os.getenv("PLATFORM", "linux/amd64")
IMAGE = os.getenv("IMAGE", "ghcr.io/prod9/fx")
IGNORES = [
    "*.docker",
    ".dockerignore",
    ".DS_Store",
    ".env",
    ".env.local",
    ".git",
    ".github",
    ".gitignore",
    ".idea",
    "build.py",
]


async def gitcommit():
    args = ["git", "rev-parse", "--short", "HEAD"]
    proc = await anyio.run_process(args)
    return proc.stdout.decode().strip()


async def build():
    config = dagger.Config(
        log_output=sys.stderr,
        execute_timeout=5*60,
    )

    async with dagger.Connection(config) as conn:
        hostdir = conn.host().directory(".", exclude=IGNORES)

        base = (
            conn.container(platform=dagger.Platform(PLATFORM))
            .from_("alpine:edge")
            .with_label(
                "org.opencontainers.image.source",
                "https://github.com/prod9/fx",
            )
            .with_workdir("/app")
        )

        builder = (
            base
            .with_exec([
                "apk", "add", "--no-cache", "build-base", "git", "go",
                "pkgconfig", "openssl-dev",
            ])
            .with_file("go.mod", hostdir.file("go.mod"))
            .with_file("go.sum", hostdir.file("go.sum"))
            .with_exec(["go", "mod", "download", "-x", "all"])

            .with_env_variable("CGO_ENABLED", "0")
            .with_directory("/app", hostdir)
            .with_exec([
                "go", "build", "-v",
                "-o", "/app/vanity",
                "fx.prodigy9.co"
            ])
        )

        image = (
            base
            .with_exec([
                "apk", "add", "--no-cache",
                "tzdata", "ca-certificates"
            ])
            .with_file("vanity", builder.file("vanity"))
            .with_default_args(["/app/vanity", "serve"])
        )

        commit = await gitcommit()
        image = await image.sync()
        async with anyio.create_task_group() as tg:
            tg.start_soon(image.publish, f"{IMAGE}:latest")
            tg.start_soon(image.publish, f"{IMAGE}:{commit}")


if __name__ == "__main__":
    anyio.run(build)
