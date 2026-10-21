"""Fail-closed parser for one Claude stream-json attempt; never equate partials with review completion."""
import json

def parse_stream(raw):
    summary = {"records": 0, "eventTypes": {}, "assistantTextBlocks": 0,
               "partialTextChunks": 0, "toolCalls": 0, "toolErrors": 0,
               "resultCount": 0, "classification": "missing-result", "terminalSuccess": False}
    results = []
    last_type = None
    try:
        lines = raw.decode("utf-8").splitlines()
        for line in lines:
            if not line.strip():
                continue
            record = json.loads(line)
            if not isinstance(record, dict) or not isinstance(record.get("type"), str):
                raise ValueError("invalid stream record")
            summary["records"] += 1
            kind = record["type"]
            summary["eventTypes"][kind] = summary["eventTypes"].get(kind, 0) + 1
            last_type = kind
            if kind == "result":
                results.append(record)
            message = record.get("message")
            if isinstance(message, dict) and isinstance(message.get("content"), list):
                for block in message["content"]:
                    if not isinstance(block, dict):
                        continue
                    if kind == "assistant" and block.get("type") == "text":
                        summary["assistantTextBlocks"] += 1
                    if kind == "assistant" and block.get("type") == "tool_use":
                        summary["toolCalls"] += 1
                    if block.get("type") == "tool_result" and block.get("is_error") is True:
                        summary["toolErrors"] += 1
            event = record.get("event")
            if kind == "stream_event" and isinstance(event, dict):
                delta = event.get("delta")
                if isinstance(delta, dict) and delta.get("type") == "text_delta":
                    summary["partialTextChunks"] += 1
    except (UnicodeError, ValueError, TypeError):
        summary["classification"] = "malformed-stream"
        summary["resultCount"] = len(results)
        return summary, None
    summary["resultCount"] = len(results)
    if len(results) != 1:
        if results:
            summary["classification"] = "multiple-results"
        elif summary["assistantTextBlocks"] or summary["partialTextChunks"] or summary["toolCalls"]:
            summary["classification"] = "partial-output-no-result"
        return summary, None
    result = results[0]
    subtype = result.get("subtype")
    summary["resultSubtype"] = subtype
    summary["resultIsError"] = result.get("is_error")
    summary["numTurns"] = result.get("num_turns")
    if last_type != "result":
        summary["classification"] = "result-not-terminal"
    elif isinstance(subtype, str) and "max_turn" in subtype:
        summary["classification"] = "max-turns"
    elif result.get("is_error") is True or (isinstance(subtype, str) and subtype.startswith("error")):
        summary["classification"] = "error-result"
    elif subtype != "success" or result.get("is_error") is not False:
        summary["classification"] = "unknown-result"
    elif not isinstance(result.get("result"), str) or not result["result"].strip():
        summary["classification"] = "empty-final-result"
    else:
        summary["classification"] = "success-result"
        summary["terminalSuccess"] = True
        return summary, result["result"]
    return summary, None
