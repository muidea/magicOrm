#!/usr/bin/env python3
"""Check home-rooted documentation links in a workspace or standalone checkout."""
from dataclasses import dataclass, field
from pathlib import Path, PurePosixPath
import re
import sys
from urllib.parse import unquote


REPOSITORY = ("~", "codespace", "magicOrm")


@dataclass
class CheckResult:
    checked: int = 0
    unavailable: dict[str, int] = field(default_factory=dict)
    failures: list[str] = field(default_factory=list)


def check_documents(root: Path, home: Path) -> CheckResult:
    result = CheckResult()
    for document in sorted(root.rglob("*.md")):
        relative = document.relative_to(root)
        if {"vendor", ".git"}.intersection(relative.parts):
            continue
        for target in re.findall(r"\[[^\]]+\]\(([^)]+)\)", document.read_text()):
            target = target.strip("<>")
            if target.startswith(("#", "https:", "http:", "mailto:", "data:")):
                continue
            path = re.sub(r":\d+$", "", unquote(target.split("#", 1)[0]))
            if not path.startswith("~/"):
                result.failures.append(
                    f"{relative}: file reference must start with ~/ : {target}"
                )
                continue
            parts = PurePosixPath(path).parts
            if ".." in parts:
                result.failures.append(f"{relative}: parent traversal is not allowed: {target}")
                continue
            if parts[:3] == REPOSITORY:
                # Resolve this repository in the checkout, never in a stale home copy.
                resolved = root.joinpath(*parts[3:])
            elif len(parts) >= 3 and parts[1] in {"codespace", "dataspace"}:
                repository = home.joinpath(*parts[1:3])
                sibling = root.parent / parts[2]
                if parts[1] == "codespace" and sibling.is_dir():
                    repository = sibling
                if not repository.is_dir():
                    key = "/".join(parts[:3])
                    result.unavailable[key] = result.unavailable.get(key, 0) + 1
                    continue
                resolved = repository.joinpath(*parts[3:])
            else:
                resolved = home.joinpath(*parts[1:])
            result.checked += 1
            if not resolved.exists():
                result.failures.append(f"{relative}: missing file: {target}")
    return result


def main() -> int:
    root = Path(__file__).resolve().parents[1]
    result = check_documents(root, Path.home())
    for repository, count in sorted(result.unavailable.items()):
        print(f"Not checked: {count} external links; repository unavailable: {repository}")
    if result.failures:
        print("\n".join(result.failures))
        return 1
    print(f"Validated {result.checked} home-rooted file links")
    return 0


if __name__ == "__main__":
    sys.exit(main())
