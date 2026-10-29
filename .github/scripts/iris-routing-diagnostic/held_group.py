"""SOURCE-ONLY lifecycle seam. Production and mocked controls share decisions."""
import os
import signal
import subprocess
import time

class LostOwnership(RuntimeError):
    pass

class OwnedGroup:
    def __init__(self, child):
        self.child = child
        self.pid = child.pid
        self.reaped = False

    def observe_exit(self):
        if self.reaped:
            raise LostOwnership('leader already reaped')
        # WNOWAIT leaves the exact direct child identity reserved. NEVER poll or
        # wait until the final signaling phase is over. External reaping is loss,
        # not permission to use the numeric PID as an ownership substitute.
        try:
            info = os.waitid(os.P_PID, self.pid, os.WEXITED | os.WNOHANG | os.WNOWAIT)
        except ChildProcessError as error:
            raise LostOwnership('exact direct child no longer waitable') from error
        return info is not None

    def members(self):
        self.observe_exit()  # Also refuses external ownership loss while live.
        p = subprocess.run(['/bin/ps', '-axo', 'pid=,pgid='],
                           env={'PATH': '/usr/bin:/bin', 'LANG': 'C'},
                           stdout=subprocess.PIPE, stderr=subprocess.PIPE, timeout=3)
        if p.returncode or p.stderr:
            raise RuntimeError('owned group observer failed')
        result = []
        for row in p.stdout.decode().splitlines():
            fields = row.split()
            if len(fields) != 2 or not all(v.isdigit() for v in fields):
                raise RuntimeError('malformed group observer output')
            pid, pgid = map(int, fields)
            if pgid == self.pid:
                result.append(pid)
        return result

    def signal(self, number):
        self.observe_exit()  # Retained waitable identity anchors this PGID.
        try:
            os.killpg(self.pid, number)
        except ProcessLookupError:
            pass

    def reap(self):
        # One terminal reap. Mark ownership lost BEFORE wait so no exceptional
        # return can enable a subsequent killpg against a recycled numeric ID.
        if self.reaped:
            raise LostOwnership('duplicate terminal reap')
        self.reaped = True
        return self.child.wait(timeout=2)

def observe_until_exit(group, now, sleep, canceled, bound, progress, start=None):
    if start is None:
        start = now()
    deadline = start + bound
    next_progress = start + 30
    while True:
        exited = group.observe_exit()
        observed_at = now()
        # Deadline is applied to terminal OBSERVATION, including a zero exit.
        # Reaching the boundary is failure even if the child exited earlier.
        if observed_at >= deadline:
            return {'timedOut': True, 'terminalObserved': exited,
                    'executionSeconds': observed_at - start}
        if canceled():
            return {'timedOut': False, 'terminalObserved': exited,
                    'executionSeconds': observed_at - start}
        if exited:
            return {'timedOut': False, 'terminalObserved': True,
                    'executionSeconds': observed_at - start}
        if observed_at >= next_progress:
            progress(observed_at - start)
            next_progress += 30
        sleep(min(0.05, deadline - observed_at))

def cleanup_owned(group, now=time.monotonic, sleep=time.sleep):
    start = now()
    forced = False
    # Exited leader is kept waitable while any descendants may need signaling.
    # Live leader is included as a member; a zombie is excluded only after its
    # WNOWAIT terminal observation. Observer/identity failure retains custody.
    def live_members():
        exited = group.observe_exit()
        return [p for p in group.members() if not (exited and p == group.pid)]
    try:
        members = live_members()
        for number, bound in ((signal.SIGTERM, 5), (signal.SIGKILL, 2)):
            if not members:
                break
            forced = True
            group.signal(number)
            until = now() + bound
            while True:
                members = live_members()
                observed_at = now()
                if not members or observed_at >= until:
                    break
                sleep(min(0.05, until - observed_at))
        terminal = group.observe_exit()
        if members or not terminal:
            return {'extraCleanupRequired': forced, 'parentReaped': False,
                    'groupGone': False, 'cleanupSeconds': now() - start,
                    'cleanupError': 'owned processes remain; identity retained, no removal'}
        status = group.reap()
        # NO group signal or numeric ownership test is allowed after this reap.
        return {'extraCleanupRequired': forced, 'parentReaped': True,
                'groupGone': True, 'childExit': status,
                'cleanupSeconds': now() - start}
    except (LostOwnership, OSError, RuntimeError, subprocess.SubprocessError) as error:
        return {'extraCleanupRequired': forced, 'parentReaped': False,
                'groupGone': False, 'cleanupSeconds': now() - start,
                'cleanupError': str(error)}

def supervise_owned(group, now, sleep, canceled, bound, progress, start, cleanup=cleanup_owned, begin=None):
    """Preserve primary observation and cleanup outcomes independently.

    No error is interpreted as success or allowed to mask a prior failure. The
    dedicated controller serializes this result before its failed disposition.
    """
    observation = {'timedOut': True, 'terminalObserved': False,
                   'executionSeconds': None}
    operation_error = None
    try:
        if begin is not None:
            begin()
        observation = observe_until_exit(group, now, sleep, canceled, bound, progress, start=start)
    except Exception as error:
        operation_error = {'type': type(error).__name__, 'message': str(error)}
        if hasattr(error, 'failures'):
            operation_error['failures'] = error.failures
    try:
        cleanup_result = cleanup(group, now=now, sleep=sleep)
    except Exception as error:
        cleanup_result = {'extraCleanupRequired': True, 'parentReaped': False,
                          'groupGone': False,
                          'cleanupError': {'type': type(error).__name__, 'message': str(error)}}
    return {'observation': observation, 'operationError': operation_error,
            'cleanup': cleanup_result}
