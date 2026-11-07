"""Offline protocol/credential controls. Real HTTP qualification is separate."""
import io
import json
import unittest
from email.message import Message
from http.client import BadStatusLine, IncompleteRead
from urllib.error import HTTPError, URLError

from read_beads import NoRedirects, Reader, ReadError

SCOPE = "https://example.invalid/team/"
DISCOVERY = dict(bdpVersion="0", profile="read", scope=SCOPE,
                 beads=SCOPE + "beads/", links=SCOPE + "links/", types=SCOPE + "types/")


def bead(name):
    return dict(id=SCOPE + "beads/" + name, type=SCOPE + "types/memory",
                revision="rev-" + name, properties={"body": name})


class Response(io.BytesIO):
    def __init__(self, value, content_type="application/json", status=200):
        super().__init__(json.dumps(value).encode())
        self.status = status
        self.headers = Message()
        self.headers["Content-Type"] = content_type


class ReaderTest(unittest.TestCase):
    def reader(self, responses):
        reader = Reader(SCOPE, "private-test-token")
        calls = []

        def open_response(request, timeout):
            calls.append(request.full_url)
            self.assertEqual(request.get_header("Authorization"), "Bearer private-test-token")
            self.assertEqual(timeout, 15)
            response = responses[request.full_url]
            if isinstance(response, Exception):
                raise response
            return response if isinstance(response, Response) else Response(response)

        reader._open = open_response
        return reader, calls

    def base(self, page):
        return {SCOPE + "bdp.json": DISCOVERY, SCOPE + "beads/?limit=1": page}

    def test_exhausts_returned_urls_including_empty_page(self):
        second = SCOPE + "beads/?cursor=opaque%2Bvalue&limit=1"
        third = SCOPE + "beads/?cursor=last"
        responses = self.base({"items": [bead("a")], "next": second})
        responses[second] = {"items": [], "next": third}
        responses[third] = {"items": [bead("b")], "next": None}
        reader, calls = self.reader(responses)
        self.assertEqual(reader.all_beads(1), [bead("a"), bead("b")])
        self.assertEqual(calls, [SCOPE + "bdp.json", SCOPE + "beads/?limit=1", second, third])

    def test_empty_inventory(self):
        reader, _ = self.reader(self.base({"items": [], "next": None}))
        self.assertEqual(reader.all_beads(1), [])

    def test_current_read_and_identity_check(self):
        responses = {SCOPE + "bdp.json": DISCOVERY, bead("a")["id"]: bead("a")}
        reader, _ = self.reader(responses)
        self.assertEqual(reader.read_bead(bead("a")["id"]), bead("a"))
        responses[bead("a")["id"]] = bead("b")
        with self.assertRaises(ReadError):
            reader.read_bead(bead("a")["id"])

    def test_continuations_cannot_forward_token_outside_collection(self):
        for next_url in ["https://evil.invalid/team/beads/?cursor=x",
                         "http://example.invalid/team/beads/?cursor=x",
                         "https://example.invalid:444/team/beads/?cursor=x",
                         "https://example.invalid/other/beads/?cursor=x",
                         SCOPE + "beads-else/?cursor=x", SCOPE + "links/?cursor=x",
                         SCOPE + "beads/../other?cursor=x", SCOPE + "beads/%2e%2e/other",
                         SCOPE + "beads/%252e%252e/other", SCOPE + "beads/%2fother",
                         SCOPE + "beads/?cursor=x#fragment", "/team/beads/?cursor=x",
                         "https://user@example.invalid/team/beads/?cursor=x", ""]:
            with self.subTest(next_url=next_url):
                reader, calls = self.reader(self.base({"items": [], "next": next_url}))
                with self.assertRaises(ReadError):
                    reader.all_beads(1)
                self.assertEqual(len(calls), 2)  # Refused before the credential-bearing request.

    def test_redirect_never_followed(self):
        with self.assertRaises(ReadError):
            NoRedirects().redirect_request(None, None, 302, "Found", {}, "https://evil.invalid/")

    def test_cycle_duplicate_and_malformed_pages_fail(self):
        for page in [{"items": [], "next": SCOPE + "beads/?limit=1"},
                     {"items": [bead("a"), bead("a")], "next": None},
                     {"items": [], "next": 4}, {"items": []},
                     {"items": {}, "next": None}, {"items": [{}], "next": None}]:
            with self.subTest(page=page):
                reader, _ = self.reader(self.base(page))
                with self.assertRaises(ReadError):
                    reader.all_beads(1)

    def test_discovery_wrong_scope_or_profile_fails(self):
        for changes in [{"scope": "https://evil.invalid/"}, {"profile": "read-update"},
                        {"bdpVersion": "1"}, {"beads": "https://evil.invalid/beads/"}]:
            reader, calls = self.reader({SCOPE + "bdp.json": {**DISCOVERY, **changes}})
            with self.assertRaises(ReadError):
                reader.discover()
            self.assertEqual(len(calls), 1)

    def test_http_transport_and_content_failures(self):
        for response in [HTTPError("secret-url", 401, "secret-message", {}, None),
                         URLError("private-test-token"), BadStatusLine("secret private-test-token"),
                         IncompleteRead(b"secret private-test-token"),
                         Response({}, "text/html"), Response({}, status=204)]:
            reader, _ = self.reader({SCOPE + "bdp.json": response})
            with self.assertRaises(ReadError) as error:
                reader.discover()
            self.assertNotIn("private-test-token", str(error.exception))
            self.assertNotIn("secret", str(error.exception))

    def test_later_page_failure_is_not_partial_success(self):
        next_url = SCOPE + "beads/?cursor=expired"
        responses = self.base({"items": [bead("a")], "next": next_url})
        responses[next_url] = HTTPError(next_url, 410, "expired", {}, None)
        reader, _ = self.reader(responses)
        with self.assertRaises(ReadError):
            reader.all_beads(1)

    def test_invalid_json_is_refused(self):
        for raw in [b'{"scope":1,"scope":2}', b'NaN', b'1e999', b'not-json', b'\xff']:
            response = Response(None)
            response.seek(0)
            response.truncate()
            response.write(raw)
            response.seek(0)
            reader, _ = self.reader({SCOPE + "bdp.json": response})
            with self.assertRaises(ReadError):
                reader.discover()


if __name__ == "__main__":
    unittest.main()
