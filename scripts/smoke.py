#!/usr/bin/env python3
"""Exercise the real HTTP process and restart/export/restore on disposable data only."""
import argparse
import json
import os
from pathlib import Path
import socket
import subprocess
import tempfile
import time
from urllib.error import HTTPError, URLError
from urllib.request import Request, urlopen

ROOT = Path(__file__).resolve().parents[1]
parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument('--compose', action='store_true', help='build and test a disposable Compose project')
parser.add_argument('--binary', default='./portal', help='native binary (build with go build -o portal .)')
args = parser.parse_args()
with socket.socket() as sock:
    sock.bind(('127.0.0.1', 0))
    port = sock.getsockname()[1]
base = f'http://127.0.0.1:{port}'
env = os.environ.copy()
env['PORT'] = str(port)
compose = ['docker', 'compose', '--project-name', f'portal-smoke-{os.getpid()}']
process = None


def command(argv, **kwargs):
    return subprocess.run(argv, cwd=ROOT, env=env, check=True, **kwargs)


def request(path, data=None, expected=200, method='GET'):
    body = None if data is None else json.dumps(data).encode()
    req = Request(base + path, data=body, method=method,
                  headers={'Content-Type': 'application/json'})
    try:
        response = urlopen(req, timeout=5)
    except HTTPError as error:
        response = error
    with response:
        content = response.read()
        assert response.status == expected, (path, response.status, content)
        return json.loads(content) if 'application/json' in response.headers.get('Content-Type', '') else content


def must_fail(argv):
    result = subprocess.run(argv, cwd=ROOT, env=env, capture_output=True, text=True)
    assert result.returncode != 0, 'invalid backup unexpectedly accepted'


def wait_ready():
    deadline = time.monotonic() + 45
    while time.monotonic() < deadline:
        try:
            request('/api/state')
            return
        except (URLError, TimeoutError, ConnectionError):
            if process is not None and process.poll() is not None:
                raise RuntimeError('portal exited before startup')
            time.sleep(0.2)
    raise RuntimeError('portal did not become ready')


def stop_native():
    if process is not None and process.poll() is None:
        process.terminate()
        try:
            process.wait(timeout=10)
        except subprocess.TimeoutExpired:
            process.kill()
            process.wait()
            raise


with tempfile.TemporaryDirectory(prefix='helicopter-smoke-') as directory:
    temp = Path(directory)
    log = open(temp / 'portal.log', 'w+')
    binary = str((ROOT / args.binary).resolve())
    env.update(ADDR=f'127.0.0.1:{port}', DATA_FILE=str(temp / 'progress.json'))
    try:
        if args.compose:
            command(compose + ['up', '--build', '-d'])
        else:
            process = subprocess.Popen([binary], cwd=ROOT, env=env, stdout=log, stderr=log)
        wait_ready()
        assert b'id="app"' in request('/')
        assert b'function' in request('/app.js')
        curriculum = request('/api/curriculum')
        assert len(curriculum['stages']) == 8
        assert sum(len(s['modules']) for s in curriculum['stages']) == 42
        assert request('/api/state') == {'progress': {}, 'sessions': []}
        passing = dict(status='passed', score=80, evidence='portfolio/example.md', practicalConfirmed=True)
        request('/api/progress/S1M1', dict(status='passed', score=100), 400, 'PUT')
        request('/api/progress/S1M2', passing, 409, 'PUT')
        assessment = request('/api/progress/S1M1', passing, method='PUT')
        assert assessment['completedAt']
        request('/api/sessions', dict(moduleId='S1M1', minutes=45, date='2026-10-30', note='smoke study'), 201, 'POST')
        before = request('/api/state')
        if args.compose:
            command(compose + ['restart', 'portal'])
        else:
            stop_native()
            process = subprocess.Popen([binary], cwd=ROOT, env=env, stdout=log, stderr=log)
        wait_ready()
        assert request('/api/state') == before, 'restart lost progress'
        exported = request('/api/export')
        assert exported == before
        backup = temp / 'backup.json'
        backup.write_text(json.dumps(exported))
        bad_backup = temp / 'invalid.json'
        bad_backup.write_text('{')
        if args.compose:
            # Bind only synthetic smoke records; the runtime user needs read permission.
            temp.chmod(0o755)
            backup.chmod(0o644)
            bad_backup.chmod(0o644)
            mount = ['--volume', f'{temp}:/backups:ro']
            command(compose + ['stop', 'portal'])
            must_fail(compose + ['run', '--rm', '-T', '--no-deps'] + mount + ['portal', '-restore', '/backups/invalid.json'])
            must_fail(compose + ['run', '--rm', '-T', '--no-deps'] + mount + ['portal', '-restore', '/backups/missing.json'])
            command(compose + ['run', '--rm', '-T', '--no-deps'] + mount + ['portal', '-check-data', '/backups/backup.json'])
            command(compose + ['run', '--rm', '-T', '--no-deps'] + mount + ['portal', '-restore', '/backups/backup.json'])
            command(compose + ['up', '-d'])
        else:
            stop_native()
            must_fail([binary, '-restore', str(bad_backup)])
            must_fail([binary, '-restore', str(temp / 'missing.json')])
            assert json.loads((temp / 'progress.json').read_text()) == before
            command([binary, '-check-data', str(backup)])
            env['DATA_FILE'] = str(temp / 'restored.json')
            command([binary, '-restore', str(backup)])
            process = subprocess.Popen([binary], cwd=ROOT, env=env, stdout=log, stderr=log)
        wait_ready()
        assert request('/api/state') == before, 'restore changed records'
        print('PASS: dashboard/assets, 8 stages/42 modules, grading, prerequisites, real restart, export and restore')
    except Exception:
        if args.compose:
            subprocess.run(compose + ['logs', '--no-color'], cwd=ROOT, env=env)
        else:
            log.flush()
            log.seek(0)
            print(log.read())
        raise
    finally:
        stop_native()
        if args.compose:
            # The project and volume were created by this script; never targets normal portal data.
            subprocess.run(compose + ['down', '--volumes', '--remove-orphans'], cwd=ROOT, env=env, check=True)
        log.close()
