"""Descriptor-anchored private build output; successful receipt is published last."""
import hashlib
import json
import os
from pathlib import Path
import stat


def json_bytes(value):
    return (json.dumps(value, indent=2, sort_keys=True) + "\n").encode()


def key(metadata):
    return metadata.st_dev, metadata.st_ino


class BuildOutput:
    def __init__(self, root, path, max_file_bytes):
        path = path.absolute()
        root = root.resolve()
        admitted = path.resolve()
        if path.exists() or path.is_symlink():
            raise ValueError("build output already exists")
        if admitted.is_relative_to(root):
            raise ValueError("build output must be outside the source checkout")
        path.parent.mkdir(parents=True, exist_ok=True)
        self.path = path.parent.resolve() / path.name
        if self.path != admitted or self.path.is_relative_to(root):
            raise ValueError("build output parent changed during initialization")
        self.parent = os.open(self.path.parent, os.O_RDONLY | os.O_DIRECTORY | os.O_NOFOLLOW)
        self.fd = None
        self.created = {}
        self.max_file_bytes = max_file_bytes
        try:
            parent = self.path.parent.lstat()
            if self.path.resolve() != self.path or not stat.S_ISDIR(parent.st_mode) or key(parent) != key(os.fstat(self.parent)):
                raise ValueError("build output parent changed during initialization")
            os.mkdir(path.name, mode=0o700, dir_fd=self.parent)
            self.fd = os.open(path.name, os.O_RDONLY | os.O_DIRECTORY | os.O_NOFOLLOW, dir_fd=self.parent)
            self.require()
        except BaseException:
            self.close()
            raise

    def require(self):
        parent = self.path.parent.lstat()
        entry = os.stat(self.path.name, dir_fd=self.parent, follow_symlinks=False)
        if self.path.resolve() != self.path or not stat.S_ISDIR(parent.st_mode) or not stat.S_ISDIR(entry.st_mode) or key(parent) != key(os.fstat(self.parent)) or key(entry) != key(os.fstat(self.fd)):
            raise ValueError("build output directory or parent was replaced")

    def create(self, name, mode=0o644):
        self.require()
        if Path(name).name != name or name in {"", ".", ".."}:
            raise ValueError("invalid build output filename")
        fd = os.open(name, os.O_WRONLY | os.O_CREAT | os.O_EXCL | os.O_NOFOLLOW, mode, dir_fd=self.fd)
        self.created[name] = key(os.fstat(fd))
        try:
            os.fchmod(fd, mode)
            return os.fdopen(fd, "wb")
        except BaseException:
            os.close(fd)
            self.remove_created(name)
            raise

    def remove_created(self, name):
        try:
            current = os.stat(name, dir_fd=self.fd, follow_symlinks=False)
            if stat.S_ISREG(current.st_mode) and key(current) == self.created.get(name):
                os.unlink(name, dir_fd=self.fd)
        except FileNotFoundError:
            pass

    def write_json(self, name, value):
        data = json_bytes(value)
        if len(data) > self.max_file_bytes:
            raise ValueError("build JSON exceeds file bound")
        created = False
        try:
            with self.create(name) as stream:
                created = True
                stream.write(data); stream.flush(); os.fsync(stream.fileno())
        except BaseException:
            if created: self.remove_created(name)
            raise

    def copy(self, source, name, mode=0o644):
        source_fd = os.open(source, os.O_RDONLY | os.O_NOFOLLOW | os.O_NONBLOCK)
        created = False
        try:
            metadata = os.fstat(source_fd)
            if not stat.S_ISREG(metadata.st_mode):
                raise ValueError("build artifact is not a regular file")
            if metadata.st_size > self.max_file_bytes:
                raise ValueError("build artifact exceeds file bound")
            count = 0
            with self.create(name, mode) as destination:
                created = True
                while chunk := os.read(source_fd, 64 * 1024):
                    count += len(chunk)
                    if count > self.max_file_bytes:
                        raise ValueError("build artifact exceeds file bound")
                    destination.write(chunk)
                destination.flush(); os.fsync(destination.fileno())
        except BaseException:
            if created: self.remove_created(name)
            raise
        finally:
            os.close(source_fd)

    def identity(self, name):
        self.require()
        if Path(name).name != name or name in {"", ".", ".."}:
            raise ValueError("invalid build output filename")
        fd = os.open(name, os.O_RDONLY | os.O_NOFOLLOW | os.O_NONBLOCK, dir_fd=self.fd)
        try:
            before = os.fstat(fd)
            if not stat.S_ISREG(before.st_mode):
                raise ValueError("build output is not a regular file")
            if key(before) != self.created.get(name):
                raise ValueError("owned build output file was replaced")
            digest, count = hashlib.sha256(), 0
            while chunk := os.read(fd, 64 * 1024):
                count += len(chunk)
                if count > self.max_file_bytes:
                    raise ValueError("build output exceeds file bound")
                digest.update(chunk)
            after = os.fstat(fd)
            signature = lambda value: (key(value), value.st_size, value.st_mtime_ns, value.st_ctime_ns)
            if signature(before) != signature(after):
                raise ValueError("build output changed while hashing")
            entry = os.stat(name, dir_fd=self.fd, follow_symlinks=False)
            if not stat.S_ISREG(entry.st_mode) or key(entry) != self.created.get(name):
                raise ValueError("owned build output file was replaced while hashing")
            self.require()
            return {"file": name, "bytes": count, "sha256": digest.hexdigest()}
        finally:
            os.close(fd)

    def close(self):
        if self.fd is not None:
            os.close(self.fd)
            self.fd = None
        if self.parent is not None:
            os.close(self.parent)
            self.parent = None
