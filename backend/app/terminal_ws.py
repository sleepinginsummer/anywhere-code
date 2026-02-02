from __future__ import annotations

import asyncio
import subprocess
import threading

from fastapi import WebSocket, WebSocketDisconnect

from app.wsl_tmux import build_tmux_attach_command


def normalize_input(payload: str) -> str:
    """Normalize websocket input before sending to tmux."""
    return payload


def _start_output_pump(process: subprocess.Popen[str], websocket: WebSocket, loop: asyncio.AbstractEventLoop) -> threading.Event:
    """Start a background thread that forwards tmux output to the websocket."""
    stop_event = threading.Event()

    def _pump() -> None:
        for line in process.stdout or []:
            # Forward tmux output to the websocket.
            asyncio.run_coroutine_threadsafe(websocket.send_text(line), loop)
        stop_event.set()

    thread = threading.Thread(target=_pump, daemon=True)
    thread.start()
    return stop_event


def _spawn_tmux_attach(session_id: str) -> subprocess.Popen[str]:
    """Spawn tmux attach process in WSL."""
    command = build_tmux_attach_command(session_id)
    return subprocess.Popen(
        command,
        stdin=subprocess.PIPE,
        stdout=subprocess.PIPE,
        stderr=subprocess.STDOUT,
        text=True,
        bufsize=0,
    )


def _write_to_process(process: subprocess.Popen[str], payload: str) -> None:
    """Write normalized payload into tmux stdin."""
    if process.stdin is None:
        return
    process.stdin.write(payload)
    process.stdin.flush()


def _terminate_process(process: subprocess.Popen[str]) -> None:
    """Terminate the tmux attach process if still running."""
    if process.poll() is None:
        process.terminate()


async def handle_terminal_ws(websocket: WebSocket, session_id: str) -> None:
    """Bridge websocket traffic to a tmux session running in WSL."""
    await websocket.accept()
    process = _spawn_tmux_attach(session_id)
    loop = asyncio.get_running_loop()
    stop_event = _start_output_pump(process, websocket, loop)

    try:
        while not stop_event.is_set():
            message = await websocket.receive_text()
            _write_to_process(process, normalize_input(message))
    except WebSocketDisconnect:
        # Client disconnected; stop the bridge gracefully.
        pass
    finally:
        _terminate_process(process)
