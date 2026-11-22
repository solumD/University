#!/usr/bin/env python3
"""Hold a local demonstration DB write lock while running a client command."""

import argparse
import pathlib
import sqlite3
import subprocess
import threading


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--db", default="data/broker.db")
    parser.add_argument("--seconds", type=float, required=True)
    parser.add_argument("command", nargs=argparse.REMAINDER)
    args = parser.parse_args()
    command = args.command[1:] if args.command[:1] == ["--"] else args.command
    if not command or not 0 < args.seconds <= 60:
        parser.error("provide a client command after -- and a duration in (0, 60]")
    if not pathlib.Path(args.db).is_file():
        parser.error("database must already exist; start the server first")

    ready = threading.Event()
    release = threading.Event()
    failures = []

    def hold():
        try:
            with sqlite3.connect(args.db, timeout=5) as connection:
                connection.execute("BEGIN IMMEDIATE")
                ready.set()
                release.wait(args.seconds)
                connection.rollback()
        except sqlite3.Error as error:
            failures.append(str(error))
            ready.set()

    worker = threading.Thread(target=hold)
    worker.start()
    try:
        if not ready.wait(6) or failures:
            parser.error("could not acquire database lock")
        print("Database write lock acquired for up to", args.seconds, "seconds", flush=True)
        result = subprocess.run(command, timeout=90, check=False)
        return result.returncode
    finally:
        release.set()
        worker.join()


if __name__ == "__main__":
    raise SystemExit(main())
