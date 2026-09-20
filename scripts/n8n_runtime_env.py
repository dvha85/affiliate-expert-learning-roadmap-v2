"""Build a hermetic environment for disposable n8n regression runs.

The regression runners execute n8n in a temporary SQLite workspace.  A caller's
shell may already contain PostgreSQL, queue-mode, config-file, or task-runner
settings; inheriting those values would make the test mutate or observe an
external runtime.  Keep this policy in one small module so both engine runners
apply the same boundary.
"""

import os
from pathlib import Path
from typing import Mapping, Optional


_EXTERNAL_RUNTIME_PREFIXES = (
    "N8N_",
    "DB_",
    "QUEUE_",
    "TASK_RUNNERS_",
    "EXECUTIONS_",
)


def isolated_n8n_environment(
    runtime: Path,
    *,
    broker_port: int,
    base: Optional[Mapping[str, str]] = None,
    node_path: Optional[Path] = None,
) -> dict[str, str]:
    """Return an n8n environment whose state and execution mode are local.

    ``base`` is injectable for tests; production callers use ``os.environ``.
    The prefix removal deliberately drops inherited ``N8N_*``
    settings plus direct and ``*_FILE`` database/queue forms. n8n resolves
    those variables before selecting its backend, so retaining an arbitrary
    setting could redirect logs, task runners, or state outside the fixture.
    ``N8N_USER_FOLDER`` is then set to the parent used by n8n for ``.n8n`` state
    and SQLite database files.
    """

    environment = dict(os.environ if base is None else base)
    for key in list(environment):
        if key.startswith(_EXTERNAL_RUNTIME_PREFIXES):
            environment.pop(key, None)

    n8n_home = runtime / "n8n"
    environment.update(
        {
            "N8N_USER_FOLDER": str(n8n_home),
            "N8N_ENCRYPTION_KEY": "n8n-ci-isolated-fixture-key-not-a-secret",
            "N8N_DIAGNOSTICS_ENABLED": "false",
            "N8N_PERSONALIZATION_ENABLED": "false",
            "N8N_ENFORCE_SETTINGS_FILE_PERMISSIONS": "false",
            "DB_TYPE": "sqlite",
            "DB_SQLITE_POOL_SIZE": "0",
            "EXECUTIONS_MODE": "regular",
            "N8N_RUNNERS_BROKER_PORT": str(broker_port),
            "GOWORK": "off",
            "GOCACHE": str(runtime / "go-cache"),
        }
    )
    if node_path is not None:
        environment["PATH"] = str(node_path.resolve().parent) + os.pathsep + environment.get("PATH", "")
    return environment
