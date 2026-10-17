#!/usr/bin/env python3
"""Small BDP v0 Read consumer; HTTP only, Python standard library only."""
import argparse
import json
import math
import os
import sys
from http.client import HTTPException
from urllib.error import HTTPError, URLError
from urllib.parse import unquote, urlsplit
from urllib.request import HTTPRedirectHandler, ProxyHandler, Request, build_opener


class ReadError(Exception):
    """An incomplete or invalid read; callers must not treat it as an empty result."""


def require(condition, message):
    if not condition:
        raise ReadError(message)


def http_url(url):
    require(isinstance(url, str) and url and not any(ord(c) <= 32 or ord(c) >= 127 for c in url),
            "Expected an ASCII-encoded absolute HTTP URL without whitespace")
    try:
        parts = urlsplit(url)
        port = parts.port  # Reject malformed ports before urllib sees them.
    except ValueError:
        raise ReadError("Invalid URL") from None
    require(parts.scheme in ("http", "https") and parts.hostname and parts.username is None
            and parts.password is None and not parts.fragment and "\\" not in url,
            "Invalid HTTP URL authority or fragment")
    require(port is None or 1 <= port <= 65535, "Invalid URL port")
    for segment in parts.path.split("/"):
        try:
            decoded = unquote(segment, errors="strict")
        except UnicodeError:
            raise ReadError("Invalid URL path encoding") from None
        require(decoded not in (".", "..") and not any(c in decoded for c in ("/", "\\", "%"))
                and not any(ord(c) <= 32 or ord(c) == 127 for c in decoded), "Unsafe URL path")
    return parts


class NoRedirects(HTTPRedirectHandler):
    def redirect_request(self, req, fp, code, msg, headers, newurl):
        raise ReadError("HTTP redirects are refused; configure the canonical Scope URL")


class Reader:
    def __init__(self, scope, token=None):
        self.scope = scope
        self.origin = http_url(scope)
        require(self.origin.path.endswith("/") and not self.origin.query,
                "Scope must end in / and have no query")
        require(not token or all(33 <= ord(c) <= 126 for c in token), "Invalid bearer token")
        self._token = token
        # Do not inherit ambient proxy routing or redirect credentials elsewhere.
        self._open = build_opener(ProxyHandler({}), NoRedirects()).open
        self.discovery = None

    def authorized(self, url, exact_path=None):
        parts = http_url(url)
        require((parts.scheme, parts.netloc) == (self.origin.scheme, self.origin.netloc)
                and parts.path.startswith(self.origin.path), "URL leaves the authorized Scope")
        if exact_path is not None:
            require(parts.path == exact_path, "Continuation changed the collection path")
        return url

    def get(self, url):
        self.authorized(url)
        headers = {"Accept": "application/json"}
        if self._token:
            headers["Authorization"] = "Bearer " + self._token
        try:
            with self._open(Request(url, headers=headers), timeout=15) as response:
                require(response.status == 200, "Expected HTTP 200")
                require(response.headers.get_content_type() == "application/json", "Expected JSON response")
                raw = response.read(8 * 1024 * 1024 + 1)
                require(len(raw) <= 8 * 1024 * 1024, "Response exceeds example's 8 MiB bound")
        except HTTPError as error:
            error.close()
            # Do not print response bodies, tokens, continuation URLs or headers.
            raise ReadError("BDP HTTP request failed (status %d)" % error.code) from None
        except (URLError, OSError, HTTPException):
            raise ReadError("BDP transport failed") from None
        try:
            return json.loads(raw.decode("utf-8"), parse_constant=lambda _: invalid_json(),
                              object_pairs_hook=unique_object, parse_float=finite_float)
        except (ValueError, UnicodeError, RecursionError):
            raise ReadError("Malformed JSON response") from None

    def discover(self):
        doc = self.get(self.scope + "bdp.json")
        require(isinstance(doc, dict) and doc.get("bdpVersion") == "0"
                and doc.get("profile") == "read" and doc.get("scope") == self.scope,
                "Expected BDP v0 Read discovery for the configured Scope")
        for collection in ("beads", "links", "types"):
            require(doc.get(collection) == self.scope + collection + "/",
                    "Discovery has an unexpected collection URL")
        self.discovery = doc
        return doc

    def bead(self, item):
        require(isinstance(item, dict) and isinstance(item.get("properties"), dict)
                and isinstance(item.get("revision"), str) and item["revision"], "Malformed Bead record")
        identity = item.get("id")
        self.authorized(identity)
        require(identity.startswith(self.scope + "beads/") and not http_url(identity).query
                and len(identity) > len(self.scope + "beads/"), "Invalid Bead identity")
        http_url(item.get("type"))  # Types may be published by another Scope; do not fetch them.
        return item

    def all_beads(self, limit=100):
        require(type(limit) is int and 1 <= limit <= 1000, "Limit must be from 1 to 1000")
        doc = self.discovery or self.discover()
        path = http_url(doc["beads"]).path
        url = doc["beads"] + "?limit=" + str(limit)
        pages, identities, records = set(), set(), []
        while url is not None:
            self.authorized(url, exact_path=path)
            require(url not in pages and len(pages) < 10000, "Repeated continuation or page bound exceeded")
            pages.add(url)
            page = self.get(url)
            require(isinstance(page, dict) and set(page) == {"items", "next"}
                    and isinstance(page["items"], list), "Malformed Bead collection")
            require(page["next"] is None or isinstance(page["next"], str), "Malformed continuation")
            for item in page["items"]:
                self.bead(item)
                require(item["id"] not in identities, "Repeated Bead identity across pages")
                identities.add(item["id"])
                records.append(item)
                require(len(records) <= 10000, "Example's 10000-Bead bound exceeded")
            # Keep the returned complete URL intact; do not reconstruct its query.
            url = page["next"]
        return records

    def read_bead(self, identity):
        if self.discovery is None:
            self.discover()
        self.authorized(identity)
        require(identity.startswith(self.scope + "beads/") and not http_url(identity).query,
                "Use a canonical current Bead identity")
        record = self.bead(self.get(identity))
        require(record["id"] == identity, "Returned Bead identity differs from request")
        return record


def finite_float(value):
    number = float(value)
    if not math.isfinite(number):
        raise ValueError("Non-finite JSON number")
    return number


def invalid_json():
    raise ValueError("Non-finite JSON number")


def unique_object(pairs):
    result = {}
    for key, value in pairs:
        if key in result:
            raise ValueError("Duplicate JSON key")
        result[key] = value
    return result


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--scope", required=True, help="Canonical Scope URL, including trailing /")
    parser.add_argument("--limit", type=int, default=100, help="Page size when enumerating (1–1000)")
    parser.add_argument("--id", help="Read this canonical Bead URL instead of enumerating")
    args = parser.parse_args()
    try:
        reader = Reader(args.scope, os.environ.get("BDP_TOKEN"))
        result = reader.read_bead(args.id) if args.id else reader.all_beads(args.limit)
        # Emit once, only after every requested read has succeeded.
        print(json.dumps(result, ensure_ascii=False, indent=2))
    except ReadError as error:
        print("BDP read failed: " + str(error), file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main())
